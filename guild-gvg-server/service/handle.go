package service

import (
	"context"
	"math"
	"sort"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventloop"
	"coin-server/common/gopool"
	"coin-server/common/handler"
	"coin-server/common/im"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/guild_filter_service"
	"coin-server/common/proto/gvgguild"
	"coin-server/common/proto/models"
	"coin-server/common/proto/rank_service"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/utils/percent"
	"coin-server/common/values"
	formationDao "coin-server/game-server/service/formation/dao"
	guildDao "coin-server/game-server/service/guild/dao"
	userDao "coin-server/game-server/service/user/db"
	"coin-server/rule"

	jsoniter "github.com/json-iterator/go"

	"go.uber.org/zap"
)

var errStatusMap = map[GVGStatus]func() *errmsg.ErrMsg{
	StatusSignup: func() *errmsg.ErrMsg {
		return errmsg.NewErrorGuildGVGInSignup()
	},
	StatusMatch: func() *errmsg.ErrMsg {
		return errmsg.NewErrorGuildGVGInMatch()
	},
	StatusFighting: func() *errmsg.ErrMsg {
		return errmsg.NewErrorGuildGVGInFighting()
	},
	StatusSettlement: func() *errmsg.ErrMsg {
		return errmsg.NewErrorGuildGVGInFighting()
	},
}

func (this_ *Service) StatusCheck(status GVGStatus) func(handler.HandleFunc) handler.HandleFunc {
	return func(next handler.HandleFunc) handler.HandleFunc {
		return func(ctx *ctx.Context) *errmsg.ErrMsg {
			gvgStatus := this_.GetGVGStats()
			if gvgStatus != status {
				f, ok := errStatusMap[gvgStatus]
				if ok {
					return f()
				}
				return errmsg.NewErrGuildNotExist()
			}
			return next(ctx)
		}
	}
}

func (this_ *Service) Serve() {
	this_.init()
	this_.Router()
	this_.Start(func(event interface{}) {
		this_.log.Warn("unknown event", zap.Any("event", event))
	})
}

func (this_ *Service) Router() {
	needInSignupHandler := this_.handler.Group(this_.StatusCheck(StatusSignup))
	needInSignupHandler.RegisterFunc("报名", this_.HandleSignup)

	this_.handler.RegisterFunc("进入工会主界面查询活动情况弹窗", this_.QueryStatus)
	this_.handler.RegisterFunc("进入活动主界面查询信息", this_.QueryActiveInfo)

	needFightHandler := this_.handler.Group(this_.StatusCheck(StatusFighting))
	needFightHandler.RegisterFunc("获取工会信息", this_.QueryActiveGuild)
	needFightHandler.RegisterFunc("标记工会", this_.MaskGuild)
	needFightHandler.RegisterFunc("获取建筑详情", this_.QueryBuildInfo)
	needFightHandler.RegisterFunc("请求攻击玩家战斗信息", this_.FightRole)
	needFightHandler.RegisterFunc("请求攻击玩家战斗结果", this_.FightRoleFinish)
	needFightHandler.RegisterFunc("请求攻击建筑", this_.FightBuild)
	needFightHandler.RegisterFunc("查询工会排行信息", this_.QueryGuildRank)
	needFightHandler.RegisterFunc("查询个人排行信息", this_.QueryPersonalRank)
	needFightHandler.RegisterFunc("查询战斗记录", this_.QueryFightingInfo)

	eventloop.RegisterFuncChanEventLoop(this_.signupQueue, this_.CheckSignup)
	eventloop.RegisterFuncChanEventLoop(this_.signupQueue, this_.CheckAndSetSignup)
	eventloop.RegisterFuncChanEventLoop(this_.signupQueue, this_.DeleteSignupInfo)
}

// QueryStatus 查询自己是否可以报名，是否已经报名。工会主界面弹窗用
func (this_ *Service) QueryStatus(c *ctx.Context, _ *gvgguild.GuildGVG_QueryStatusRequest) (*gvgguild.GuildGVG_QueryStatusResponse, *errmsg.ErrMsg) {
	ss := gvgguild.GuildGVG_SignStatus(-1)
	status := this_.GetGVGStats()
	switch status {
	case StatusMatch:
		ss = gvgguild.GuildGVG_Matching
	case StatusFighting:
		ss = gvgguild.GuildGVG_Fighting
	case StatusSettlement:
		ss = gvgguild.GuildGVG_Settling
	default:

	}
	if ss != gvgguild.GuildGVG_SignStatus(-1) {
		timestamp := int64(0)
		if ss == gvgguild.GuildGVG_Settling {
			timestamp = this_.activeStartTime
		}
		return &gvgguild.GuildGVG_QueryStatusResponse{
			Status:               ss, // 活动已经开始
			ActiveStartTimestamp: timestamp,
		}, nil
	}
	g, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if g.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	resp, err := eventloop.RequestChanEventLoop[*CheckSignInfo, *CheckSignInfo](this_.signupQueue, &CheckSignInfo{
		Id: g.GuildId,
	})
	if err != nil {
		return nil, err
	}
	if resp.IsSignup {
		return &gvgguild.GuildGVG_QueryStatusResponse{
			Status:               gvgguild.GuildGVG_SignupSuccess, // 已经报名
			SignupName:           resp.Nickname,
			ActiveStartTimestamp: this_.activeStartTime,
		}, nil
	}

	out := &guild_filter_service.Guild_GuildGVGInfoResponse{}
	err = this_.nc.RequestWithOut(c, 0, &guild_filter_service.Guild_GuildGVGInfoRequest{Id: g.GuildId}, out)
	if err != nil {
		return nil, err
	}
	r := rule.MustGetReader(c)
	needNum, ok := r.KeyValue.GetInt64("GuildContendNum")
	utils.MustTrue(ok)
	guildCnf, ok := r.Guild.GetGuildById(out.Level)
	utils.MustTrue(ok)

	fo := false
	for _, v := range guildCnf.FunctionOpen {
		if v == GuildFunctionGVG {
			fo = true
			break
		}
	}

	if out.Active == 0 || needNum > out.Count || !fo {
		return &gvgguild.GuildGVG_QueryStatusResponse{
			Status:               gvgguild.GuildGVG_CannotSignup, // 不满足条件
			ActiveStartTimestamp: this_.activeStartTime,
		}, nil
	}

	member, ok, err := guildDao.NewGuildMember(g.GuildId).GetOne(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	gp, ok := r.GuildPosition.GetGuildPositionById(member.Position)
	utils.MustTrue(ok)
	if !gp.GvgSign {
		return &gvgguild.GuildGVG_QueryStatusResponse{
			Status:               gvgguild.GuildGVG_NoAuthoritySignup, // 条件满足。但是没得权限
			ActiveStartTimestamp: this_.activeStartTime,
		}, nil
	}
	return &gvgguild.GuildGVG_QueryStatusResponse{
		Status:               gvgguild.GuildGVG_CanSignup, // 条件满足，可以去报名
		ActiveStartTimestamp: this_.activeStartTime,
	}, nil
}

func (this_ *Service) QueryActiveInfo(c *ctx.Context, _ *gvgguild.GuildGVG_QueryActiveInfoRequest) (*gvgguild.GuildGVG_QueryActiveInfoResponse, *errmsg.ErrMsg) {
	// 先查询本公会是否报名，如果没报名提示没有报名，不能进入活动
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	resp, err := eventloop.RequestChanEventLoop[*CheckSignInfo, *CheckSignInfo](this_.signupQueue, &CheckSignInfo{
		Id: gu.GuildId,
	})
	if err != nil {
		return nil, err
	}
	if !resp.IsSignup {
		return nil, errmsg.NewErrorGuildGVGNotSignup()
	}

	status := this_.GetGVGStats()
	out := &gvgguild.GuildGVG_QueryActiveInfoResponse{
		Status: gvgguild.GuildGVG_SignStatus(status),
		GroupData: &gvgguild.GuildGVG_GroupInfo{
			GroupId: 0,
			Infos:   nil,
		},
	}
	if status != StatusFighting { // 活动未在进行中，包含报名，匹配，结算阶段
		// 只返回本工会信息
		g, err := guildDao.NewGuild(gu.GuildId).Get(c)
		if err != nil {
			return nil, err
		}
		gg := &gvgguild.GuildGVG_GuildInfo{
			GuildId: g.Id,
			Flag:    g.Flag,
			Name:    g.Name,
			Level:   g.Level,
		}
		out.GroupData.Infos = append(out.GroupData.Infos, gg)
	} else { // 只包含战斗阶段
		resp, err := eventloop.CallChanEventLoop(this_.signupQueue, gu.GuildId, func(guildId string) ([]*GuildInfoTempInfo, *errmsg.ErrMsg) {
			groupId, ok := this_.guildGroupInfo[guildId]
			utils.MustTrue(ok) // 必定有报名，前面已经在内存验证过了，如果!ok 就是bug
			group, ok := this_.groupMap[groupId]
			utils.MustTrue(ok) // 必定有这个分组，没有的话就是bug
			resp := make([]*GuildInfoTempInfo, 0, 10)

			this_.calcGuildScoreGroup(group)
			for i := range group.Infos {
				v := &group.Infos[i]
				ggi := &gvgguild.GuildGVG_GuildInfo{
					GuildId:    v.Id,
					DefineData: nil,
				}
				if v.Id == gu.GuildId {
					out.GroupData.MaskGuildId = v.FlagGuildId
				}
				if v.LastFightInfo != 0 {
					fi, ok := group.fiMap[v.LastFightInfo]
					if ok {
						ggi.DefineData = &gvgguild.GuildGVG_DefendInfo{
							Timestamp:   fi.CreateTime.UTC().Unix(),
							Nickname:    fi.Attacker,
							AvatarId:    0,
							AvatarFrame: 0,
							Result:      fi.IsWin,
						}
					}
				}

				gt := &GuildInfoTempInfo{Score: v.Score, LastScoreChange: v.LastScoreChange, GuildGVG_GuildInfo: ggi}
				resp = append(resp, gt)
				isCheckBurn := false
				for ii := range v.Data {
					vv := &v.Data[ii]
					if vv.Id == 1 {
						if vv.Blood == 0 {
							ggi.IsBurn = true
						}
						isCheckBurn = true
						if ggi.IsSmoke {
							break
						}
					}
					if vv.Blood < vv.MaxBlood {
						ggi.IsSmoke = true
						if isCheckBurn {
							break
						}
					}
				}
			}
			return resp, nil
		})
		if err != nil {
			return nil, err
		}

		guilds := make([]string, len(resp))
		definedInfo := make([]string, 0, len(resp))
		for _, v := range resp {
			guilds = append(guilds, v.GuildId)
			if v.DefineData != nil {
				definedInfo = append(definedInfo, v.DefineData.Nickname)
			}
		}

		roleMap, err := userDao.GetRoles(c, definedInfo)
		if err != nil {
			return nil, err
		}
		daoGuilds, err := guildDao.NewGuild("").GetMulti(c, guilds)
		if err != nil {
			return nil, err
		}
		for _, v := range resp {
			dg, ok := daoGuilds[v.GuildId]
			if ok {
				v.Level = dg.Level
				v.Flag = dg.Flag
				v.Name = dg.Name
			}
			if v.DefineData != nil {
				role, ok := roleMap[v.DefineData.Nickname]
				if ok {
					v.DefineData.Nickname = role.Nickname
					v.DefineData.AvatarId = role.AvatarId
					v.DefineData.AvatarFrame = role.AvatarFrame
				}
			}
		}
		sort.Slice(resp, func(i, j int) bool {
			ri := resp[i]
			rj := resp[j]
			if ri.Score > rj.Score {
				return true
			} else if ri.Score < rj.Score {
				return false
			}
			if ri.LastScoreChange < rj.LastScoreChange {
				return true
			} else if ri.LastScoreChange > rj.LastScoreChange {
				return false
			}
			return ri.GuildId < rj.GuildId
		})
		for i, v := range resp {
			v.Rank = int64(i + 1)
			out.GroupData.Infos = append(out.GroupData.Infos, v.GuildGVG_GuildInfo)
		}

	}
	switch GVGStatus(out.Status) {
	case StatusSignup:
		out.RemainTimestamp = this_.activeStartTime
		out.Status = gvgguild.GuildGVG_SignupSuccess
	case StatusSettlement:
		out.Status = gvgguild.GuildGVG_Settling
		out.RemainTimestamp = this_.activeStartTime
	case StatusFighting:
		out.Status = gvgguild.GuildGVG_Fighting
		out.RemainTimestamp = this_.activeEndTime
	case StatusMatch:
		out.Status = gvgguild.GuildGVG_Matching
		out.RemainTimestamp = this_.activeEndTime
	}

	return out, nil
}

func (this_ *Service) QueryActiveGuild(c *ctx.Context, req *gvgguild.GuildGVG_QueryActiveGuildRequest) (*gvgguild.GuildGVG_QueryActiveGuildResponse, *errmsg.ErrMsg) {
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}

	resp := &gvgguild.GuildGVG_QueryActiveGuildResponse{GuildId: req.Guild}
	saveDB := make([]interface{}, 0, 2)
	_, err = eventloop.CallChanEventLoop(this_.signupQueue, resp, func(out *gvgguild.GuildGVG_QueryActiveGuildResponse) (struct{}, *errmsg.ErrMsg) {
		bri, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return struct{}{}, errmsg.NewErrorGuildGVGNotJoin()
		}
		groupId, ok := this_.guildGroupInfo[resp.GuildId]
		if !ok {
			return struct{}{}, errmsg.NewErrorGuildGVGNotSignup()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		var gii *GuildInfo
		this_.calcGuildScoreGroup(group)
		SortGuildInfos(group.Infos)
		for i := range group.Infos {
			v := &group.Infos[i]
			if v.Id == resp.GuildId {
				gii = v
				resp.Rank = int64(i + 1)
				break
			}
		}
		if gii == nil {
			return struct{}{}, errmsg.NewErrorGuildGVGNotSignup()
		}
		r := rule.MustGetReader(c)
		resp.Score = gii.Score
		var selfBuild *BuildInfo
		for i := range gii.Data {
			bv := &gii.Data[i]
			build := &gvgguild.GuildGVG_BuildInfo{
				Id: bv.Id,
			}
			if bv.MaxBlood > 0 {
				build.BloodPer = int64(math.Ceil(float64(bv.Blood) / float64(bv.MaxBlood) * 100))
			}
			surviveNum := int64(0)
			for ii := range bv.Roles {
				rv := &bv.Roles[ii]
				if rv.IsHead {
					build.Nickname = rv.RoleId    //用昵称字段代替ID，出去查询玩家信息
					build.HeadIsDead = rv.IsDeath // 守将是否死亡
				}
				if !rv.IsDeath {
					surviveNum++
				}
			}
			bc, ok := r.GuildContendbuild.GetGuildContendbuildById(bv.Id)
			utils.MustTrue(ok)
			build.IsSmoke = bv.Blood < bc.BuildHp
			build.RoleSurviveNum = surviveNum
			build.RoleTotalNum = int64(len(bv.Roles))
			resp.Builds = append(resp.Builds, build)
			if bv.Id == bri.BuildId {
				selfBuild = bv
			}
		}
		// 查找自己的信息，剩余可挑战次数。最大可挑战次数，下次增加挑战次数时间
		attackCount, needSave := this_.getAttackCount(bri)
		resp.CanFightCount = attackCount
		resp.NextAddTimesTimestamp = bri.NextAddTimes
		if needSave {
			bis := CreateBuildInfoSave(groupId, gii.Id, selfBuild)
			saveDB = append(saveDB, bis)
		}
		return struct{}{}, nil
	})
	if err != nil {
		return nil, err
	}

	if resp.GuildId != gu.GuildId { // 查询是否有标记权限
		gm, ok, err := guildDao.NewGuildMember(gu.GuildId).GetOne(c, gu.RoleId)
		if err != nil {
			return nil, err
		}
		utils.MustTrue(ok)
		gp, ok := rule.MustGetReader(c).GuildPosition.GetGuildPositionById(gm.Position)
		utils.MustTrue(ok)
		if gp.GvgMark {
			resp.CanMask = true
		}
	}
	// 查询最大可累计战斗次数
	guildContendChallengeNumMax, ok := rule.MustGetReader(c).KeyValue.GetInt64("GuildContendChallengeNumMax")
	utils.MustTrue(ok)
	resp.MaxFightCount = guildContendChallengeNumMax

	g, err := guildDao.NewGuild(resp.GuildId).Get(c)
	if err != nil {
		return nil, err
	}
	resp.Flag = g.Flag
	resp.Level = g.Level
	resp.Name = g.Name
	randId, err := guildDao.GetRankId()
	if err != nil {
		return nil, err
	}
	out := &rank_service.RankService_GetRankValueByOwnerIdResponse{}
	if err := this_.nc.RequestWithOut(c, 0, &rank_service.RankService_GetRankValueByOwnerIdRequest{
		RankId:  randId,
		OwnerId: resp.GuildId,
	}, out); err != nil {
		return nil, err
	}
	if out.RankValue != nil {
		resp.Power = out.RankValue.Value1
	}

	roleIdS := make([]string, 0, len(resp.Builds))
	for i := range resp.Builds {
		v := resp.Builds[i]
		if v.Nickname != "" {
			roleIdS = append(roleIdS, v.Nickname)
		}
	}
	roleMap, err := userDao.GetRoles(c, roleIdS)
	if err != nil {
		return nil, err
	}
	for i := range resp.Builds {
		v := resp.Builds[i]
		if v.Nickname != "" {
			role, ok := roleMap[v.Nickname]
			if ok {
				v.Nickname = role.Nickname
				v.AvatarId = role.AvatarId
				v.AvatarFrame = role.AvatarFrame
				v.Power = role.Power
			}
		}
	}
	if len(saveDB) > 0 {
		this_.AsyncSaveDB(saveDB)
	}
	return resp, nil
}

func (this_ *Service) calcGuildScore(gi *GuildInfo, other ...*GuildInfo) {
	timestamp := timer.Now().UTC().Unix()
	this_.log.Info("calcGuildScoreGroup", zap.String("guild_id", gi.Id))
	var saveDB []interface{}
	sub := timestamp - gi.LastScoreChange
	if sub > 3600 {
		count := sub / 3600
		gi.LastScoreChange += count * 3600
		gi.Score += (this_.addScorePerHour + gi.AddExtScore) * count
		gis := CreateGuildInfoSave(gi)
		saveDB = append(saveDB, gis)
	}
	for _, ogi := range other {
		this_.log.Info("calcGuildScoreGroup", zap.String("guild_id", ogi.Id))
		sub := timestamp - ogi.LastScoreChange
		if sub > 3600 {
			count := sub / 3600
			ogi.LastScoreChange += count * 3600
			ogi.Score += (this_.addScorePerHour + gi.AddExtScore) * count
			gis := CreateGuildInfoSave(ogi)
			saveDB = append(saveDB, gis)
		}
	}
	if len(saveDB) > 0 {
		this_.AsyncSaveDB(saveDB)
	}
}

func (this_ *Service) calcGuildScoreGroup(group *GroupInfo) {
	this_.log.Info("calcGuildScoreGroup", zap.Int64("group_id", group.Id))
	timestamp := timer.Now().UTC().Unix()
	var saveDB []interface{}
	for i := range group.Infos {
		gi := &group.Infos[i]
		sub := timestamp - gi.LastScoreChange
		if sub > 3600 {
			count := sub / 3600
			gi.LastScoreChange += count * 3600
			gi.Score += (this_.addScorePerHour + gi.AddExtScore) * count

			gis := CreateGuildInfoSave(gi)
			saveDB = append(saveDB, gis)
		}
	}
	if len(saveDB) > 0 {
		this_.AsyncSaveDB(saveDB)
		this_.log.Info("calcGuildScoreGroup save", zap.Int64("group_id", group.Id))
	}
}

func (this_ *Service) HandleSignup(c *ctx.Context, _ *gvgguild.GuildGVG_SignupRequest) (*gvgguild.GuildGVG_SignupResponse, *errmsg.ErrMsg) {
	g, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if g.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}

	member, ok, err := guildDao.NewGuildMember(g.GuildId).GetOne(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	r := rule.MustGetReader(c)
	gp, ok := r.GuildPosition.GetGuildPositionById(member.Position)
	utils.MustTrue(ok)
	if !gp.GvgSign {
		return nil, errmsg.NewErrorGuildGVGSignupNoAccess()
	}

	resp, err := eventloop.RequestChanEventLoop[*CheckSignInfo, *CheckSignInfo](this_.signupQueue, &CheckSignInfo{
		Id: g.GuildId,
	})
	if err != nil {
		return nil, err
	}
	if resp.IsSignup {
		return nil, errmsg.NewErrorGuildGVGAlreadySignup()
	}

	out := &guild_filter_service.Guild_GuildGVGInfoResponse{}
	err = this_.nc.RequestWithOut(c, 0, &guild_filter_service.Guild_GuildGVGInfoRequest{Id: g.GuildId}, out)
	if err != nil {
		return nil, err
	}
	if out.Active == 0 {
		return nil, errmsg.NewErrorGuildGVGSignupFailedActive()
	}
	guildCnf, ok := r.Guild.GetGuildById(out.Level)
	utils.MustTrue(ok)

	fo := false
	for _, v := range guildCnf.FunctionOpen {
		if v == GuildFunctionGVG {
			fo = true
			break
		}
	}
	if !fo {
		return nil, errmsg.NewErrorGuildGVGSignupFailedFunction()
	}
	needNum, ok := r.KeyValue.GetInt64("GuildContendNum")
	utils.MustTrue(ok)
	if needNum > out.Count {
		return nil, errmsg.NewErrorGuildGVGSignupFailedMember()
	}
	role, err := userDao.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	setInfo, err := eventloop.RequestChanEventLoop[*CheckAndSetSignInfo, *CheckAndSetSignInfo](this_.signupQueue, &CheckAndSetSignInfo{
		Id:       g.GuildId,
		NickName: role.Nickname,
	})
	if err != nil {
		return nil, err
	}
	if !setInfo.SignupSuccess {
		return nil, errmsg.NewErrorGuildGVGAlreadySignup()
	}
	_, _, err = this_.mysql.Exec("INSERT INTO gvg.signup(act_id,guild_id,nick_name)VALUES(?,?,?) ON DUPLICATE KEY UPDATE create_time=NOW()", this_.activeId, setInfo.Id, setInfo.NickName)
	if err != nil {
		_, _ = eventloop.RequestChanEventLoop[*DeleteSignupInfo, *DeleteSignupInfo](this_.signupQueue, &DeleteSignupInfo{setInfo.Id})
		return nil, err
	}
	return &gvgguild.GuildGVG_SignupResponse{}, nil
}

func (this_ *Service) MaskGuild(c *ctx.Context, req *gvgguild.GuildGVG_MaskGuildRequest) (*gvgguild.GuildGVG_MaskGuildResponse, *errmsg.ErrMsg) {
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	if gu.GuildId == req.GuildId {
		return nil, errmsg.NewErrorGuildGVGCanNotMaskSelf()
	}
	gm, ok, err := guildDao.NewGuildMember(gu.GuildId).GetOne(c, gu.RoleId)
	if err != nil {
		return nil, err
	}
	utils.MustTrue(ok)
	gp, ok := rule.MustGetReader(c).GuildPosition.GetGuildPositionById(gm.Position)
	utils.MustTrue(ok)
	if !gp.GvgMark {
		return nil, errmsg.NewErrorGuildGVGNotCanMask()
	}
	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, req.IsTry, func(try bool) (*gvgguild.GuildGVG_MaskGuildResponse, *errmsg.ErrMsg) {
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		groupId, ok := this_.guildGroupInfo[gu.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotSignup()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		var selfGuild *GuildInfo
		var findFlagGuildId bool
		for i := range group.Infos {
			v := &group.Infos[i]
			if v.Id == gu.GuildId {
				if v.FlagGuildId != "" && try {
					return &gvgguild.GuildGVG_MaskGuildResponse{
						Success: false,
					}, nil
				}
				selfGuild = v
				if findFlagGuildId {
					break
				}
			}
			if v.Id == req.GuildId {
				findFlagGuildId = true
				if selfGuild != nil {
					break
				}
			}
		}
		if !findFlagGuildId { // 如果没找到需要标记的工会。则报错
			return nil, errmsg.NewErrorGuildGVGNotCanMask()
		}
		selfGuild.FlagGuildId = req.GuildId
		return &gvgguild.GuildGVG_MaskGuildResponse{
			Success: true,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func CanEnterBuild(c *ctx.Context, gi *GuildInfo, buildId int64) (bool, *errmsg.ErrMsg) {
	r := rule.MustGetReader(c)
	buildCnf, ok := r.GuildContendbuild.GetGuildContendbuildById(buildId)
	if !ok {
		return false, errmsg.NewErrInvalidRequestParam()
	}
	if buildCnf.BuildPriority == 1 {
		return true, nil
	}
	bp := buildCnf.BuildPriority - 1

	for i := range gi.Data {
		v := &gi.Data[i]
		if v.Blood == 0 && bp == v.Priority {
			return true, nil
		}
	}
	return false, nil
}

func (this_ *Service) QueryBuildInfo(c *ctx.Context, req *gvgguild.GuildGVG_QueryBuildInfoRequest) (*gvgguild.GuildGVG_QueryBuildInfoResponse, *errmsg.ErrMsg) {
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}

	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, gu.GuildId, func(_ string) (*gvgguild.GuildGVG_QueryBuildInfoResponse, *errmsg.ErrMsg) {
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}

		groupId, ok := this_.guildGroupInfo[req.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		var gi *GuildInfo
		for i := range group.Infos {
			v := &group.Infos[i]
			if v.Id == req.GuildId {
				gi = v
				break
			}
		}
		if gi == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if gu.GuildId == req.GuildId {

		} else {

		}
		canEnter := true
		if gu.GuildId != req.GuildId {
			canEnter, err = CanEnterBuild(c, gi, req.BuildId)
			if err != nil {
				return nil, err
			}

			if !canEnter {
				return &gvgguild.GuildGVG_QueryBuildInfoResponse{
					CanEnter: false,
				}, nil
			}
		}

		var bi *BuildInfo
		for i := range gi.Data {
			v := &gi.Data[i]
			if v.Id == req.BuildId {
				bi = v
				break
			}
		}
		if bi == nil {
			return nil, errmsg.NewErrorGuildGVGNotFoundBuild()
		}
		out := &gvgguild.GuildGVG_QueryBuildInfoResponse{
			BuildId:  bi.Id,
			Blood:    bi.Blood,
			BloodPer: int64(math.Ceil(float64(bi.Blood) / float64(bi.MaxBlood) * 100)),
			Roles:    make([]*gvgguild.GuildGVG_BuildRoleInfo, 0, len(bi.Roles)),
			CanEnter: true,
		}
		for i := range bi.Roles {
			v := &bi.Roles[i]
			bri := &gvgguild.GuildGVG_BuildRoleInfo{
				RoleId: v.RoleId,
				IsHead: v.IsHead,
			}
			if v.IsDeath {
				bri.Status = 2 // 死亡
			} else {
				if v.IsFighting && time.Now().UTC().Unix()-v.StartFightTime < onceFightSeconds {
					bri.Status = 1 // 战斗中
				} else {
					bri.Status = 0
				}
			}
			out.Roles = append(out.Roles, bri)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	if !resp.CanEnter {
		return resp, nil
	}
	rLen := len(resp.Roles)
	if rLen == 0 {
		return resp, nil
	}
	roleIds := make([]string, 0, rLen)
	for _, v := range resp.Roles {
		roleIds = append(roleIds, v.RoleId)
	}

	roleMap, err := userDao.GetRoles(c, roleIds)
	if err != nil {
		return nil, err
	}
	formationMap, err := formationDao.GetMulti(c, roleIds)
	for _, v := range resp.Roles {
		role, ok := roleMap[v.RoleId]
		if ok {
			v.Power = role.Power
			v.Nickname = role.Nickname
			v.Level = role.Level
			v.AvatarId = role.AvatarId
			v.AvatarFrame = role.AvatarFrame
		}
		formation, ok := formationMap[v.RoleId]
		if ok {
			v.HeroConfig_0 = formation.Assembles[formation.DefaultIndex].HeroOrigin_0
			v.HeroConfig_0 = formation.Assembles[formation.DefaultIndex].HeroOrigin_1
		}
	}

	return resp, nil
}

func (this_ *Service) FightRole(c *ctx.Context, req *gvgguild.GuildGVG_FightRoleRequest) (*gvgguild.GuildGVG_FightRoleResponse, *errmsg.ErrMsg) {
	if req.RoleId == "" || req.GuildId == "" || req.BuildId <= 0 {
		return nil, errmsg.NewErrInvalidRequestParam()
	}
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	var bufferIdS []int64
	saveDB := make([]interface{}, 0, 2)
	if gu.GuildId == req.GuildId {
		return nil, errmsg.NewErrorGuildGVGCanNotFightSelfGuild()
	}
	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, gu.GuildId, func(_ string) (*gvgguild.GuildGVG_FightRoleResponse, *errmsg.ErrMsg) {
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		attackCount, _ := this_.getAttackCount(br)
		if attackCount <= 0 {
			return nil, errmsg.NewErrorGuildGVGNoAttackCount()
		}
		groupId, ok := this_.guildGroupInfo[req.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		var gi, selfGi *GuildInfo
		for i := range group.Infos {
			v := &group.Infos[i]
			if v.Id == req.GuildId {
				gi = v
				if selfGi != nil {
					break
				}
			}
			if v.Id == gu.GuildId {
				selfGi = v
				if gi != nil {
					break
				}
			}
		}
		if gi == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if selfGi == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}

		bufferIdS = append(bufferIdS, selfGi.AddBuffIds...)
		var bi *BuildInfo
		for i := range gi.Data {
			v := &gi.Data[i]
			if v.Id == req.BuildId {
				bi = v
				break
			}
		}
		if bi == nil {
			return nil, errmsg.NewErrorGuildGVGNotFoundBuild()
		}
		var bri *BuildRole
		for i := range bi.Roles {
			v := &bi.Roles[i]
			if v.RoleId == req.RoleId {
				bri = v
				break
			}
		}
		if bri == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if bri.IsDeath {
			return nil, errmsg.NewErrorGuildGVGFightDead()
		}
		now := time.Now().UTC().Unix()
		if bri.IsFighting && now-bri.StartFightTime < onceFightSeconds {
			return nil, errmsg.NewErrorGuildGVGFightInFight()
		}
		bri.StartFightTime = now
		bri.Attacker = c.RoleId
		// TODO
		// br.CanAttackCount--
		/////////保存数据/////
		{

			bis := CreateBuildInfoSave(groupId, gi.Id, bi)
			saveDB = append(saveDB, bis)

			for i := range selfGi.Data {
				v := &selfGi.Data[i]
				if v.Id == br.BuildId {
					CreateBuildInfoSave(groupId, selfGi.Id, v)
					saveDB = append(saveDB, bis)
					break
				}
			}

		}
		//////////////////

		return &gvgguild.GuildGVG_FightRoleResponse{GuildId: req.GuildId, RoleId: req.RoleId}, nil
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_, _ = eventloop.CallChanEventLoop(this_.signupQueue, resp.RoleId, func(roleId string) (bool, *errmsg.ErrMsg) {
				brt, ok := this_.userGroupInfo[roleId]
				if ok {
					brt.StartFightTime -= onceFightSeconds
				}
				br, ok := this_.userGroupInfo[c.RoleId]
				if ok {
					br.CanAttackCount++
				}
				return true, nil
			})
		} else {
			this_.AsyncSaveDB(saveDB) // 没有出错再保存
		}
	}()
	resp.BattleId = -10000 // 写死。
	selfOut := &servicepb.GuildGVG_GetGVGFightInfoResponse{}
	err = this_.nc.RequestWithOut(c, c.InServerId, &servicepb.GuildGVG_GetGVGFightInfoRequest{RoleId: c.RoleId}, selfOut)
	if err != nil {
		return nil, err
	}
	targetRole, err := userDao.GetRole(c, resp.RoleId)
	if err != nil {
		return nil, err
	}
	user, err := userDao.GetUser(c, targetRole.UserId)
	if err != nil {
		return nil, err
	}
	targetOut := &servicepb.GuildGVG_GetGVGFightInfoResponse{}
	err = this_.nc.RequestWithOut(c, user.ServerId, &servicepb.GuildGVG_GetGVGFightInfoRequest{RoleId: req.RoleId}, targetOut)
	if err != nil {
		return nil, err
	}
	resp.Sbp = selfOut.Sbp
	resp.Sbp.CountDown = onceFightSeconds
	for _, v := range resp.Sbp.Heroes {
		v.BuffIds = append(v.BuffIds, bufferIdS...)
	}

	resp.Sbp.HostilePlayers = append(resp.Sbp.HostilePlayers, &models.SinglePlayerInfo{
		Role:   targetOut.Sbp.Role,
		Heroes: targetOut.Sbp.Heroes,
	})
	resp.BuildId = req.BuildId
	return resp, nil
}

func (this_ *Service) FightRoleFinish(c *ctx.Context, req *gvgguild.GuildGVG_FightRoleFinishRequest) (*gvgguild.GuildGVG_FightRoleFinishResponse, *errmsg.ErrMsg) {
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	var cfi ChatFightInfo
	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, gu.GuildId, func(_ string) (*gvgguild.GuildGVG_FightRoleFinishResponse, *errmsg.ErrMsg) {
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		groupId, ok := this_.guildGroupInfo[req.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		var gi, selfGi *GuildInfo
		for i := range group.Infos {
			v := &group.Infos[i]
			if v.Id == req.GuildId {
				gi = v
				if selfGi != nil {
					break
				}
			}
			if v.Id == gu.GuildId {
				selfGi = v
				if gi != nil {
					break
				}
			}
		}
		if gi == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if selfGi == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		this_.calcGuildScore(gi, selfGi)
		var gib *BuildInfo
		for i := range gi.Data {
			v := &gi.Data[i]
			if v.Id == req.BuildId {
				gib = v
				break
			}
		}
		if gib == nil {
			return nil, errmsg.NewErrorGuildGVGNotFoundBuild()
		}
		exists := false
		var defender *BuildRole
		for i := range gib.Roles {
			v := &gib.Roles[i]
			if v.RoleId == req.RoleId {
				if v.Attacker == c.RoleId {
					exists = true
					defender = v
				}
				break
			}
		}
		if !exists {
			return nil, errmsg.NewErrorGuildGVGFightInvalid()
		}
		attackIntegralIndex := int64(2)
		defendIntegralIndex := int64(5)
		subHpKey := "GuildContendFailBuildHp"
		if req.Success { // 成功，被攻击的必定死了
			attackIntegralIndex = 1
			defendIntegralIndex = 6
			subHpKey = "GuildContendWinBuildHp"
			defender.IsDeath = true
		}
		r := rule.MustGetReader(c)
		atkScore, ok := r.GuildIntegral.GetGuildIntegralById(attackIntegralIndex)
		utils.MustTrue(ok)
		defendScore, ok := r.GuildIntegral.GetGuildIntegralById(defendIntegralIndex)
		utils.MustTrue(ok)
		subHp, ok := r.KeyValue.GetInt64(subHpKey)
		utils.MustTrue(ok)
		br.Score += atkScore.RankNum
		br.ScoreChangeTime = c.StartTime
		guildAddScore := atkScore.RankNum
		if defender.IsHead && defender.IsDeath {
			switch gib.Id {
			case 1:
				build, ok := r.GuildContendbuild.GetGuildContendbuildById(1)
				utils.MustTrue(ok && len(build.BuildEff) == 2)
				extScore := int64(0)
				if build.BuildEff[0] == 1 {
					extScore += gi.Score * build.BuildEff[1] / int64(percent.BASE)
				} else {
					extScore += build.BuildEff[1]
				}
				guildAddScore += extScore
				gi.Score -= extScore

			case 2:
				build, ok := r.GuildContendbuild.GetGuildContendbuildById(1)
				utils.MustTrue(ok && len(build.BuildEff) == 2)
				selfGi.AddBuffIds = append(selfGi.AddBuffIds, build.BuildEff[1])
			case 3:
				build, ok := r.GuildContendbuild.GetGuildContendbuildById(1)
				utils.MustTrue(ok && len(build.BuildEff) == 2)
				selfGi.AddExtScore += build.BuildEff[1]
			case 4:
				build, ok := r.GuildContendbuild.GetGuildContendbuildById(1)
				utils.MustTrue(ok && len(build.BuildEff) == 2)
				selfGi.AddBuffIds = append(selfGi.AddBuffIds, build.BuildEff[1])
			default:

			}
		}

		selfGi.Score += guildAddScore
		defender.Score += defendScore.RankNum
		defender.ScoreChangeTime = c.StartTime
		gi.Score += defendScore.RankNum
		build, ok := r.GuildContendbuild.GetGuildContendbuildById(gib.Id)
		utils.MustTrue(ok)
		lowestHp := gib.MaxBlood * build.BuildGeneral / int64(percent.BASE)
		if defender.IsHead { // 如果是首领死了，扣除百分比的血量
			subHp += lowestHp
		} else { //如果首领没死，最少保留百分比的血量
			if gib.Blood-subHp < lowestHp {
				subHp = gib.Blood - lowestHp
			}
		}
		if gib.Blood-subHp < 0 {
			subHp = gib.Blood
		}
		gib.Blood -= subHp
		br.BuildHurt += subHp
		defender.IsFighting = false
		defender.StartFightTime = 0
		defender.Attacker = ""
		group.sl.Insert(br.RoleId, &models.RankValue{
			OwnerId:   br.RoleId,
			Value1:    br.Score,
			CreatedAt: br.ScoreChangeTime,
		})
		group.sl.Insert(defender.RoleId, &models.RankValue{
			OwnerId:   defender.RoleId,
			Value1:    defender.Score,
			CreatedAt: defender.ScoreChangeTime,
		})

		out := &gvgguild.GuildGVG_FightRoleFinishResponse{
			Success:          req.Success,
			PersonalScoreAdd: atkScore.RankNum,
			GuildScoreAdd:    guildAddScore,
			BuildId:          req.BuildId,
			RemainBlood:      gib.Blood,
			SubBlood:         subHp,
		}
		this_.fightIndex++
		fi := &FightInfo{
			Id:               this_.fightIndex,
			GroupId:          group.Id,
			AttackGuildId:    br.GuildId,
			Attacker:         br.RoleId,
			DefendGuildId:    defender.GuildId,
			Defender:         defender.RoleId,
			IsBuilder:        0,
			Blood:            subHp,
			IsWin:            0,
			CreateTime:       time.Now().UTC(),
			PersonalScoreAdd: atkScore.RankNum,
			GuildScoreAdd:    guildAddScore,
		}
		if req.Success {
			fi.IsWin = 1
		}
		group.fiMap[fi.Id] = fi
		gi.LastFightInfo = fi.Id
		br.FightLog = append(br.FightLog, fi.Id)
		defender.FightLog = append(defender.FightLog, fi.Id)
		//////组装聊天数据///
		{
			cfi.GroupId = groupId
			cfi.AttackGuildId = fi.AttackGuildId
			cfi.Attacker = fi.Attacker
			cfi.DefendGuildId = fi.DefendGuildId
			cfi.Defender = fi.Defender
			cfi.IsBuilder = fi.IsBuilder
			cfi.Blood = fi.Blood
			cfi.IsWin = fi.IsWin
			cfi.PersonalScoreAdd = fi.PersonalScoreAdd
			cfi.GuildScoreAdd = fi.GuildScoreAdd
			cfi.BuildId = fi.BuildId
		}
		//////////////
		///////写数据库///////
		{
			saveDB := make([]interface{}, 0, 5)

			selfGis := CreateGuildInfoSave(selfGi)
			saveDB = append(saveDB, selfGis)

			if gib.Id == 1 && defender.IsHead && defender.IsDeath {
				targetGis := CreateGuildInfoSave(gi)
				saveDB = append(saveDB, targetGis)
			}

			bis := CreateBuildInfoSave(gi.GroupId, gi.Id, gib)
			saveDB = append(saveDB, bis)

			for i := range selfGi.Data {
				v := &selfGi.Data[i]
				if v.Id == br.BuildId {
					bis := CreateBuildInfoSave(selfGi.GroupId, selfGi.Id, v)
					saveDB = append(saveDB, bis)
					break
				}
			}

			saveDB = append(saveDB, (*FightingSave)(fi))
			this_.AsyncSaveDB(saveDB)
		}
		////////////////////
		return out, nil
	})
	if err != nil {
		return nil, err
	}

	var guildIds = []string{cfi.AttackGuildId, cfi.DefendGuildId}
	guildMap, err := guildDao.NewGuild("").GetMulti(c, guildIds)
	if err == nil { //这里不能返回错误，不然需要处理战斗回退
		defendGuild, ok := guildMap[req.GuildId]
		if ok {
			resp.Name = defendGuild.Name
			cfi.DefendGuildName = defendGuild.Name
		}
		attackGuild, ok := guildMap[cfi.AttackGuildId]
		if ok {
			cfi.AttackGuildName = attackGuild.Name
		}
	}
	roleIdS := make([]string, 0, 2)
	roleIdS = append(roleIdS, cfi.Attacker, cfi.Defender)
	roleMap, err := userDao.GetRoles(c, roleIdS)
	if err == nil {
		defendRole, ok := roleMap[req.RoleId]
		if ok {
			cfi.DefenderName = defendRole.Nickname
		}
		attackRole, ok := roleMap[cfi.Attacker]
		if ok {
			cfi.AttackerName = attackRole.Nickname
		}
	}
	this_.sendGVGFightInfo(&cfi)
	return resp, nil
}

func (this_ *Service) sendGVGFightInfo(cfi *ChatFightInfo) {
	gopool.Submit(func() {
		msg, err := jsoniter.MarshalToString(cfi)
		if err != nil {
			this_.log.Error("jsoniter.MarshalToString ChatFightInfo error", zap.Error(err))
			return
		}
		roomId := GetChatRoomID(cfi.GroupId)
		err = im.DefaultClient.SendMessage(context.Background(), &im.Message{
			Type:      im.MsgTypeRoom,
			RoleID:    "system",
			RoleName:  "system",
			TargetID:  roomId,
			Content:   msg,
			ParseType: im.ParseTypeGVGGuild,
		})
		if err != nil {
			this_.log.Error("im.DefaultClient.SendMessage sendGVGFightInfo error", zap.Error(err), zap.String("msg", msg), zap.String("roomId", roomId))
		}
	})
}

func (this_ *Service) CheckRelive(t int64, changeGroupId int64, guildId string) {
	util := time.Unix(t, 0)
	timer.UntilFunc(util, func() {
		this_.signupQueue.PostFuncQueue(func() {
			groupId, ok := this_.guildGroupInfo[guildId]
			if !ok || groupId != changeGroupId {
				return
			}
			group, ok := this_.groupMap[groupId]
			if !ok {
				return
			}
			var guild *GuildInfo
			for i := range group.Infos {
				v := &group.Infos[i]
				if v.Id == guildId {
					guild = v
					break
				}
			}
			if guild == nil {
				return
			}
			if guild.ReliveTime != t {
				return
			}
			saveDB := make([]interface{}, 0, len(guild.Data))
			for i := range guild.Data {
				v := &guild.Data[i]
				v.Blood = v.MaxBlood
				for ii := range v.Roles {
					vv := &v.Roles[ii]
					vv.IsDeath = false
				}
				bis := CreateBuildInfoSave(groupId, guild.Id, v)
				saveDB = append(saveDB, bis)
			}
			if len(saveDB) > 0 {
				this_.AsyncSaveDB(saveDB)
			}
		})
	})
}

func (this_ *Service) getAttackCount(br *BuildRole) (int64, bool) {
	if br.NextAddTimes == 0 {
		return br.CanAttackCount, false
	}
	now := time.Now().UTC().Unix()
	save := false
	for now > br.NextAddTimes && br.NextAddTimes != 0 {
		save = true
		br.CanAttackCount++
		if br.CanAttackCount < this_.maxAttackCount {
			br.NextAddTimes += this_.addAttackCountSeconds
		} else {
			br.NextAddTimes = 0
		}
	}
	return br.CanAttackCount, save
}

func (this_ *Service) subAttackCount(br *BuildRole) {
	this_.getAttackCount(br)
	br.CanAttackCount--
	if br.CanAttackCount < this_.maxAttackCount && br.NextAddTimes == 0 {
		br.NextAddTimes = time.Now().UTC().Add(time.Second * time.Duration(this_.addAttackCountSeconds)).Unix()
	}
}

func (this_ *Service) FightBuild(c *ctx.Context, req *gvgguild.GuildGVG_FightBuildRequest) (*gvgguild.GuildGVG_FightBuildResponse, *errmsg.ErrMsg) {
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	if gu.GuildId == req.GuildId {
		return nil, errmsg.NewErrorGuildGVGCanNotFightSelfGuild()
	}
	var cfi ChatFightInfo
	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, gu.GuildId, func(_ string) (*gvgguild.GuildGVG_FightBuildResponse, *errmsg.ErrMsg) {
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		attackCount, _ := this_.getAttackCount(br)
		if attackCount <= 0 {
			return nil, errmsg.NewErrorGuildGVGNoAttackCount()
		}

		groupId, ok := this_.guildGroupInfo[req.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		var gi, selfGi *GuildInfo
		for i := range group.Infos {
			v := &group.Infos[i]
			if v.Id == req.GuildId {
				gi = v
				if selfGi != nil {
					break
				}
			}
			if v.Id == gu.GuildId {
				selfGi = v
				if gi != nil {
					break
				}
			}
		}
		if gi == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if selfGi == nil {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		this_.calcGuildScore(gi, selfGi)
		var gib *BuildInfo
		for i := range gi.Data {
			v := &gi.Data[i]
			if v.Id == req.BuildId {
				gib = v
				break
			}
		}
		if gib == nil {
			return nil, errmsg.NewErrorGuildGVGNotFoundBuild()
		}
		if gib.Blood <= 0 {
			return nil, errmsg.NewErrorGuildGVGNotCanFightBuild()
		}
		canAttackBuild := true
		for i := range gib.Roles {
			v := &gib.Roles[i]
			if !v.IsDeath {
				canAttackBuild = false
				break
			}
		}
		if !canAttackBuild {
			return nil, errmsg.NewErrorGuildGVGNotCanFightBuild()
		}

		// TODO
		// br.CanAttackCount--
		now := time.Now().UTC().Unix()
		r := rule.MustGetReader(c)
		guildContendChallengeTime, ok := r.KeyValue.GetInt64("GuildContendChallengeTime")
		utils.MustTrue(ok)
		if now-br.StartFightTime < guildContendChallengeTime {
			br.StartFightTime = now
		}
		guildContendBuildHp, ok := r.KeyValue.GetInt64("GuildContendBuildHp")
		utils.MustTrue(ok)
		if gib.Blood-guildContendBuildHp < 0 {
			guildContendBuildHp = gib.Blood
		}
		gib.Blood -= guildContendBuildHp
		integralId := int64(3)
		if gib.Blood <= 0 {
			gib.Blood = 0
			// 被摧毁
			integralId = 4
			gi.ReliveTime = time.Now().UTC().Unix()
		}
		br.BuildHurt += guildContendBuildHp
		giScore, ok := r.GuildIntegral.GetGuildIntegralById(integralId)
		utils.MustTrue(ok)

		br.Score += giScore.RankNum
		br.ScoreChangeTime = c.StartTime
		selfGi.Score += giScore.RankNum
		group.sl.Insert(br.RoleId, &models.RankValue{
			OwnerId:   br.RoleId,
			Value1:    br.Score,
			CreatedAt: br.ScoreChangeTime,
		})
		out := &gvgguild.GuildGVG_FightBuildResponse{
			Success:          true,
			PersonalScoreAdd: giScore.RankNum,
			GuildScoreAdd:    giScore.RankNum,
			BuildId:          req.BuildId,
			RemainBlood:      gib.Blood,
			SubBlood:         -guildContendBuildHp,
		}
		this_.fightIndex++
		fi := &FightInfo{
			Id:               this_.fightIndex,
			GroupId:          group.Id,
			AttackGuildId:    br.GuildId,
			Attacker:         br.RoleId,
			DefendGuildId:    req.GuildId,
			IsBuilder:        1,
			Blood:            guildContendBuildHp,
			IsWin:            1,
			CreateTime:       time.Now().UTC(),
			PersonalScoreAdd: giScore.RankNum,
			GuildScoreAdd:    giScore.RankNum,
			BuildId:          req.BuildId,
		}
		group.fiMap[fi.Id] = fi
		gi.LastFightInfo = fi.Id
		br.FightLog = append(br.FightLog, fi.Id)
		//////组装聊天数据///

		{
			cfi.GroupId = groupId
			cfi.AttackGuildId = fi.AttackGuildId
			cfi.Attacker = fi.Attacker
			cfi.DefendGuildId = fi.DefendGuildId
			cfi.Defender = fi.Defender
			cfi.IsBuilder = fi.IsBuilder
			cfi.Blood = fi.Blood
			cfi.IsWin = fi.IsWin
			cfi.PersonalScoreAdd = fi.PersonalScoreAdd
			cfi.GuildScoreAdd = fi.GuildScoreAdd
			cfi.BuildId = fi.BuildId
		}
		//////////////

		{ // 写数据库
			saveDB := make([]interface{}, 0, 4)
			selfGis := CreateGuildInfoSave(selfGi)
			saveDB = append(saveDB, selfGis)
			if integralId == 4 && req.BuildId == 1 { // 英灵殿被摧毁了
				gis := CreateGuildInfoSave(gi)
				saveDB = append(saveDB, gis)
			}

			bis := CreateBuildInfoSave(gi.GroupId, gi.Id, gib)
			saveDB = append(saveDB, bis)

			for i := range selfGi.Data {
				v := &selfGi.Data[i]
				if v.Id == br.BuildId {
					selfBis := CreateBuildInfoSave(selfGi.GroupId, selfGi.Id, v)
					saveDB = append(saveDB, selfBis)
					break
				}
			}

			saveDB = append(saveDB, (*FightingSave)(fi))
			this_.AsyncSaveDB(saveDB)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	var guildIds = []string{cfi.AttackGuildId, cfi.DefendGuildId}
	guildMap, err := guildDao.NewGuild("").GetMulti(c, guildIds)
	if err == nil { //这里不能返回错误，不然需要处理战斗回退
		defendGuild, ok := guildMap[req.GuildId]
		if ok {
			resp.Name = defendGuild.Name
			cfi.DefendGuildName = defendGuild.Name
		}
		attackGuild, ok := guildMap[cfi.AttackGuildId]
		if ok {
			cfi.AttackGuildName = attackGuild.Name
		}
	}
	role, err := userDao.GetRole(c, cfi.Attacker)
	if err == nil {
		cfi.AttackerName = role.Nickname
	}
	this_.sendGVGFightInfo(&cfi)
	return resp, nil
}

func (this_ *Service) QueryGuildRank(c *ctx.Context, _ *gvgguild.GuildGVG_QueryGuildRankRequest) (*gvgguild.GuildGVG_QueryGuildRankResponse, *errmsg.ErrMsg) {
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, gu.GuildId, func(_ string) ([]*gvgguild.GuildGVG_GuildRankInfo, *errmsg.ErrMsg) {
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		groupId, ok := this_.guildGroupInfo[gu.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		this_.calcGuildScoreGroup(group)
		SortGuildInfos(group.Infos)
		out := make([]*gvgguild.GuildGVG_GuildRankInfo, 0, len(group.Infos))
		for i := range group.Infos {
			v := &group.Infos[i]
			totalBlood := int64(0)
			currBlood := int64(0)
			build1Per := int64(0)
			for ii := range v.Data {
				vv := &v.Data[ii]
				totalBlood += vv.MaxBlood
				currBlood += vv.Blood
				if vv.Id == 1 && vv.MaxBlood > 0 {
					build1Per = int64(math.Ceil(float64(vv.Blood) / float64(vv.MaxBlood) * 100.0))
				}
			}
			allBuildPer := int64(0)
			if totalBlood > 0 {
				allBuildPer = int64(math.Ceil(float64(currBlood) / float64(totalBlood) * 100.0))
			}

			gri := &gvgguild.GuildGVG_GuildRankInfo{
				Id: v.Id,
				//Flag:       0,
				//Name:       "",
				Score: v.Score,
				Rank:  int64(i + 1),
				//Power:      0,
				Build_1Per: build1Per,
				TotalPer:   allBuildPer,
			}
			out = append(out, gri)
		}
		return out, nil
	})
	if err != nil {
		return nil, err
	}
	guildIds := make([]string, 0, len(resp))
	for _, v := range resp {
		guildIds = append(guildIds, v.Id)
	}
	guildMap, err := guildDao.NewGuild(gu.GuildId).GetMulti(c, guildIds)
	if err != nil {
		return nil, err
	}
	for _, v := range resp {
		guild, ok := guildMap[v.Id]
		if ok {
			v.Name = guild.Name
			v.Flag = guild.Flag
			v.Level = guild.Level
		}
	}
	guildRankId, err := guildDao.GetRankId()
	if err != nil {
		return nil, err
	}
	rankOut := &rank_service.RankService_GetRankValueSByOwnerIdSResponse{}
	err = this_.nc.RequestWithOut(c, 0, &rank_service.RankService_GetRankValueSByOwnerIdSRequest{RankId: guildRankId, OwnerIds: guildIds}, rankOut)
	if err != nil {
		return nil, err
	}
	for _, v := range resp {
		rank, ok := rankOut.Ranks[v.Id]
		if ok {
			v.Power = rank.Value1
		}
	}
	return &gvgguild.GuildGVG_QueryGuildRankResponse{
		Ranks: resp,
	}, nil
}

func (this_ *Service) QueryPersonalRank(c *ctx.Context, req *gvgguild.GuildGVG_QueryPersonalRankRequest) (*gvgguild.GuildGVG_QueryPersonalRankResponse, *errmsg.ErrMsg) {
	if req.Count > 100 || req.Count <= 0 {
		req.Count = 100
	}
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, gu.GuildId, func(_ string) (*gvgguild.GuildGVG_QueryPersonalRankResponse, *errmsg.ErrMsg) {
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		groupId, ok := this_.guildGroupInfo[gu.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		group, ok := this_.groupMap[groupId]
		utils.MustTrue(ok)
		if req.Offset <= 0 {
			req.Offset = 1
		}
		ranks := make([]*gvgguild.GuildGVG_PersonalRankInfo, 0, req.Count)
		elem := group.sl.FindByRank(int(req.Offset))
		index := int64(0)
		var selfRank *gvgguild.GuildGVG_PersonalRankInfo
		for elem != nil {
			rv := elem.GetVal().(*models.RankValue)
			buildRole := this_.userGroupInfo[rv.OwnerId]
			rank := &gvgguild.GuildGVG_PersonalRankInfo{
				Rank:      req.Offset + index,
				Id:        buildRole.RoleId,
				Score:     buildRole.Score,
				KillNum:   buildRole.KillCount,
				BuildHarm: buildRole.BuildHurt,
			}
			ranks = append(ranks, rank)
			if buildRole.RoleId == c.RoleId {
				selfRank = rank
			}
			index++
			if index >= req.Count {
				break
			}
			elem = elem.Next()
		}
		if selfRank == nil {
			selfValue, ok := group.sl.Get(c.RoleId)
			if ok {
				selfRank = &gvgguild.GuildGVG_PersonalRankInfo{}
				rank, _ := group.sl.GetRank(c.RoleId)
				selfRank.Rank = int64(rank)
				rv := selfValue.(*models.RankValue)
				sv := this_.userGroupInfo[rv.OwnerId]
				selfRank.Id = c.RoleId
				selfRank.KillNum = sv.KillCount
				selfRank.BuildHarm = sv.BuildHurt
				selfRank.Score = sv.Score
			}
		}

		return &gvgguild.GuildGVG_QueryPersonalRankResponse{
			Ranks:      ranks,
			Self:       selfRank,
			TotalCount: int64(group.sl.GetNodeCount()),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Ranks) != 0 || resp.Self != nil {
		roleIds := make([]string, 0, len(resp.Ranks)+1)
		for _, v := range resp.Ranks {
			roleIds = append(roleIds, v.Id)
		}
		if resp.Self != nil {
			roleIds = append(roleIds, resp.Self.Id)
		}
		roleMap, err := userDao.GetRoles(c, roleIds)
		if err != nil {
			return nil, err
		}
		if resp.Self != nil {
			role, ok := roleMap[resp.Self.Id]
			if ok {
				resp.Self.Power = role.Power
				resp.Self.Nickname = role.Nickname
				resp.Self.Level = role.Level
				resp.Self.AvatarId = role.AvatarId
				resp.Self.AvatarFrame = role.AvatarFrame
			}
		}
		for _, v := range resp.Ranks {
			role, ok := roleMap[v.Id]
			if ok {
				v.Power = role.Power
				v.Nickname = role.Nickname
				v.Level = role.Level
				v.AvatarId = role.AvatarId
				v.AvatarFrame = role.AvatarFrame
			}

		}
	}

	return resp, nil
}

func (this_ *Service) QueryFightingInfo(c *ctx.Context, req *gvgguild.GuildGVG_QueryFightingInfoRequest) (*gvgguild.GuildGVG_QueryFightingInfoResponse, *errmsg.ErrMsg) {
	if req.Count > 100 || req.Count <= 0 {
		req.Count = 100
	}
	gu, err := guildDao.NewGuildUser(c.RoleId).Get(c)
	if err != nil {
		return nil, err
	}
	if gu.GuildId == "" {
		return nil, errmsg.NewErrGuildNoGuild()
	}
	resp, err := eventloop.CallChanEventLoop(this_.signupQueue, req.RoleId, func(queryRoleId string) (*gvgguild.GuildGVG_QueryFightingInfoResponse, *errmsg.ErrMsg) {
		qbr, ok := this_.userGroupInfo[queryRoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		br, ok := this_.userGroupInfo[c.RoleId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if br.GuildId != gu.GuildId { // 说明可能活动期间离开了工会，换了新公会
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		qgi, ok := this_.guildGroupInfo[qbr.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		qi, ok := this_.guildGroupInfo[br.GuildId]
		if !ok {
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		if qi != qgi { // 说明不在一个分组
			return nil, errmsg.NewErrorGuildGVGNotJoin()
		}
		group, ok := this_.groupMap[qgi]
		utils.MustTrue(ok)
		fl := int64(len(qbr.FightLog))
		index := fl - req.Offset - 1
		if index < 0 {
			return nil, nil
		}
		count := req.Count
		out := make([]*gvgguild.GuildGVG_FightingInfo, 0, count)
		for index >= 0 && count > 0 {
			fId := qbr.FightLog[index]
			logInfo, ok := group.fiMap[fId]
			if ok {
				fi := &gvgguild.GuildGVG_FightingInfo{
					Id:            logInfo.Id,
					Timestamp:     logInfo.CreateTime.UTC().Unix(),
					IsAttack:      logInfo.Attacker == queryRoleId,
					IsAttackBuild: logInfo.IsBuilder == 1,
					//GuildName:        ,
					//NickName:         "",
					BuildId:          logInfo.BuildId,
					Damage:           logInfo.Blood,
					PersonalScoreAdd: logInfo.PersonalScoreAdd,
					GuildScoreAdd:    logInfo.GuildScoreAdd,
				}
				if !fi.IsAttackBuild { //如果是攻击建筑。攻击者就是他自己，不需要处理else
					if fi.IsAttack {
						fi.GuildName = logInfo.DefendGuildId
						fi.NickName = logInfo.Defender
					} else {
						fi.GuildName = logInfo.AttackGuildId
						fi.NickName = logInfo.Attacker
					}
				} else {
					fi.GuildName = logInfo.DefendGuildId
				}
				out = append(out, fi)
			}
			index--
			count--
		}
		return &gvgguild.GuildGVG_QueryFightingInfoResponse{Infos: out, TotalCount: int64(len(qbr.FightLog))}, nil
	})
	if err != nil {
		return nil, err
	}
	guildIds := make([]string, 0, len(resp.Infos))
	roleIds := make([]string, 0, len(resp.Infos))
	for _, v := range resp.Infos {
		if v.GuildName != "" {
			guildIds = append(guildIds, v.GuildName)
		}
		if v.NickName != "" {
			roleIds = append(roleIds, v.NickName)
		}
	}
	var guildMap map[values.GuildId]*dao.Guild
	var roleMap map[string]*dao.Role
	if len(guildIds) != 0 {
		guildMap, err = guildDao.NewGuild("").GetMulti(c, guildIds)
		if err != nil {
			return nil, err
		}
	}
	if len(roleIds) != 0 {
		roleMap, err = userDao.GetRoles(c, roleIds)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range resp.Infos {
		if v.GuildName != "" {
			guild, ok := guildMap[v.GuildName]
			if ok {
				v.GuildName = guild.Name
			}
		}
		if v.NickName != "" {
			role, ok := roleMap[v.NickName]
			if ok {
				v.NickName = role.Nickname
			}
		}
	}
	queryGuild, err := guildDao.NewGuild(gu.GuildId).Get(c)
	if err != nil {
		return nil, err
	}
	resp.RoleGuildName = queryGuild.Name
	return resp, nil
}

func (this_ *Service) DeleteSignupInfo(req *DeleteSignupInfo) (*DeleteSignupInfo, *errmsg.ErrMsg) {
	delete(this_.signupMap, req.Id)
	return req, nil
}

func (this_ *Service) CheckSignup(req *CheckSignInfo) (*CheckSignInfo, *errmsg.ErrMsg) {
	v, ok := this_.signupMap[req.Id]
	req.IsSignup = ok
	if ok {
		req.Nickname = v.NickName
	}

	return req, nil
}

func (this_ *Service) CheckAndSetSignup(req *CheckAndSetSignInfo) (*CheckAndSetSignInfo, *errmsg.ErrMsg) {
	_, ok := this_.signupMap[req.Id]
	if ok {
		req.SignupSuccess = false // 表示已经报名
	}
	this_.signupMap[req.Id] = signupInfo{
		NickName: req.NickName,
		SignTime: time.Now().UTC(),
	}
	req.SignupSuccess = true //表示报名成功
	return req, nil
}
