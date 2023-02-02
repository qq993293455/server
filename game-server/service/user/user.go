package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"coin-server/common/proto/cppbattle"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/eventlocal"
	"coin-server/common/gopool"
	"coin-server/common/handler"
	"coin-server/common/idgenerate"
	"coin-server/common/iggsdk"
	"coin-server/common/im"
	"coin-server/common/logger"
	mapdata "coin-server/common/map-data"
	"coin-server/common/proto/broadcast"
	"coin-server/common/proto/dao"
	lessservicepb "coin-server/common/proto/less_service"
	"coin-server/common/proto/models"
	"coin-server/common/proto/rank_service"
	"coin-server/common/proto/recommend"
	writepb "coin-server/common/proto/role_state_write"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/redisclient"
	"coin-server/common/sensitive"
	"coin-server/common/service"
	"coin-server/common/statistical"
	models2 "coin-server/common/statistical/models"
	"coin-server/common/statistical2"
	models3 "coin-server/common/statistical2/models"
	"coin-server/common/syncrole"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/common/values/enum/ItemType"
	"coin-server/game-server/event"
	"coin-server/game-server/module"
	"coin-server/game-server/service/user/db"
	rule2 "coin-server/game-server/service/user/rule"
	"coin-server/game-server/util/trans"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"

	"github.com/golang-jwt/jwt"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

type Service struct {
	serverId   values.ServerId
	serverType models.ServerType
	svc        *service.Service
	log        *logger.Logger
	*module.Module
}

func NewUserService(
	serverId values.ServerId,
	serverType models.ServerType,
	svc *service.Service,
	module *module.Module,
	log *logger.Logger,
) *Service {
	s := &Service{
		serverId:   serverId,
		serverType: serverType,
		svc:        svc,
		log:        log,
		Module:     module,
	}
	module.UserService = s
	return s
}

func (this_ *Service) Router() {
	this_.svc.RegisterFunc("用户登录", this_.Login)
	this_.svc.RegisterFunc("提玩家下线", this_.KickOffUser)
	this_.svc.RegisterFunc("获取角色信息", this_.GetRoleRequest)
	this_.svc.RegisterFunc("搜索角色信息", this_.SearchRoleRequest)
	this_.svc.RegisterFunc("更换头像", this_.SetAvatar)
	this_.svc.RegisterFunc("更换头像框", this_.SetAvatarFrame)
	this_.svc.RegisterFunc("更换昵称", this_.ChangeNickName)
	this_.svc.RegisterFunc("更换语言", this_.ChangeLanguage)
	this_.svc.RegisterFunc("获取别的玩家", this_.CheatGetRolesRequest)
	this_.svc.RegisterEvent("玩家离线通知", this_.Logout)
	this_.svc.RegisterEvent("修改玩家所在战斗服", this_.ChangeBattleId)
	this_.svc.RegisterFunc("获取头像、头像框列表", this_.GetOwnAvatar)

	this_.svc.RegisterFunc("获取最近聊天列表", this_.GetRecentChatIds)
	this_.svc.RegisterFunc("添加最近聊天列表", this_.AddRecentChatIds)
	this_.svc.RegisterFunc("删除最近聊天列表", this_.DeleteRecentChatIds)
	this_.svc.RegisterFunc("获取用户简版信息", this_.GetSimpleRoles)

	this_.svc.RegisterFunc("获取使用道具立即获得对应时间的挂机经验收益信息", this_.GetUseHangUpExpItemInfo)
	this_.svc.RegisterFunc("使用道具立即获得对应时间的挂机经验收益", this_.UseHangUpExpItem)
	this_.svc.RegisterFunc("升一级", this_.LevelUpgrade)
	// this_.svc.RegisterFunc("升多级", this_.LevelUpgradeMany)
	this_.svc.RegisterFunc("开启突破Boss战", this_.AdvanceOpen)
	this_.svc.RegisterFunc("突破Boss战开始", this_.AdvanceBattleStart)
	this_.svc.RegisterFunc("突破Boss战开结束", this_.AdvanceBattleFinish)
	// this_.svc.RegisterEvent("突破Boss战开结束", this_.AdvanceBattleEnd)
	this_.svc.RegisterFunc("突破Boss战开胜利", this_.AdvanceBattleVictory)
	this_.svc.RegisterFunc("获取突破信息", this_.AdvanceInfo)
	this_.svc.RegisterFunc("使用道具削减挑战boss cd", this_.AdvanceUseItem)

	this_.svc.RegisterFunc("推荐好友", this_.RecommendFriend)
	this_.svc.RegisterFunc("改游戏服时间", this_.CheatModifyTimeRequest)
	this_.svc.RegisterFunc("设置主角等级", this_.CheatSetLevel)
	this_.svc.RegisterFunc("改创建角色时间", this_.CheatModifyCreateTimeRequest)

	this_.svc.RegisterFunc("更新战斗设置", this_.UpdateBattleSettingData)
	this_.svc.RegisterFunc("获取战斗设置", this_.GetBattleSettingData)
	this_.svc.RegisterFunc("获取红点", this_.GetRedPoints)
	this_.svc.RegisterFunc("添加红点", this_.AddRedPoints)
	this_.svc.RegisterFunc("设置红点", this_.SetRedPoints)
	this_.svc.RegisterFunc("获取开场动画Id", this_.GetCutSceneId)
	this_.svc.RegisterFunc("设置开场动画Id", this_.SetCutSceneId)

	this_.svc.RegisterFunc("设置单机战斗速度", this_.SetBattleSpeed)

	this_.svc.RegisterFunc("获取头衔奖励信息", this_.GetTitleRewardsInfo)
	this_.svc.RegisterFunc("领取头衔奖励", this_.DrawTitleRewards)
	this_.svc.RegisterFunc("获取额外特殊技能", this_.GetExtraSkillCntPb)

	this_.svc.RegisterFunc("战力对比", this_.UserCombatValueDetails)

	this_.svc.RegisterFunc("获取玩家信息", this_.GetUserSimpleInfo)
	this_.svc.RegisterFunc("昵称是否存在", this_.NameExist)

	this_.svc.RegisterFunc("融魂重置", this_.CheatResetExpSkip)
	this_.svc.RegisterFunc("领取兑换码", this_.UseCdKey)

	this_.svc.RegisterFunc("作弊推送", this_.CheatPushMsgToMobile)
	this_.svc.RegisterFunc("作弊向前修改玩家注册天数", this_.CheatAheadRegisterDay)

	eventlocal.SubscribeEventLocal(this_.HandleLoginEvent)
	eventlocal.SubscribeEventLocal(this_.HandleLogoutEvent)
	// eventlocal.SubscribeEventLocal(this_.HandleTaskUpdateEvent)
	eventlocal.SubscribeEventLocal(this_.HandleAttrUpdateToRole)
	eventlocal.SubscribeEventLocal(this_.HandleRoleSkillUpdate)
	eventlocal.SubscribeEventLocal(this_.HeroUpdate)
	eventlocal.SubscribeEventLocal(this_.HandleReadPointAdd)
	eventlocal.SubscribeEventLocal(this_.HandleReadPointChange)
	eventlocal.SubscribeEventLocal(this_.HandleRoleLvUpEvent)
	eventlocal.SubscribeEventLocal(this_.HandleTitleChange)
	eventlocal.SubscribeEventLocal(this_.HandleRecentChatAdd)
	eventlocal.SubscribeEventLocal(this_.HandleExtraSkillAdd)
	eventlocal.SubscribeEventLocal(this_.HandleRoguelikeExtra)
	eventlocal.SubscribeEventLocal(this_.HandleUserRechargeChange)
	eventlocal.SubscribeEventLocal(this_.HandleNormalPaySuccess)

	h := this_.svc.Group(PayAuth)
	h.RegisterEvent("支付成功", this_.PaySuccess)

	this_.svc.RegisterFunc("充点扣点", this_.UpdateUserCurrency, handler.GMAuth)

	this_.BagService.RegisterUpdaterByType(ItemType.Avatar, func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg {
		cfg, ok := rule.MustGetReader(ctx).Item.GetItemById(itemId)
		if !ok {
			panic(errors.New("item_config_not_found: " + strconv.Itoa(int(itemId))))
		}
		err := this_.AddAvatar(ctx, itemId, cfg.TargetId, cfg.ExpiredTime)
		return err
	})

	return
}

func (this_ *Service) GetRole(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.Role, *errmsg.ErrMsg) {
	roles := make([]*dao.Role, len(roleIds))
	for idx := range roleIds {
		roles[idx] = &dao.Role{RoleId: roleIds[idx]}
	}
	newList, err := db.GetMultiRole(ctx, roles)
	if err != nil {
		return nil, err
	}

	rm := make(map[values.RoleId]*dao.Role, len(newList))
	for i := range newList {
		rm[newList[i].RoleId] = newList[i]
	}

	return rm, nil
}

func (this_ *Service) GetUserById(c *ctx.Context, userId string) (*dao.User, *errmsg.ErrMsg) {
	return db.GetUser(c, userId)
}

func (this_ *Service) GetUserByRoleIds(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.User, *errmsg.ErrMsg) {
	roles := make([]*dao.Role, len(roleIds))
	for idx := range roleIds {
		roles[idx] = &dao.Role{RoleId: roleIds[idx]}
	}
	newList, err := db.GetMultiRole(ctx, roles)
	if err != nil {
		return nil, err
	}
	users := make([]*dao.User, 0, len(newList))
	for _, roleData := range newList {
		users = append(users, &dao.User{UserId: roleData.UserId})
	}
	if err = db.GetMultiUser(ctx, users); err != nil {
		return nil, err
	}
	m := make(map[values.RoleId]*dao.User)
	for i := range users {
		m[users[i].RoleId] = users[i]
	}
	return m, nil
}

func (this_ *Service) SaveUser(c *ctx.Context, u *dao.User) {
	db.SaveUser(c, u)
}

func (this_ *Service) GetRoleByRoleId(ctx *ctx.Context, roleId values.RoleId) (*dao.Role, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, roleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r, nil
}

func (this_ *Service) GetRegisterDay(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, roleId)
	if err != nil {
		return 0, err
	}
	registerTime := time.Unix(r.CreateTime/1000, 0).UTC()
	registerDayBegin := this_.RefreshService.GetCurrDayFreshTimeWith(ctx, registerTime).Unix()
	todayBegin := this_.RefreshService.GetCurrDayFreshTime(ctx).Unix()
	if todayBegin <= registerDayBegin {
		return 1, nil
	}
	return (todayBegin-registerDayBegin)/86400 + 1, nil
}

func (this_ *Service) GetRoleModelByRoleId(ctx *ctx.Context, roleId values.RoleId) (*models.Role, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, roleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	return RoleDao2Model(r), nil
}

func (this_ *Service) GetRoleByUserId(ctx *ctx.Context, userId string) (*dao.Role, *errmsg.ErrMsg) {
	u, err := db.GetUser(ctx, userId)
	if err != nil {
		return nil, err
	}
	r, err := db.GetRole(ctx, u.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r, nil
}

func (this_ *Service) GetAvatar(ctx *ctx.Context) (values.Integer, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, ctx.RoleId)
	if err != nil {
		return 0, err
	}
	if r == nil {
		return 0, nil
	}
	return r.AvatarId, nil
}

func (this_ *Service) GetExtraSkillCnt(ctx *ctx.Context, typId, logicId values.Integer) (values.Integer, *errmsg.ErrMsg) {
	cnt, err := db.GetExtraSkill(ctx, ctx.RoleId)
	if err != nil {
		return 0, err
	}
	if data, exist := cnt.Data[typId]; !exist {
		return 0, nil
	} else {
		return data.Cnt[logicId], nil
	}
}

func (this_ *Service) GetAvatarFrame(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, roleId)
	if err != nil {
		return 0, err
	}
	if r == nil {
		return 0, nil
	}
	return r.AvatarFrame, nil
}

func (this_ *Service) GetTitle(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, roleId)
	if err != nil {
		return 0, err
	}
	if r == nil {
		return 0, nil
	}
	return r.Title, nil
}

func (this_ *Service) GetLevel(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, roleId)
	if err != nil {
		return 0, err
	}
	if r == nil {
		return 0, nil
	}
	return r.Level, nil
}

func (this_ *Service) GetPower(ctx *ctx.Context, roleId values.RoleId) (values.Integer, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, roleId)
	if err != nil {
		return 0, err
	}
	if r == nil {
		return 0, nil
	}
	return r.Power, nil
}

func (this_ *Service) GetMap(ctx *ctx.Context, userId string) (values.MapId, *errmsg.ErrMsg) {
	u, err := db.GetUser(ctx, userId)
	if err != nil {
		return 0, err
	}
	if u == nil {
		return 0, nil
	}
	return u.MapId, nil
}

func (this_ *Service) GetRoleAttr(ctx *ctx.Context, roleId values.RoleId) ([]*models.RoleAttr, *errmsg.ErrMsg) {
	r, err := db.GetRoleAttr(ctx, roleId)
	if err != nil {
		return nil, err
	}

	return RoleAttrDao2Models(r), nil
}

func (this_ *Service) GetRoleAttrByType(ctx *ctx.Context, roleId values.RoleId, typ models.AttrBonusType) (*models.RoleAttr, *errmsg.ErrMsg) {
	r, err := db.GetRoleAttrByType(ctx, roleId, typ)
	if err != nil {
		return nil, err
	}
	return RoleAttrDao2Model(r), nil
}

func (this_ *Service) SaveRole(ctx *ctx.Context, role *dao.Role) *errmsg.ErrMsg {
	db.SaveRole(ctx, role)
	return nil
}

// 防止玩家单次获得大量经验，返回为本次实际应该获得的经验
func (this_ *Service) realGainExp(ctx *ctx.Context, hangTimer *dao.HangExpTimer, cfgExp, have, gain values.Integer, isHang bool) (values.Integer, bool) {
	if !isHang {
		return gain, false
	}
	max := cfgExp
	// 经验已满，已经开始计时
	if hangTimer.Timing {
		sec := rule2.GetExpGetTimeLimit(ctx)
		// 已经超过配置时间，不再获得挂机经验，不更新hangTimer
		if timer.StartTime(ctx.StartTime).Unix()-hangTimer.Time >= sec {
			return 0, false
		}
		// 未超过配置时间，继续获得经验
		return gain, false
	}
	// 获得本次经验后经验已满，开始计时
	if have+gain >= max {
		hangTimer.Timing = true
		hangTimer.Time = timer.StartTime(ctx.StartTime).Unix()
		return gain, true
	}
	return gain, false

}

func (this_ *Service) GetRoguelikeCnt(ctx *ctx.Context) ([2]values.Integer, *errmsg.ErrMsg) {
	res := [2]values.Integer{}
	cnt, err := db.GetRLCnt(ctx, ctx.RoleId)
	if err != nil {
		return res, err
	}
	isChange := false
	n := this_.RefreshService.GetCurrDayFreshTime(ctx)
	if cnt.LastJoinAt < n.Unix() {
		cnt.TodayJoin = 0
		cnt.LastJoinAt = n.Unix()
		isChange = true
	}
	var oneDayStartLimit values.Integer = 5
	o, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("RougueDungeonsNum")
	if ok {
		oneDayStartLimit = o
	}
	res[0] = cnt.TodayJoin
	res[1] = cnt.ExtraCnt + oneDayStartLimit
	if isChange {
		db.SaveRLCnt(ctx, cnt)
	}
	return res, nil
}

func (this_ *Service) AddExp(ctx *ctx.Context, roleId values.RoleId, count values.Integer, isHang bool) *errmsg.ErrMsg {
	if count <= 0 {
		return nil
	}
	if err := ctx.DRLock(redisclient.GetLocker(), addExpLock+roleId); err != nil {
		return err
	}

	role, err := db.GetRole(ctx, roleId)
	if err != nil {
		return err
	}
	have, err := this_.BagService.GetItem(ctx, roleId, enum.RoleExp)
	if err != nil {
		return err
	}
	hangTimer, err := db.GetHangExpTimer(ctx)
	if err != nil {
		return err
	}
	reader := rule.MustGetReader(ctx)
	maxLv := reader.RoleLv.MaxRoleLevel()
	// 满级后继续获得经验
	if role.LevelIndex >= maxLv {
		cfg, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex)
		if !ok {
			return errmsg.NewInternalErr("role_lv not found")
		}
		val, update := this_.realGainExp(ctx, hangTimer, cfg.Exp, have, count, isHang)
		if val > 0 {
			if err := this_.BagService.AddItem(ctx, roleId, enum.RoleExp, val); err != nil {
				return err
			}
		} else if !hangTimer.Pushed {
			ctx.PushMessage(&lessservicepb.User_UserCanGetExpFromHangPush{Can: false})
			update = true
			hangTimer.Pushed = true
		}
		if update {
			db.SaveHangExpTimer(ctx, hangTimer)
		}
		return nil
	}

	cfg, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex)
	if !ok {
		return errmsg.NewInternalErr("role_lv not found")
	}
	cfgNext, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex + 1)
	if !ok {
		return errmsg.NewInternalErr("role_lv not found")
	}
	// cfgExp := cfg.Exp
	// // 突破这一级经验配置为0，需要取上一级的经验来计算上限
	// if cfgExp == 0 {
	//	cfgLast, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex - 1)
	//	if !ok {
	//		return errmsg.NewInternalErr("role_lv not found")
	//	}
	//	cfgExp = cfgLast.Exp
	// }

	// 需要玩家主动点升级，直接将经验加至玩家背包
	if cfg.LvHero != cfgNext.LvHero {
		exp, update := this_.realGainExp(ctx, hangTimer, cfg.Exp, have, count, isHang)
		if exp > 0 {
			if err := this_.BagService.AddItem(ctx, roleId, enum.RoleExp, exp); err != nil {
				return err
			}
		} else if !hangTimer.Pushed {
			ctx.PushMessage(&lessservicepb.User_UserCanGetExpFromHangPush{Can: false})
			update = true
			hangTimer.Pushed = true
		}
		if update {
			db.SaveHangExpTimer(ctx, hangTimer)
		}
		return nil
	}
	// 处理遗物 升级所需经验值减少{0} 效果
	// 这里返回的discount是万分比
	discount, err := this_.GetExtraSkillCnt(ctx, 4, 1)
	if err != nil {
		return err
	}
	total := have + count // 当前拥有的经验
	lastIndex := role.LevelIndex
	total, _, err = this_.autoUpgrade(ctx, role, total, discount)
	if err != nil {
		return err
	}
	// 处理升级后剩余经验
	if total > have {
		if err := this_.BagService.AddItem(ctx, roleId, enum.RoleExp, total-have); err != nil {
			return err
		}
	} else if total < have {
		if err := this_.BagService.SubItem(ctx, roleId, enum.RoleExp, have-total); err != nil {
			return err
		}
	}
	if lastIndex != role.LevelIndex {
		db.SaveRole(ctx, role)
		ctx.PublishEventLocal(&event.UserLevelChange{
			Level:          role.Level,
			Incr:           0,
			IsAdvance:      false,
			LevelIndex:     role.LevelIndex,
			LevelIndexIncr: role.LevelIndex - lastIndex,
		})
	}
	// 等级没有变化不需要同步至mysql
	// utils.Must(syncrole.Update(ctx, role))
	return nil
}

func genUserId(userId string, serverId values.ServerId) string {
	d := make([]byte, 0, len(userId)+20)
	d = strconv.AppendInt(d, serverId, 10)
	d = append(d, ':')
	d = append(d, userId...)
	return string(d)
}

func genRoleId(roleId string) string {
	d := make([]byte, 0, len(roleId)+20)
	d = append(d, "role:"...)
	d = append(d, roleId...)
	return string(d)
}

func (this_ *Service) ChangeBattleId(c *ctx.Context, request *lessservicepb.User_ChangeBattleMapEvent) {
	u, err := db.GetUser(c, c.UserId)
	if err != nil {
		c.Warn("ChangeBattleId error", zap.Error(err), zap.Any("req", request))
		return
	}
	if u == nil {
		return
	}
	u.MapId = request.MapId
	u.BattleServerId = request.BattleServerId
	c.PublishEventLocal(&event.BattleMapChange{MapId: request.MapId})
	db.SaveUser(c, u)
}

func (this_ *Service) GetRoleRequest(c *ctx.Context, req *lessservicepb.User_GetRoleRequest) (*lessservicepb.User_GetRoleResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(c, req.RoleId)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	heroes, err := this_.GetAllHero(c, req.RoleId)
	if err != nil {
		return nil, err
	}
	assemble, err := this_.FormationService.GetDefaultHeroes(c, req.RoleId)
	if err != nil {
		return nil, err
	}
	formation := make([]values.Integer, 0)
	if assemble.Hero_0 > 0 {
		formation = append(formation, assemble.Hero_0)
	}
	if assemble.Hero_1 > 0 {
		formation = append(formation, assemble.Hero_1)
	}
	return &lessservicepb.User_GetRoleResponse{
		Role:      RoleDao2Model(r),
		Heroes:    heroes,
		Formation: formation,
	}, nil
}

func (this_ *Service) SearchRoleRequest(c *ctx.Context, req *lessservicepb.User_SearchRoleRequest) (*lessservicepb.User_SearchRoleResponse, *errmsg.ErrMsg) {
	i, err := strconv.Atoi(req.Input)
	if err != nil {
		return nil, nil
	}
	roleId := utils.Base34EncodeToString(uint64(i))
	roles := make([]*dao.Role, 0)
	r, err1 := db.GetRole(c, roleId)
	if err1 != nil {
		return nil, err1
	}
	if r != nil {
		roles = append(roles, r)
	}
	return &lessservicepb.User_SearchRoleResponse{Role: RoleDao2Models(roles)}, nil
}

func verifyLanguage(c *ctx.Context, language int64) int64 {
	_, ok := rule.MustGetReader(c).VerifyLanguage.GetVerifyLanguageById(language)
	if ok {
		return language
	}
	return enum.DefaultLanguage // EN
}

func (this_ *Service) verifyClientVersion(version string) *errmsg.ErrMsg {
	versionData, err := db.GetClientVersion()
	if err != nil {
		return err
	}
	minVersion := versionData.MinVersion
	if minVersion == "" {
		minVersion = versionData.Version
	}
	mainVersionList := strings.Split(versionData.Version, ".")
	minVersionList := strings.Split(minVersion, ".")
	clientVersionList := strings.Split(version, ".")
	if len(mainVersionList) != 3 || len(minVersionList) != 3 || len(clientVersionList) != 3 {
		return errmsg.NewErrClientVersionNotMatch()
	}
	mainVersion1, err1 := strconv.Atoi(mainVersionList[0])
	if err1 != nil {
		return errmsg.NewErrClientVersionNotMatch()
	}
	mainVersion2, err1 := strconv.Atoi(mainVersionList[1])
	if err1 != nil {
		return errmsg.NewErrClientVersionNotMatch()
	}
	minVersion1, err1 := strconv.Atoi(minVersionList[0])
	if err1 != nil {
		return errmsg.NewErrClientVersionNotMatch()
	}
	minVersion2, err1 := strconv.Atoi(minVersionList[1])
	if err1 != nil {
		return errmsg.NewErrClientVersionNotMatch()
	}
	clientVersion1, err1 := strconv.Atoi(clientVersionList[0])
	if err1 != nil {
		return errmsg.NewErrClientVersionNotMatch()
	}
	clientVersion2, err1 := strconv.Atoi(clientVersionList[1])
	if err1 != nil {
		return errmsg.NewErrClientVersionNotMatch()
	}
	max := mainVersion1*10000 + mainVersion2
	min := minVersion1*10000 + minVersion2
	cur := clientVersion1*10000 + clientVersion2
	if cur > max || cur < min {
		return errmsg.NewErrClientVersionNotMatch()
	}
	return nil
}
func (this_ *Service) KickOffUser(c *ctx.Context, req *servicepb.User_KickOffUserRequest) (*servicepb.User_KickOffUserResponse, *errmsg.ErrMsg) {
	if c.RoleId == "" {
		return nil, errmsg.NewErrUserNotFound()
	}
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	u, err := db.GetUser(c, r.UserId)
	if err != nil {
		return nil, err
	}
	freezeTime := req.KickoffSeconds
	// if req.Status == 1 {
	//	if freezeTime <= 0 {
	//		freezeTime = 5 * 60
	//	}
	// }
	u.FreezeTime = timer.Now().Add(time.Second * time.Duration(freezeTime)).Unix()
	db.SaveUser(c, u)
	err = this_.svc.GetNatsClient().Publish(0, c.ServerHeader, &broadcast.GatewayStdTcp_KickOffUserPush{RoleId: r.RoleId, Status: req.Status})
	if err != nil {
		return nil, err
	}
	return &servicepb.User_KickOffUserResponse{}, nil
}

func (this_ *Service) Login(c *ctx.Context, request *lessservicepb.User_RoleLoginRequest) (*lessservicepb.User_RoleLoginResponse, *errmsg.ErrMsg) {
	if request.UserId == "" && request.IggId == "" {
		return nil, errmsg.NewErrInvalidRequestParam()
	}
	var userId string
	if request.IggId != "" {
		userId = request.IggId
		this_.log.Info("login with iggsdk", zap.String("igg_id", request.IggId))
	} else {
		userId = request.UserId
	}
	if err := this_.verifyClientVersion(request.ClientVersion); err != nil {
		return nil, err
	}
	if err := lock(c, userId); err != nil {
		return nil, err
	}
	// TODO：创建用的request的serverId
	u, err := db.GetUser(c, userId)
	if err != nil {
		return nil, err
	}
	var r *dao.Role
	register := false
	heroes := make([]*models.Hero, 0)
	gameId, _ := strconv.ParseInt(request.GameId, 10, 64)
	sessionId := this_.md5Base32(xid.New().String())
	if u == nil {
		// 统计注册人数
		db.RegisterCountIncr(c)

		var roleIdInt int64
		roleIdInt, err = idgenerate.GenerateID(context.Background(), idgenerate.RoleIDKey)
		if err != nil {
			return nil, err
		}
		roleId := utils.Base34EncodeToString(uint64(roleIdInt))
		u = &dao.User{
			UserId:      userId,
			RoleId:      roleId,
			DeviceId:    request.DeviceId,
			ServerId:    c.InServerId, // TODO：this_.serverId  => request.ServerId 不然查不到
			HangupMapId: mapdata.GetDefaultMapId(c),
		}
		// if !sensitive.TextValid(request.UserId) {
		//	return nil, errmsg.NewErrSensitive()
		// }
		originHead := rule2.MustGetInitialUseOfAvatar(c)
		r = &dao.Role{
			RoleId:      roleId,
			Nickname:    trans.GenPlayerName(roleIdInt),
			Level:       enum.INIT_LEVEL,
			AvatarId:    originHead[0],
			AvatarFrame: originHead[1],
			Power:       1000,
			Title:       1,
			Language:    verifyLanguage(c, request.Language),
			Login:       time.Unix(0, c.StartTime).UnixMilli(),
			Logout:      0,
			ChangeName:  0,
			CreateTime:  time.Unix(0, c.StartTime).UnixMilli(),
			// ExpProfit:   2, // 已移到 RoleTempBag
			LevelIndex:  enum.INIT_LEVEL,
			BattleSpeed: 1, // 初始为1
			UserId:      u.UserId,
			GameId:      gameId,
			SessionId:   sessionId,
		}
		c.PublishEventLocal(&event.GainTalentCommonPoint{
			Num: rule.MustGetReader(c).OriginTalentPoint(),
		})
		this_.InitBagConfig(c)

		register = true
		db.SaveUser(c, u)
		db.SaveRole(c, r)
		err = this_.saveRoleId(c, roleIdInt, roleId)
		if err != nil {
			return nil, err
		}
		c.RoleId = u.RoleId
		err = this_.titleUnlock(c)
		if err != nil {
			return nil, err
		}

		heroes, err = this_.getDefaultResource(c)
		if err != nil {
			return nil, err
		}
		msg := &recommend.Recommend_UserEnterEvent{Role: RoleDao2Model(r)}
		err = this_.svc.GetNatsClient().Publish(0, c.ServerHeader, msg)
		if err != nil {
			c.Warn(msg.XXX_MessageName())
		}
	} else {
		if u.FreezeTime > timer.Unix() { // 账号被冻结，不可以登陆
			return &lessservicepb.User_RoleLoginResponse{
				Status: models.Status_FREEZE,
				RoleId: u.RoleId,
			}, nil
		}

		r, err = db.GetRole(c, u.RoleId)
		if err != nil {
			return nil, err
		}
		r.Login = time.Unix(0, c.StartTime).UnixMilli()
		if r.UserId == "" {
			r.UserId = u.UserId
		}
		r.GameId = gameId
		r.SessionId = sessionId
		db.SaveRole(c, r)
		if request.DeviceId != u.DeviceId {
			u.DeviceId = request.DeviceId
			db.SaveUser(c, u)
		}
	}
	status := models.Status_SUCCESS

	c.RoleId = u.RoleId
	c.UserId = u.UserId
	c.RuleVersion = request.RuleVersion
	c.BattleMapId = u.MapId
	guildId, err := this_.GuildService.GetGuildIdByRole(c)
	if err != nil {
		return nil, err
	}
	position, err := this_.GuildService.GetUserGuildPositionByGuildId(c, guildId)
	if err != nil {
		return nil, err
	}
	if register {
		utils.Must(im.DefaultClient.JoinRoom(c, &im.RoomRole{
			RoomID:  strconv.Itoa(int(r.Language)),
			RoleIDs: []string{c.RoleId},
		}))
		statistical.Save(c.NewLogServer(), &models2.Register{
			IggId:       iggsdk.ConvertToIGGId(c.UserId),
			EventTime:   timer.Now(),
			GwId:        statistical.GwId(),
			RoleId:      u.RoleId,
			UserId:      u.UserId,
			DeviceId:    u.DeviceId,
			IP:          c.Ip,
			RuleVersion: request.RuleVersion,
		})
		statistical2.Save(c.NewLogServer2(), &models3.Register{
			Time:        timer.Now(),
			IggId:       c.UserId,
			ServerId:    c.ServerId,
			Xid:         xid.New().String(),
			RoleId:      u.RoleId,
			UserId:      u.UserId,
			DeviceId:    u.DeviceId,
			IP:          c.Ip,
			RuleVersion: request.RuleVersion,
		})
		statistical.Save(c.NewLogServer(), &models2.PlayerLevel{
			IggId:     iggsdk.ConvertToIGGId(c.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    u.RoleId,
			Level:     enum.INIT_LEVEL,
		})
		statistical2.Save(c.NewLogServer2(), &models3.PlayerLevel{
			Time:     timer.Now(),
			IggId:    c.UserId,
			ServerId: c.ServerId,
			Xid:      xid.New().String(),
			RoleId:   c.RoleId,
			Level:    enum.INIT_LEVEL,
		})
		utils.Must(syncrole.Create(c, r)) // 同步至mysql
		this_.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskLevel, 0, 1, true)
	} else {
		heroes, err = this_.HeroService.GetAllHero(c, c.RoleId)
		if err != nil {
			return nil, err
		}
		if err := this_.FashionExpiredCheck(c, r); err != nil {
			return nil, err
		}
	}
	statistical.Save(c.NewLogServer(), &models2.Login{
		IggId:         iggsdk.ConvertToIGGId(c.UserId),
		EventTime:     timer.Now(),
		GwId:          statistical.GwId(),
		RoleId:        u.RoleId,
		UserId:        u.UserId,
		DeviceId:      u.DeviceId,
		IP:            c.Ip,
		RuleVersion:   request.RuleVersion,
		ClientVersion: request.ClientVersion,
	})
	statistical2.Save(c.NewLogServer2(), &models3.Login{
		Time:          timer.Now(),
		IggId:         c.UserId,
		ServerId:      c.ServerId,
		Xid:           sessionId,
		RoleId:        u.RoleId,
		UserId:        u.UserId,
		DeviceId:      u.DeviceId,
		IP:            c.Ip,
		RuleVersion:   request.RuleVersion,
		GameId:        strconv.FormatInt(r.GameId, 10),
		ClientVersion: request.ClientVersion,
	})
	roleStateNotify := &writepb.RoleStateRW_LoginNotifyEvent{}
	err = this_.svc.GetNatsClient().Publish(0, c.ServerHeader, roleStateNotify)
	if err != nil {
		c.Warn(roleStateNotify.XXX_MessageName())
	}
	if err := this_.HangExpTimerPush(c); err != nil {
		return nil, err
	}
	// 登录打点
	this_.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskLogin, 0, 1)
	c.PublishEventLocal(&event.Login{IsRegister: register, ServerId: u.ServerId, UserId: request.UserId, RuleVersion: request.RuleVersion, RoleId: u.RoleId})
	return &lessservicepb.User_RoleLoginResponse{
		Status:        status,
		RoleId:        u.RoleId,
		MapId:         u.MapId,
		Level:         r.Level,
		AvatarId:      r.AvatarId,
		AvatarFrame:   r.AvatarFrame,
		Power:         r.Power,
		Title:         r.Title,
		GuildId:       guildId,
		GuildPosition: position,
		Nickname:      r.Nickname,
		Language:      r.Language,
		Heroes:        heroes,
		CreateTime:    r.CreateTime,
		LevelIndex:    r.LevelIndex,
		ImHttp:        im.DefaultClient.Config().ImHttp,
		ImTcp:         im.DefaultClient.Config().ImTcp,
		BattleSpeed:   r.BattleSpeed,
		ServerId:      u.ServerId,
		Recharge:      r.Recharge,
	}, nil
}

func (this_ *Service) GetExtraSkillCntPb(ctx *ctx.Context, req *servicepb.User_GetExtraSkillCntRequest) (*servicepb.User_GetExtraSkillCntResponse, *errmsg.ErrMsg) {
	cnt, err := db.GetExtraSkill(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if _, exist := cnt.Data[req.TypId]; !exist {
		return nil, nil
	}
	return &servicepb.User_GetExtraSkillCntResponse{
		Cnt: cnt.Data[req.TypId].Cnt,
	}, nil
}

func (this_ *Service) getDefaultResource(ctx *ctx.Context) ([]*models.Hero, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(ctx)
	list := reader.Begining.List()
	var (
		itemMap = map[values.ItemId]values.Integer{}
		heroes  = make([]*models.Hero, 0)
	)

	for idx := range list {
		switch list[idx].Typ {
		case 1:
			itemMap[list[idx].Rewardid] = list[idx].Count
		case 2:
			hero, err := this_.HeroService.AddHero(ctx, list[idx].Rewardid, true)
			if err != nil {
				return nil, err
			}
			heroes = append(heroes, hero)
		}
	}

	_, err := this_.BagService.AddManyItem(ctx, ctx.RoleId, itemMap)
	if err != nil {
		return nil, err
	}
	return heroes, nil
}

func (this_ *Service) Logout(c *ctx.Context, req *lessservicepb.User_RoleLogoutPush) {
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return
	}
	if r == nil {
		c.Error("notify game server logout error", zap.String("info", "role not found"))
		return
	}
	r.Logout = time.Unix(0, c.StartTime).UnixMilli()
	onlineSec := (r.Logout - r.Login) / 1000
	db.SaveRole(c, r)

	if err := syncrole.Update(c, r); err != nil {
		c.Error("syncrole.Update err", zap.Error(err), zap.Any("role", r))
	}

	c.PublishEventLocal(event.Logout{})
	roleStateNotify := &writepb.RoleStateRW_LogoutNotifyEvent{}
	err = this_.svc.GetNatsClient().Publish(0, c.ServerHeader, roleStateNotify)
	if err != nil {
		c.Warn(roleStateNotify.XXX_MessageName())
	}
	u, err := db.GetUser(c, c.UserId)
	if u != nil && err == nil {
		statistical.Save(c.NewLogServer(), &models2.Logout{
			IggId:         iggsdk.ConvertToIGGId(c.UserId),
			EventTime:     timer.Now(),
			GwId:          statistical.GwId(),
			RoleId:        u.RoleId,
			UserId:        u.UserId,
			DeviceId:      u.DeviceId,
			IP:            c.Ip,
			RuleVersion:   "",
			OnlineSeconds: onlineSec,
			ClientVersion: req.ClientVersion,
		})
		statistical2.Save(c.NewLogServer2(), &models3.Logout{
			Time:          timer.Now(),
			IggId:         c.UserId,
			ServerId:      c.ServerId,
			Xid:           r.SessionId,
			RoleId:        u.RoleId,
			UserId:        u.UserId,
			DeviceId:      u.DeviceId,
			IP:            c.Ip,
			RuleVersion:   "",
			OnlineSeconds: onlineSec,
			GameId:        strconv.FormatInt(r.GameId, 10),
			ClientVersion: req.ClientVersion,
		})
	}

}

func (this_ *Service) SetCutSceneId(c *ctx.Context, req *servicepb.User_SetCutSceneIdRequest) (*servicepb.User_SetCutSceneIdResponse, *errmsg.ErrMsg) {
	d, err := db.GetCutScene(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	d.CutId = req.CutId
	d.Data = req.Data
	db.SaveCutScene(c, d)
	return &servicepb.User_SetCutSceneIdResponse{}, nil
}

func (this_ *Service) GetCutSceneId(c *ctx.Context, _ *servicepb.User_GetCutSceneIdRequest) (*servicepb.User_GetCutSceneIdResponse, *errmsg.ErrMsg) {
	d, err := db.GetCutScene(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	return &servicepb.User_GetCutSceneIdResponse{CutId: d.CutId, Data: d.Data}, nil
}

func (this_ *Service) ChangeNickName(c *ctx.Context, request *lessservicepb.User_ChangeNicknameRequest) (*lessservicepb.User_ChangeNicknameResponse, *errmsg.ErrMsg) {
	if !sensitive.TextValid(request.Nickname) {
		return nil, errmsg.NewErrSensitive()
	}
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	doCostFlag := false
	costList := rule.MustGetReader(c).RenameCost()
	for _, cost := range costList {
		err := this_.Module.BagService.SubItem(c, c.RoleId, cost[0], cost[1])
		if err == nil {
			doCostFlag = true
			break
		}
	}
	if !doCostFlag {
		return nil, errmsg.NewErrBagNotEnough()
	}
	friendIds, err := this_.FriendService.GetFriendIds(c)
	if err != nil {
		return nil, err
	}
	notice := fmt.Sprintf(`{"old_name":"%s", "new_name":"%s"}`, r.Nickname, request.Nickname)
	this_.noticeChangeSysMsg(c, friendIds, notice)
	r.Nickname = request.Nickname
	r.ChangeName++
	err = this_.BattleService.NicknameChange(c, r.Nickname)
	if err != nil {
		return nil, err
	}

	db.SaveRole(c, r)
	utils.Must(syncrole.Update(c, r))
	// this_.svc.TickFuncCtx(c, time.Second*2, func(ctx *ctx.Context) bool {
	//	ctx.TraceLogger.Info("after func", zap.Any("header", ctx.ServerHeader))
	//	return true
	// })
	// c.PushMessageToRole("13H2Y", &servicepb.User_LogoutPush{RoleId: c.RoleId})
	// c.PushMessageToRole("13HYC", &servicepb.User_LogoutPush{RoleId: c.RoleId})
	return &lessservicepb.User_ChangeNicknameResponse{Nickname: r.Nickname}, nil
}

func (this_ *Service) noticeChangeSysMsg(ctx *ctx.Context, roleIds []values.RoleId, notice string) {
	gopool.Submit(func() {
		for _, roleId := range roleIds {
			_ = im.DefaultClient.SendMessage(ctx, &im.Message{
				Type:      im.MsgTypePrivate,
				RoleID:    "system",
				RoleName:  "system",
				TargetID:  roleId,
				Content:   notice,
				ParseType: im.ParseChangeNickName,
			})
		}
	})
}

func (this_ *Service) GetBattleSettingData(c *ctx.Context, _ *servicepb.User_GetBattleSettingDataRequest) (*servicepb.User_GetBattleSettingDataResponse, *errmsg.ErrMsg) {
	d, err := db.GetBattleSetting(c)
	if err != nil {
		return nil, err
	}
	return &servicepb.User_GetBattleSettingDataResponse{Data: d.Data}, nil
}

func (this_ *Service) UpdateBattleSettingData(c *ctx.Context, req *servicepb.User_UpdateBattleSettingDataRequest) (*servicepb.User_UpdateBattleSettingDataResponse, *errmsg.ErrMsg) {
	d, err := db.GetBattleSetting(c)
	if err != nil {
		return nil, err
	}

	// 吃药设置是否超过上限
	hpMax, ok1 := rule.MustGetReader(c).KeyValue.GetInt64("TakeMedicineHPPercentage")
	mpMax, ok2 := rule.MustGetReader(c).KeyValue.GetInt64("TakeMedicineMPPercentage")
	if !ok1 || !ok2 {
		return nil, errmsg.NewInternalErr("TakeMedicineHPPercentage and TakeMedicineMPPercentage not found")
	}
	if req.Data.Hp > hpMax {
		return nil, errmsg.NewErrMedicineHpMax()
	}
	if req.Data.Mp > mpMax {
		return nil, errmsg.NewErrMedicineHpMax()
	}

	d.Data = req.Data
	db.SaveBattleSetting(c, d)
	c.PublishEventLocal(&event.BattleSettingChange{Setting: d.Data})

	battleServerId, err1 := this_.Module.GetCurBattleSrvId(c)
	if err1 != nil {
		return nil, err1
	}
	if battleServerId > 0 {
		err3 := this_.svc.GetNatsClient().Publish(battleServerId, c.ServerHeader, &cppbattle.CPPBattle_UserAutoSoulSkillPush{
			ObjId:          c.RoleId,
			AutoSouleSkill: req.Data.AutoSoulSkill,
		})
		if err3 != nil {
			return nil, err3
		}
	}

	return &servicepb.User_UpdateBattleSettingDataResponse{}, nil
}

func (this_ *Service) ChangeLanguage(c *ctx.Context, request *lessservicepb.User_ChangeLanguageRequest) (*lessservicepb.User_ChangeLanguageResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	utils.Must(im.DefaultClient.LeaveRoom(c, &im.RoomRole{
		RoomID:  strconv.Itoa(int(r.Language)),
		RoleIDs: []string{c.RoleId},
	}))
	r.Language = request.Language
	utils.Must(im.DefaultClient.JoinRoom(c, &im.RoomRole{
		RoomID:  strconv.Itoa(int(r.Language)),
		RoleIDs: []string{c.RoleId},
	}))
	db.SaveRole(c, r)

	// this_.svc.TickFuncCtx(c, time.Second*2, func(ctx *ctx.Context) bool {
	//	ctx.TraceLogger.Info("after func", zap.Any("header", ctx.ServerHeader))
	//	return true
	// })
	// c.PushMessageToRole("13H2Y", &servicepb.User_LogoutPush{RoleId: c.RoleId})
	// c.PushMessageToRole("13HYC", &servicepb.User_LogoutPush{RoleId: c.RoleId})
	return &lessservicepb.User_ChangeLanguageResponse{Language: r.Language}, nil
}

func (this_ *Service) GetRecentChatIds(c *ctx.Context, req *lessservicepb.User_GetRecentChatIdsRequest) (*lessservicepb.User_GetRecentChatIdsResponse, *errmsg.ErrMsg) {
	data, err := db.GetRecentChat(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	res, idx := make([]values.RoleId, len(data.TarRoleIds)), 0
	for roleId := range data.TarRoleIds {
		res[idx] = roleId
		idx++
	}
	sort.Slice(res, func(i, j int) bool {
		return data.TarRoleIds[res[i]] > data.TarRoleIds[res[j]]
	})
	return &lessservicepb.User_GetRecentChatIdsResponse{
		RoleIds: res,
	}, nil
}

func (this_ *Service) AddRecentChatIds(c *ctx.Context, req *lessservicepb.User_AddRecentChatIdsRequest) (*lessservicepb.User_AddRecentChatIdsResponse, *errmsg.ErrMsg) {
	if err := this_.addRecentChatIds(c, c.RoleId, req.RoleId); err != nil {
		return nil, err
	}
	return &lessservicepb.User_AddRecentChatIdsResponse{}, nil
}

func (this_ *Service) addRecentChatIds(c *ctx.Context, myRoleId, tarRoleId values.RoleId) *errmsg.ErrMsg {
	err := c.DRLock(redisclient.GetLocker(), getLockKey(myRoleId), getLockKey(tarRoleId))
	if err != nil {
		return err
	}
	data, err := db.GetRecentChat(c, myRoleId)
	if err != nil {
		return err
	}
	tarData, err := db.GetRecentChat(c, tarRoleId)
	if err != nil {
		return err
	}
	now := timer.Unix()
	data.TarRoleIds[tarRoleId] = now
	tarData.TarRoleIds[myRoleId] = now
	db.SaveRecentChat(c, data)
	db.SaveRecentChat(c, tarData)
	return nil
}

func (this_ *Service) DeleteRecentChatIds(c *ctx.Context, req *lessservicepb.User_DeleteRecentChatIdsRequest) (*lessservicepb.User_DeleteRecentChatIdsResponse, *errmsg.ErrMsg) {
	data, err := db.GetRecentChat(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	delete(data.TarRoleIds, req.RoleId)
	db.SaveRecentChat(c, data)
	return &lessservicepb.User_DeleteRecentChatIdsResponse{}, nil
}

func getLockKey(roleId values.RoleId) string {
	return "server:recent_chat:" + roleId
}

func (this_ *Service) LevelUpgrade(c *ctx.Context, _ *lessservicepb.User_LevelUpgradeRequest) (*lessservicepb.User_LevelUpgradeResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(c)
	maxLv := reader.RoleLv.MaxRoleLevel()
	if r.LevelIndex >= maxLv {
		return nil, errmsg.NewErrMaxLevel()
	}
	hangExpTimer, err := db.GetHangExpTimer(c)
	if err != nil {
		return nil, err
	}
	cfg, ok := reader.RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	lastIndex := r.LevelIndex
	r.LevelIndex++
	cfgNext, ok := reader.RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}

	var isAdvance bool
	lastTitle := r.Title
	if cfg.AdvancedHero != cfgNext.AdvancedHero {
		isAdvance = true
		dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
		if err != nil {
			return nil, err
		}
		if dungeon == nil || dungeon.ClearedAt == 0 {
			return nil, errmsg.NewErrNeedBeatBoss()
		}
		r.Title++
	}
	totalExp, err := this_.BagService.GetItem(c, c.RoleId, enum.RoleExp)
	if err != nil {
		return nil, err
	}
	// 处理遗物 升级所需经验值减少{0} 效果
	// 这里返回的discount是万分比
	discount, err := this_.GetExtraSkillCnt(c, 4, 1)
	if err != nil {
		return nil, err
	}
	need := cfg.Exp - values.Integer(math.Floor(values.Float(cfg.Exp*discount)/10000.0))
	// if err := this_.BagService.SubItem(c, c.RoleId, enum.RoleExp, need); err != nil {
	//	return nil, errmsg.NewErrExpNotEnough()
	// }
	if totalExp < need {
		c.Debug("LevelUpgrade",
			zap.String("role_id", c.RoleId),
			zap.Int64("need", need),
			zap.Int64("has", totalExp),
			zap.Int64("level_index", r.LevelIndex),
		)
		return nil, errmsg.NewErrExpNotEnough()
	}
	deductExp := need
	totalExp -= need

	// 突破，在挑战boss的时候已经扣除了材料，所以这里不需要再扣除材料
	if !isAdvance {
		items := make(map[values.ItemId]values.Integer)
		var hasItem bool
		for id, count := range cfg.AdvancedItem {
			items[id] = +count
			if id > 0 && count > 0 {
				hasItem = true
			}
		}
		if hasItem {
			if err := this_.BagService.SubManyItem(c, c.RoleId, items); err != nil {
				return nil, errmsg.NewErrMaterialNotEnough()
			}
		}
	}
	last := r.Level
	r.Level = cfgNext.LvHero
	// 可能存在经验溢出的情况，玩家手动升级后，剩余的经验足够自动升级，处理自动升级流程
	_, deduct, err := this_.autoUpgrade(c, r, totalExp, discount)
	if err != nil {
		return nil, err
	}
	deductExp += deduct
	if err := this_.BagService.SubItem(c, c.RoleId, enum.RoleExp, deductExp); err != nil {
		return nil, err
	}

	hangExpTimer.Timing = false
	hangExpTimer.Pushed = false
	db.SaveHangExpTimer(c, hangExpTimer)
	db.SaveRole(c, r)
	utils.Must(syncrole.Update(c, r)) // 同步至mysql
	this_.TaskService.UpdateTarget(c, c.RoleId, models.TaskType_TaskLevel, 0, r.Level, true)
	c.PublishEventLocal(&event.UserLevelChange{
		Level:          r.Level,
		Incr:           r.Level - last,
		IsAdvance:      isAdvance,
		LevelIndex:     r.LevelIndex,
		LevelIndexIncr: r.LevelIndex - lastIndex,
	})
	if last != r.Level {
		statistical.Save(c.NewLogServer(), &models2.PlayerLevel{
			IggId:     iggsdk.ConvertToIGGId(c.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    r.RoleId,
			Level:     r.Level,
		})
		statistical2.Save(c.NewLogServer2(), &models3.PlayerLevel{
			Time:     time.Now(),
			IggId:    c.UserId,
			ServerId: c.ServerId,
			Xid:      xid.New().String(),
			RoleId:   c.RoleId,
			Level:    r.Level,
		})
	}
	if lastTitle != r.Title {
		c.PublishEventLocal(&event.UserTitleChange{
			RoleId:       c.RoleId,
			LastTitle:    lastTitle,
			CurrentTitle: r.Title,
			CombatValue:  r.Power,
		})
		statistical.Save(c.NewLogServer(), &models2.PlayerTitle{
			IggId:     iggsdk.ConvertToIGGId(c.UserId),
			EventTime: timer.Now(),
			GwId:      statistical.GwId(),
			RoleId:    r.RoleId,
			Title:     r.Title,
		})
	}
	c.PushMessage(&lessservicepb.User_UserCanGetExpFromHangPush{Can: true})
	return &lessservicepb.User_LevelUpgradeResponse{
		Level:      r.Level,
		LevelIndex: r.LevelIndex,
	}, nil
}

func (this_ *Service) autoUpgrade(
	ctx *ctx.Context,
	role *dao.Role,
	totalExp, discount values.Integer, // discount为遗物升级所需经验值减少{0} 效果（万分比）
) (values.Integer, values.Integer, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(ctx)
	cfg, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex)
	if !ok {
		return 0, 0, errmsg.NewInternalErr("role_lv not found")
	}
	maxLv := reader.RoleLv.MaxRoleLevel()
	need := cfg.Exp - values.Integer(math.Floor(values.Float(cfg.Exp*discount)/10000.0))
	cfgNext, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex + 1)
	if !ok {
		return 0, 0, errmsg.NewInternalErr("role_lv not found")
	}
	var deduct values.Integer
	for totalExp >= need && cfg.LvHero == cfgNext.LvHero {
		role.LevelIndex++
		totalExp -= cfg.Exp
		deduct += cfg.Exp
		cfg, ok = reader.RoleLv.GetRoleLvById(role.LevelIndex)
		if !ok {
			return 0, 0, errmsg.NewInternalErr("role_lv not found")
		}
		// 满级
		if role.LevelIndex >= maxLv {
			break
		}
		cfgNext, ok = reader.RoleLv.GetRoleLvById(role.LevelIndex + 1)
		if !ok {
			return 0, 0, errmsg.NewInternalErr("role_lv not found")
		}
		// 需要手动升级
		if cfg.LvHero != cfgNext.LvHero {
			break
		}
		need = cfg.Exp - values.Integer(math.Floor(values.Float(cfg.Exp*discount)/10000.0))
	}
	// 返回：剩余经验和升级消耗经验
	return totalExp, deduct, nil
}

func (this_ *Service) LevelUpgradeMany(c *ctx.Context, req *lessservicepb.User_LevelUpgradeManyRequest) (*lessservicepb.User_LevelUpgradeManyResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(c)
	maxLv := reader.RoleLv.MaxRoleLevel()
	if r.LevelIndex >= maxLv {
		return nil, errmsg.NewErrMaxLevel()
	}
	targetLv := r.LevelIndex + req.Count
	if targetLv > maxLv {
		targetLv = maxLv
	}
	material := make(map[values.ItemId]values.Integer)
	var (
		exp     values.Integer
		hasItem bool
	)

	for i := r.LevelIndex; i < targetLv; i++ {
		cfg, ok := reader.RoleLv.GetRoleLvById(i)
		if !ok {
			return nil, errmsg.NewInternalErr("role_lv not found")
		}
		if cfg.CopyId > 0 {
			return nil, errmsg.NewErrNeedBeatBoss()
		}
		exp += cfg.Exp
		for id, v := range cfg.AdvancedItem {
			if id > 0 && v > 0 {
				hasItem = true
				material[id] += v
			}
		}
	}
	if err := this_.BagService.SubItem(c, c.RoleId, enum.RoleExp, exp); err != nil {
		return nil, errmsg.NewErrExpNotEnough()
	}
	if hasItem {
		if err := this_.BagService.SubManyItem(c, c.RoleId, material); err != nil {
			return nil, errmsg.NewErrMaterialNotEnough()
		}
	}
	last := r.Level
	r.LevelIndex += req.Count
	cfg, ok := reader.RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	r.Level = cfg.LvHero
	db.SaveRole(c, r)

	utils.Must(syncrole.Update(c, r)) // 同步至mysql

	c.PublishEventLocal(&event.UserLevelChange{
		Level:      r.Level,
		Incr:       r.Level - last,
		IsAdvance:  false,
		LevelIndex: r.LevelIndex,
	})
	return &lessservicepb.User_LevelUpgradeManyResponse{
		Level:      r.Level,
		LevelIndex: r.LevelIndex,
	}, nil
}

func (this_ *Service) AdvanceOpen(c *ctx.Context, _ *lessservicepb.User_AdvanceOpenRequest) (*lessservicepb.User_AdvanceOpenResponse, *errmsg.ErrMsg) {
	role, err := this_.Module.GetRoleModelByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(c)
	if role.LevelIndex >= reader.RoleLv.MaxRoleLevel() {
		return nil, errmsg.NewErrMaxLevel()
	}
	cfg, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	if cfg.CopyId == 0 {
		return nil, errmsg.NewErrCanNotAdvance()
	}
	dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
	if err != nil {
		return nil, err
	}
	// 有存档，说明已经开启过了
	if dungeon != nil {
		return &lessservicepb.User_AdvanceOpenResponse{}, nil
	}
	// 只扣道具不扣经验
	items := make(map[values.ItemId]values.Integer)
	for id, count := range cfg.AdvancedItem {
		items[id] += count
	}
	// if cfg.Exp > 0 {
	// 	items[enum.RoleExp] += cfg.Exp
	// }
	if err := this_.SubManyItem(c, c.RoleId, items); err != nil {
		return nil, err
	}
	dungeon = &dao.Dungeon{
		Id:        cfg.CopyId,
		Count:     0,
		ClearedAt: 0,
	}

	db.SaveAdvanceDungeon(c, dungeon)
	return &lessservicepb.User_AdvanceOpenResponse{}, nil
}

func (this_ *Service) AdvanceBattleStart(c *ctx.Context, _ *lessservicepb.User_AdvanceBattleStartRequest) (*lessservicepb.User_AdvanceBattleStartResponse, *errmsg.ErrMsg) {
	role, err := this_.Module.GetRoleModelByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(c)
	if role.LevelIndex >= reader.RoleLv.MaxRoleLevel() {
		return nil, errmsg.NewErrMaxLevel()
	}
	cfg, ok := reader.RoleLv.GetRoleLvById(role.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	if cfg.CopyId == 0 {
		return nil, errmsg.NewErrCanNotAdvance()
	}
	dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
	if err != nil {
		return nil, err
	}
	if dungeon == nil {
		return nil, errmsg.NewErrAdvanceNeedOpen()
	}
	if dungeon.ChallengeTime > timer.StartTime(c.StartTime).Unix() {
		return nil, errmsg.NewErrAdvanceBattleCD()
	}
	rrd, ok := reader.RoleReachDungeon.GetRoleReachDungeonById(cfg.CopyId)
	if !ok {
		return nil, errmsg.NewInternalErr("role_reach_dungeon not found")
	}
	buffs := this_.getBossDebuff(dungeon, rrd)
	dungeon.Count++
	// 发起战斗的时候更新一下这个挑战时间，防止未走结算
	dungeon.ChallengeTime = timer.StartTime(c.StartTime).Add(time.Duration(rrd.BossReduceTime) * time.Second).Unix()
	db.SaveAdvanceDungeon(c, dungeon)
	// // 挑战不扣经验，只扣道具，但需要判断道具和经验是否足够
	// if dungeon.Id == 0 {
	// 	if err := this_.BagService.SubManyItem(c, c.RoleId, cfg.AdvancedItem); err != nil {
	// 		return nil, errmsg.NewErrExpNotEnough()
	// 	}
	// 	expCount, err := this_.BagService.GetItem(c, c.RoleId, enum.RoleExp)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if expCount < cfg.Exp {
	// 		return nil, errmsg.NewErrExpNotEnough()
	// 	}
	// 	dungeon.Id = cfg.CopyId
	// 	db.SaveAdvanceDungeon(c, dungeon)
	// }

	mapId := cfg.CopyId
	tokenInfo := utils.TokenInfo{
		MapId:  mapId,
		RoleId: c.RoleId,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, utils.Claims{
		TokenInfo: tokenInfo,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: timer.StartTime(c.StartTime).Add(time.Hour * 1).Unix(),
		},
	})
	_token, err1 := token.SignedString(utils.JwtKey)
	if err1 != nil {
		return nil, errmsg.NewInternalErr("jwt sign failed")
	}
	// 获取英雄信息
	heroesFormation, err := this_.Module.FormationService.GetDefaultHeroes(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	heroIds := make([]int64, 0, 2)
	if heroesFormation.Hero_0 > 0 && heroesFormation.HeroOrigin_0 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_0)
	}
	if heroesFormation.Hero_1 > 0 && heroesFormation.HeroOrigin_1 > 0 {
		heroIds = append(heroIds, heroesFormation.HeroOrigin_1)
	}
	heroes, err := this_.Module.GetHeroes(c, c.RoleId, heroIds)
	if err != nil {
		return nil, err
	}
	equips, err := this_.GetManyEquipBagMap(c, c.RoleId, this_.GetHeroesEquippedEquipId(heroes)...)
	if err != nil {
		return nil, err
	}
	cppHeroes := trans.Heroes2CppHeroes(c, heroes, equips)
	if len(cppHeroes) == 0 {
		return nil, errmsg.NewErrHeroNotFound()
	}
	d, err00 := db.GetBattleSetting(c)
	if err00 != nil {
		return nil, err00
	}
	return &lessservicepb.User_AdvanceBattleStartResponse{
		MapId:    mapId,
		Token:    _token,
		BattleId: -10000, // 直接写死-10000 客户端必须，对于服务器无意义
		Sbp: &models.SingleBattleParam{
			Role:             role,
			Heroes:           cppHeroes,
			MonsterGroupInfo: rrd.MonsterGroupInfo,
			CountDown:        rrd.Times,
			MonsterBuffs:     buffs,
			AutoSoulSkill:    d.Data.AutoSoulSkill,
		},
	}, nil
}

func (this_ *Service) AdvanceBattleFinish(c *ctx.Context, req *lessservicepb.User_AdvanceBattleFinishRequest) (*lessservicepb.User_AdvanceBattleFinishResponse, *errmsg.ErrMsg) {
	token, err1 := jwt.ParseWithClaims(req.Token, &utils.Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return utils.JwtKey, nil
	})
	if err1 != nil {
		return nil, errmsg.NewInternalErr(err1.Error())
	}
	claims, ok := token.Claims.(*utils.Claims)
	if !ok || !token.Valid {
		return nil, errmsg.NewProtocolErrorInfo("invalid token")
	}
	if claims.RoleId != c.RoleId {
		return nil, errmsg.NewProtocolErrorInfo("invalid token")
	}
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	rrd, ok := reader.RoleReachDungeon.GetRoleReachDungeonById(cfg.CopyId)
	if !ok {
		return nil, errmsg.NewInternalErr("role_reach_dungeon not found")
	}
	dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
	if err != nil {
		return nil, err
	}
	// 说明未走正常的AdvanceBattleStart流程
	if dungeon == nil {
		return nil, errmsg.NewProtocolErrorInfo("dungeon is nil")
	}
	if req.Victory {
		dungeon.ClearedAt = time.Now().Unix()
	} else {
		dungeon.ChallengeTime = timer.StartTime(c.StartTime).Add(time.Duration(rrd.BossReduceTime) * time.Second).Unix()
	}
	db.SaveAdvanceDungeon(c, dungeon)
	return &lessservicepb.User_AdvanceBattleFinishResponse{}, nil
}

// func (this_ *Service) AdvanceBattleEnd(c *ctx.Context, req *servicepb.GameToClientBattle_CPPAdvanceBattleEndRequest) {
// 	if !req.Victory {
// 		this_.advanceBattleResultPush(c, req.Victory)
// 		return
// 	}
// 	r, err := db.GetRole(c, c.RoleId)
// 	if err != nil {
// 		this_.advanceBattleResultPush(c, false)
// 		c.Error("advanceBattleEnd get role failed", zap.Error(err))
// 		return
// 	}
// 	if r == nil {
// 		this_.advanceBattleResultPush(c, false)
// 		c.Error("advanceBattleEnd get role is nil", zap.String("roleId", c.RoleId))
// 		return
// 	}
// 	cfg, ok := rule.MustGetReader(c).RoleLv.GetRoleLvById(r.LevelIndex)
// 	if !ok {
// 		this_.advanceBattleResultPush(c, false)
// 		c.Error("advanceBattleEnd role_lv not found", zap.String("roleId", c.RoleId), zap.Int64("level index", r.LevelIndex))
// 		return
// 	}
// 	dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
// 	if err != nil {
// 		this_.advanceBattleResultPush(c, false)
// 		c.Error("advanceBattleEnd GetAdvanceDungeon failed", zap.Error(err))
// 		return
// 	}
// 	// 说明未走正常的AdvanceBattleStart流程
// 	if dungeon.Id == 0 {
// 		this_.advanceBattleResultPush(c, false)
// 		c.Error("advanceBattleEnd GetAdvanceDungeon is 0", zap.String("roleId", c.RoleId), zap.Int64("level index", r.LevelIndex))
// 		return
// 	}
// 	dungeon.ClearedAt = time.Now().Unix()
// 	if err := db.SaveAdvanceDungeon(c, dungeon); err != nil {
// 		this_.advanceBattleResultPush(c, false)
// 		c.Error("advanceBattleEnd SaveAdvanceDungeon failed", zap.Error(err))
// 		return
// 	}
// 	this_.advanceBattleResultPush(c, req.Victory)
// }

func (this_ *Service) AdvanceBattleVictory(c *ctx.Context, req *lessservicepb.User_AdvanceBattleVictoryRequest) (*lessservicepb.User_AdvanceBattleVictoryResponse, *errmsg.ErrMsg) {
	token, err1 := jwt.ParseWithClaims(req.Token, &utils.Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return utils.JwtKey, nil
	})
	if err1 != nil {
		return nil, errmsg.NewInternalErr(err1.Error())
	}
	claims, ok := token.Claims.(*utils.Claims)
	if !ok || !token.Valid {
		return nil, errmsg.NewProtocolErrorInfo("invalid token")
	}
	if claims.RoleId != c.RoleId {
		return nil, errmsg.NewProtocolErrorInfo("invalid token")
	}
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	cfg, ok := rule.MustGetReader(c).RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
	if err != nil {
		return nil, err
	}
	// 说明未走正常的AdvanceBattleStart流程
	if dungeon == nil {
		return nil, errmsg.NewProtocolErrorInfo("dungeon is nil")
	}
	dungeon.ClearedAt = time.Now().Unix()
	db.SaveAdvanceDungeon(c, dungeon)
	return &lessservicepb.User_AdvanceBattleVictoryResponse{}, nil
}

func (this_ *Service) GetSimpleRoles(ctx *ctx.Context, req *lessservicepb.User_GetSimpleRolesRequest) (*lessservicepb.User_GetSimpleRolesResponse, *errmsg.ErrMsg) {
	roles := make([]*dao.Role, 0, len(req.RoleIds))
	for _, roleId := range req.RoleIds {
		roles = append(roles, &dao.Role{RoleId: roleId})
	}
	newList, err := db.GetMultiRole(ctx, roles)
	if err != nil {
		return nil, err
	}
	return &lessservicepb.User_GetSimpleRolesResponse{Roles: RoleDao2SimpleModels(newList)}, nil
}

func (this_ *Service) AdvanceInfo(c *ctx.Context, _ *lessservicepb.User_AdvanceInfoRequest) (*lessservicepb.User_AdvanceInfoResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	// 只处理需要挑战副本的情况
	if cfg.CopyId == 0 {
		return &lessservicepb.User_AdvanceInfoResponse{}, nil
	}
	dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
	if err != nil {
		return nil, err
	}
	rrd, ok := reader.RoleReachDungeon.GetRoleReachDungeonById(cfg.CopyId)
	isOpen := dungeon != nil
	if dungeon == nil {
		dungeon = &dao.Dungeon{}
	}
	canAdvance := dungeon.ClearedAt > 0
	challengeCost := dungeon.Id == 0
	overlay := dungeon.Count
	if overlay > rrd.BossReduceMin {
		overlay = rrd.BossReduceMin
	}
	return &lessservicepb.User_AdvanceInfoResponse{
		IsOpen:        isOpen,
		BuffId:        rrd.BossReduce,
		Overlay:       overlay,
		ChallengeTime: dungeon.ChallengeTime,
		CanAdvance:    canAdvance,
		ChallengeCost: challengeCost,
	}, nil
}

func (this_ *Service) AdvanceUseItem(c *ctx.Context, req *lessservicepb.User_AdvanceUseItemRequest) (*lessservicepb.User_AdvanceUseItemResponse, *errmsg.ErrMsg) {
	item, ok := rule.MustGetReader(c).RoleReachItem.GetRoleReachItemById(req.ItemId)
	if !ok {
		return nil, errmsg.NewInternalErr("role_reach_item not exist")
	}
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(c)
	cfg, ok := reader.RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	// 只处理需要挑战副本的情况
	if cfg.CopyId == 0 {
		return &lessservicepb.User_AdvanceUseItemResponse{}, nil
	}
	dungeon, err := db.GetAdvanceDungeon(c, cfg.CopyId)
	if err != nil {
		return nil, err
	}
	if dungeon == nil || dungeon.ChallengeTime == 0 {
		return &lessservicepb.User_AdvanceUseItemResponse{}, nil
	}
	if err := this_.SubItem(c, c.RoleId, req.ItemId, 1); err != nil {
		return nil, err
	}
	dungeon.ChallengeTime -= item.ReduceTime
	db.SaveAdvanceDungeon(c, dungeon)
	return &lessservicepb.User_AdvanceUseItemResponse{
		ChallengeTime: dungeon.ChallengeTime,
	}, nil
}

func (this_ *Service) getBossDebuff(dungeon *dao.Dungeon, rrd *rulemodel.RoleReachDungeon) []*models.BuffInfo {
	buffs := make([]*models.BuffInfo, 0)
	if len(rrd.BossReduce) > 0 {
		for _, buffId := range rrd.BossReduce {
			buff := &models.BuffInfo{
				BuffId:  buffId,
				Overlay: int32(dungeon.Count),
			}
			if buff.Overlay > int32(rrd.BossReduceMin) {
				buff.Overlay = int32(rrd.BossReduceMin)
			}
			buffs = append(buffs, buff)
		}
	}
	return buffs
}

func (this_ *Service) titleUnlock(c *ctx.Context) *errmsg.ErrMsg {
	r, err := db.GetTitleRewards(c)
	if err != nil {
		return err
	}
	if r == nil {
		r = &dao.TitleRewards{RoleId: c.RoleId, Title: 0, Rewards: map[int64]bool{}}
	}
	if r.Title >= rule.MustGetReader(c).RoleLvTitle.GetMaxTitle() {
		return nil
	}
	r.Title++
	if r.Rewards == nil {
		r.Rewards = map[int64]bool{}
	}
	r.Rewards[r.Title] = true
	err = this_._titleUnlock(c, r.Title)
	if err != nil {
		return err
	}
	db.SaveTitleRewards(c, r)
	c.PushMessageToRole(c.RoleId, &servicepb.User_TitleRewardsPush{Info: trans.TitleRewardsD2M(r)})
	this_.UpdateTarget(c, c.RoleId, models.TaskType_TaskReachTitle, r.Title, 1)
	return nil
}

func (this_ *Service) _titleUnlock(c *ctx.Context, title values.Integer) *errmsg.ErrMsg {
	reader := rule.MustGetReader(c)
	titleCfg, ok := reader.RoleLvTitle.GetRoleLvTitleById(title)
	if !ok {
		return errmsg.NewInternalErr("role_lv_title config not found: " + strconv.Itoa(int(title)))
	}

	prevTitleSkill := rule2.GetPrevTitleSkill(c, title)
	attrData := event.NewAttrAdditionData()

	for entryId, added := range titleCfg.TitleSkill {
		entryCfg, ok := reader.EntrySkill.GetEntrySkillById(entryId)
		if !ok {
			return errmsg.NewInternalErr("entry_skill config not found: " + strconv.Itoa(int(entryId)))
		}

		entryTyp := entryCfg.Typ
		if len(entryTyp) < 2 {
			return errmsg.NewInternalErr("entry_skill config Typ error: " + strconv.Itoa(int(entryId)))
		}

		switch models.EntrySkillType(entryTyp[0]) {
		case models.EntrySkillType_ESTAttrAdd:
			if entryCfg.Value == 1 { // 固定加成
				attrData.AddFixed(entryTyp[1], added)
			}
			if entryCfg.Value == 2 { // 百分比加成
				attrData.AddPercent(entryTyp[1], added)
			}
		case models.EntrySkillType_ESTHeroAttrAdd:
			if len(entryTyp) < 3 {
				return errmsg.NewInternalErr("entry_skill config Typ error: " + strconv.Itoa(int(entryId)))
			}
			if entryCfg.Value == 1 { // 固定加成
				attrData.AddHeroFixed(entryTyp[2], entryTyp[1], added)
			}
			if entryCfg.Value == 2 { // 百分比加成
				attrData.AddHeroPercent(entryTyp[2], entryTyp[1], added)
			}
		default:
			c.PublishEventLocal(&event.ExtraSkillTypAdd{
				TypId:    models.EntrySkillType(entryTyp[0]),
				LogicId:  entryTyp[1],
				ValueTyp: entryCfg.Value,
				Cnt:      added - prevTitleSkill[entryId],
			})
		}
	}
	c.PublishEventLocal(&event.AttrUpdateToRole{
		Typ:         models.AttrBonusType_TypeTitle,
		AttrFixed:   attrData.Base.Fixed,
		AttrPercent: attrData.Base.Percent,
		IsCover:     true,
	})
	for heroId, addition := range attrData.Hero {
		c.PublishEventLocal(&event.AttrUpdateToRole{
			Typ:         models.AttrBonusType_TypeTitle,
			AttrFixed:   addition.Fixed,
			AttrPercent: addition.Percent,
			IsCover:     true,
			HeroId:      heroId,
		})
	}

	return nil
}

func (this_ *Service) GetTitleRewardsInfo(c *ctx.Context, _ *servicepb.User_GetTitleRewardsInfoRequest) (*servicepb.User_GetTitleRewardsInfoResponse, *errmsg.ErrMsg) {
	r, err := db.GetTitleRewards(c)
	if err != nil {
		return nil, err
	}
	if r == nil { // 解锁第一个头衔 并且添加特权
		r = &dao.TitleRewards{RoleId: c.RoleId, Title: 1, Rewards: map[int64]bool{1: true}}
		err = this_._titleUnlock(c, r.Title)
		if err != nil {
			return nil, err
		}
		db.SaveTitleRewards(c, r)
		return nil, errmsg.NewErrUserNotFound()
	}

	return &servicepb.User_GetTitleRewardsInfoResponse{
		Info: trans.TitleRewardsD2M(r),
	}, nil
}

func (this_ *Service) UseCdKey(c *ctx.Context, req *servicepb.User_UseCdKeyRequest) (*servicepb.User_UseCdKeyResponse, *errmsg.ErrMsg) {
	/*data, err := db.GetKeyGen(c, req.CdKey)
	if err != nil {
		return nil, errmsg.NewErrCdKeyIsNotExist()
	}
	if data.LimitCnt == 0 || !data.IsActive {
		return nil, errmsg.NewErrCdKeyNotActive()
	}
	isUsed, err := db.HasKeyUse(c.RoleId, data.BatchId)
	if err != nil {
		return nil, err
	}
	if isUsed {
		return nil, errmsg.NewErrCdKeyIsUsed()
	}
	usedData := &db.CdKeyUse{
		RoleId:    c.RoleId,
		BatchId:   data.BatchId,
		KeyId:     data.Id,
		CreatedAt: timer.Unix(),
	}
	switch data.LimitTyp {
	case 1:
		if err = db.KeyUseRecord(c, usedData); err != nil {
			return nil, err
		}
	case 2:
		if err = db.KeyUseDel(c, usedData); err != nil {
			return nil, err
		}
	case 3:
		if err = db.KeyUseSub(c, usedData); err != nil {
			return nil, err
		}
	}
	mapData := make(map[int64]int64)
	if len(data.Reward) < 2 {
		return nil, errmsg.NewErrCdKeyIsNotExist()
	}
	temp := strings.Split(data.Reward[1:len(data.Reward)-1], ",")
	if len(temp) <= 0 {
		return nil, errmsg.NewErrCdKeyIsNotExist()
	}
	for _, item := range temp {
		s := strings.Split(item, ":")
		if len(s) != 2 {
			return nil, errmsg.NewErrCdKeyIsNotExist()
		}
		mapK, err := strconv.Atoi(s[0])
		if err != nil {
			return nil, errmsg.NewErrCdKeyIsNotExist()
		}
		mapV, err := strconv.Atoi(s[1])
		if err != nil {
			return nil, errmsg.NewErrCdKeyIsNotExist()
		}
		mapData[int64(mapK)] = int64(mapV)
	}
	this_.BagService.AddManyItem(c, c.RoleId, mapData)
	return &servicepb.User_UseCdKeyResponse{Items: mapToItems(mapData)}, nil*/
	return nil, nil
}

func mapToItems(m map[values.Integer]values.Integer) []*models.Item {
	items := make([]*models.Item, 0, len(m))
	for k, v := range m {
		items = append(items, &models.Item{
			ItemId: k,
			Count:  v,
		})
	}
	return items
}

func (this_ *Service) DrawTitleRewards(c *ctx.Context, req *servicepb.User_DrawTitleRewardsRequest) (*servicepb.User_DrawTitleRewardsResponse, *errmsg.ErrMsg) {
	r, err := db.GetTitleRewards(c)
	if err != nil {
		return nil, err
	}
	if r == nil || !r.Rewards[req.Title] {
		return nil, errmsg.NewErrUserCannotDrawTitleReward()
	}

	cfg, ok := rule.MustGetReader(c).RoleLvTitle.GetRoleLvTitleById(req.Title)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv_title not found: " + strconv.Itoa(int(req.Title)))
	}
	_, err = this_.AddManyItem(c, c.RoleId, cfg.TitleReward)
	if err != nil {
		return nil, err
	}

	delete(r.Rewards, req.Title)
	db.SaveTitleRewards(c, r)

	return &servicepb.User_DrawTitleRewardsResponse{
		Title: req.Title,
		Items: trans.ItemMapToProto(cfg.TitleReward),
	}, nil
}

func (this_ *Service) GetRoleSkill(ctx *ctx.Context, roleId values.RoleId) ([]values.Integer, *errmsg.ErrMsg) {
	skills, err := db.GetRoleSkill(ctx, roleId)
	if err != nil {
		return nil, err
	}
	return skills.SkillId, nil
}

func (this_ *Service) RecommendFriend(c *ctx.Context, request *lessservicepb.User_RecommendFriendRequest) (*lessservicepb.User_RecommendFriendResponse, *errmsg.ErrMsg) {
	req := &recommend.Recommend_RecommendRequest{Language: request.Language}
	out := &recommend.Recommend_RecommendResponse{}
	err := this_.svc.GetNatsClient().RequestWithOut(c, 0, req, out)
	if err != nil {
		return nil, err
	}
	if len(out.Id) == 0 {
		return nil, nil
	}
	roles := make([]*dao.Role, len(out.Id))
	for i := range out.Id {
		roles[i] = &dao.Role{RoleId: out.Id[i]}
	}
	newList, err := db.GetMultiRole(c, roles)
	if err != nil {
		return nil, err
	}
	return &lessservicepb.User_RecommendFriendResponse{Roles: RoleDao2Models(newList)}, nil
}

func (this_ *Service) saveRoleId(ctx *ctx.Context, id values.Integer, roleId values.RoleId) *errmsg.ErrMsg {
	redis := redisclient.GetDefaultRedis()
	k := enum.GetRoleIdKey(id)
	if err := redis.HSet(ctx, k, roleId, 0).Err(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

func (this_ *Service) GetRedPoints(c *ctx.Context, _ *lessservicepb.User_GetReadPointRequest) (*lessservicepb.User_GetReadPointResponse, *errmsg.ErrMsg) {
	rp, err := db.GetReadPoint(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	return &lessservicepb.User_GetReadPointResponse{
		RedPoints: rp.RedPoints,
	}, nil
}

func (this_ *Service) AddRedPoints(c *ctx.Context, req *lessservicepb.User_AddReadPointRequest) (*lessservicepb.User_AddReadPointResponse, *errmsg.ErrMsg) {
	rp, err := db.GetReadPoint(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	val := rp.RedPoints[req.Key]
	val += req.Cnt
	if val < 0 {
		val = 0
	}
	rp.RedPoints[req.Key] = val
	return &lessservicepb.User_AddReadPointResponse{
		Key: req.Key,
		Cnt: val,
	}, nil
}

func (this_ *Service) SetRedPoints(c *ctx.Context, req *lessservicepb.User_SetReadPointRequest) (*lessservicepb.User_SetReadPointResponse, *errmsg.ErrMsg) {
	rp, err := db.GetReadPoint(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	val := rp.RedPoints[req.Key]
	val = req.Cnt
	if val < 0 {
		val = 0
	}
	rp.RedPoints[req.Key] = val
	return &lessservicepb.User_SetReadPointResponse{
		Key: req.Key,
		Cnt: val,
	}, nil
}

func (this_ *Service) SetBattleSpeed(ctx *ctx.Context, req *lessservicepb.User_SetBattleSpeedRequest) (*lessservicepb.User_SetBattleSpeedResponse, *errmsg.ErrMsg) {
	// TODO 最大和最小值待定
	if req.Speed < 1 || req.Speed > 10 {
		return nil, nil
	}
	role, err := db.GetRole(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	role.BattleSpeed = req.Speed
	db.SaveRole(ctx, role)
	return &lessservicepb.User_SetBattleSpeedResponse{}, nil
}

func (this_ *Service) GetUserSimpleInfo(ctx *ctx.Context, req *lessservicepb.User_GetUserSimpleInfoRequest) (*lessservicepb.User_GetUserSimpleInfoResponse, *errmsg.ErrMsg) {
	if ctx.ServerType == models.ServerType_GatewayStdTcp {
		return nil, errmsg.NewProtocolErrorInfo("not support")
	}
	role, err := db.GetRole(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	guild, err := this_.GetUserGuildInfo(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	var (
		position  values.Integer
		guildId   values.GuildId
		guildName string
	)
	if guild != nil {
		guildId = guild.Id
		guildName = guild.Name
		position, err = this_.GuildService.GetUserGuildPositionByGuildId(ctx, guild.Id)
		if err != nil {
			return nil, err
		}
	}
	list, err := this_.GetAllHero(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	heroes := make([]*models.HeroSimpleInfo, 0)
	for _, hero := range list {
		heroes = append(heroes, &models.HeroSimpleInfo{
			ConfigId: hero.Id,
			Power:    hero.CombatValue.Total,
		})
	}
	info := &models.UserSimpleInfo{
		RoleId:        role.RoleId,
		Nickname:      role.Nickname,
		Level:         role.Level,
		AvatarId:      role.AvatarId,
		AvatarFrame:   role.AvatarFrame,
		Power:         role.Power,
		Title:         role.Title,
		GuildId:       guildId,
		GuildName:     guildName,
		GuildPosition: position,
		Heroes:        heroes,
	}
	return &lessservicepb.User_GetUserSimpleInfoResponse{
		Info: info,
	}, nil
}

func (this_ *Service) UserCombatValueDetails(ctx *ctx.Context, req *lessservicepb.User_GetUserCombatValueDetailsRequest) (*lessservicepb.User_GetUserCombatValueDetailsResponse, *errmsg.ErrMsg) {
	if req.RoleId == "" {
		return nil, errmsg.NewProtocolErrorInfo("invalid request")
	}
	selfHeroes, err := this_.GetAllHero(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	targetHeroes, err := this_.GetAllHero(ctx, req.RoleId)
	if err != nil {
		return nil, err
	}
	self := make(map[values.Integer]values.Integer)
	target := make(map[values.Integer]values.Integer)
	for _, hero := range selfHeroes {
		if hero.CombatValue == nil || hero.CombatValue.Details == nil {
			continue
		}
		for id, val := range hero.CombatValue.Details {
			self[id] += val
		}
	}
	for _, hero := range targetHeroes {
		if hero.CombatValue == nil || hero.CombatValue.Details == nil {
			continue
		}
		for id, val := range hero.CombatValue.Details {
			target[id] += val
		}
	}

	return &lessservicepb.User_GetUserCombatValueDetailsResponse{
		Self:   self,
		Target: target,
	}, nil
}

func (this_ *Service) syncTopRank(ctx *ctx.Context, lastTitle, currentTitle, combatValue values.Integer) *errmsg.ErrMsg {
	if combatValue <= 0 {
		return nil
	}
	limit := &dao.TopRankLimit{Title: currentTitle}
	_, err := ctx.NewOrm().GetPB(enum.GetRedisClient(), limit)
	if err != nil {
		return err
	}
	if limit.Full && combatValue < limit.Min {
		return nil
	}
	// // TODO 3000000修正线上数据，等下次更新去掉
	// temp := values.Integer(100000)
	// if currentTitle == 1 {
	// 	temp = 100000
	// } else if currentTitle == 2 {
	// 	temp = 2000000
	// } else if currentTitle == 3 {
	// 	temp = 3000000
	// } else {
	// 	temp = limit.Min
	// }
	// if limit.Full && combatValue < temp {
	// 	return nil
	// }
	if _, err := this_.svc.GetNatsClient().Request(ctx, 0, &rank_service.RankService_TopRankUpdateRequest{
		RoleId:       ctx.RoleId,
		CombatValue:  combatValue,
		LastTitle:    lastTitle,
		CurrentTitle: currentTitle,
	}); err != nil {
		return err
	}
	return nil
}

func (this_ *Service) HangExpTimerPush(ctx *ctx.Context) *errmsg.ErrMsg {
	hangExpTimer, err := db.GetHangExpTimer(ctx)
	if err != nil {
		return err
	}
	if !hangExpTimer.Timing {
		ctx.PushMessage(&lessservicepb.User_UserCanGetExpFromHangPush{Can: true})
	} else {
		sec := rule2.GetExpGetTimeLimit(ctx)
		ctx.PushMessage(&lessservicepb.User_UserCanGetExpFromHangPush{
			Can: timer.StartTime(ctx.StartTime).Unix()-hangExpTimer.Time < sec,
		})
	}
	return nil
}

func (this_ *Service) NameExist(_ *ctx.Context, req *servicepb.User_PlayerNameExistRequest) (*servicepb.User_PlayerNameExistResponse, *errmsg.ErrMsg) {
	if req.Name == "" {
		return nil, errmsg.NewErrInvalidRequestParam()
	}
	exist, err := db.NameExist(req.Name)
	if err != nil {
		return nil, err
	}
	return &servicepb.User_PlayerNameExistResponse{
		Exist: exist,
	}, nil
}

func (this_ *Service) UpdateUserCurrency(ctx *ctx.Context, req *servicepb.User_UpdateUserCurrencyRequest) (*servicepb.User_UpdateUserCurrencyResponse, *errmsg.ErrMsg) {
	addMap := make(map[values.ItemId]values.Integer)
	subMap := make(map[values.ItemId]values.Integer)
	subList := make([]values.ItemId, 0)
	tempMap := make(map[values.ItemId]struct{})
	for _, id := range enum.CurrencyList {
		tempMap[id] = struct{}{}
	}
	for id, count := range req.Currency {
		if _, ok := tempMap[id]; !ok {
			continue
		}
		if count == 0 {
			continue
		}
		if count > 0 {
			addMap[id] += count
		} else {
			subMap[id] += -count
			subList = append(subList, id)
		}
	}
	if len(addMap) == 0 && len(subMap) == 0 {
		return &servicepb.User_UpdateUserCurrencyResponse{}, nil
	}
	if len(subMap) > 0 {
		items, err := this_.GetManyItem(ctx, ctx.RoleId, subList)
		if err != nil {
			return nil, err
		}
		for id, count := range subMap {
			if items[id] < count {
				subMap[id] = items[id]
			}
		}
	}
	if _, err := this_.AddManyItem(ctx, ctx.RoleId, addMap); err != nil {
		return nil, err
	}
	if err := this_.SubManyItem(ctx, ctx.RoleId, subMap); err != nil {
		return nil, err
	}
	return &servicepb.User_UpdateUserCurrencyResponse{}, nil
}

// -----------------------------------------------cheat-----------------------------------------------//

func (this_ *Service) CheatGetRolesRequest(ctx *ctx.Context, _ *lessservicepb.User_CheatGetRolesRequest) (*lessservicepb.User_CheatGetRolesResponse, *errmsg.ErrMsg) {
	roles := make([]*dao.Role, 1)
	roles[0] = &dao.Role{RoleId: "9G4EC"}
	newList, err := db.GetMultiRole(ctx, roles)
	if err != nil {
		return nil, err
	}
	return &lessservicepb.User_CheatGetRolesResponse{Roles: RoleDao2Models(newList)}, nil
}

func (this_ *Service) CheatModifyTimeRequest(ctx *ctx.Context, req *servicepb.User_CheatModifyTimeRequest) (*servicepb.User_CheatModifyTimeResponse, *errmsg.ErrMsg) {
	d := time.Duration(req.TimeOffset / 1000)
	timer.Timer.SetOffset(d)
	return &servicepb.User_CheatModifyTimeResponse{}, nil
}

func (this_ *Service) CheatSetLevel(ctx *ctx.Context, req *lessservicepb.User_CheatSetLevelRequest) (*lessservicepb.User_CheatSetLevelResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, ctx.RoleId)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if r == nil {
		return nil, errmsg.NewErrUserNotFound()
	}
	reader := rule.MustGetReader(ctx)
	if req.Level > reader.RoleLv.MaxRoleLevel() {
		return nil, errmsg.NewErrMaxLevel()
	}

	r.LevelIndex = req.Level
	cfg2, ok := reader.RoleLv.GetRoleLvById(r.LevelIndex)
	if !ok {
		return nil, errmsg.NewInternalErr("role_lv not found")
	}
	// 为突破等级，清除突破存档，方便反复测试
	if cfg2.CopyId > 0 {
		dungeon, err := db.GetAdvanceDungeon(ctx, cfg2.CopyId)
		if err != nil {
			return nil, err
		}
		if dungeon == nil {
			dungeon = &dao.Dungeon{}
		}
		if dungeon.Id == cfg2.CopyId {
			dungeon.ClearedAt = 0
			db.SaveAdvanceDungeon(ctx, dungeon)
		}
	}
	last := r.Level
	lastIndex := r.LevelIndex
	r.Level = cfg2.LvHero
	db.SaveRole(ctx, r)

	// 这里的incr可能为负数
	ctx.PublishEventLocal(&event.UserLevelChange{
		Level:          r.Level,
		Incr:           r.Level - last,
		IsAdvance:      false,
		LevelIndex:     r.LevelIndex,
		LevelIndexIncr: r.LevelIndex - lastIndex,
	})
	this_.UpdateTarget(ctx, ctx.RoleId, models.TaskType_TaskLevel, 0, r.Level, true)
	return &lessservicepb.User_CheatSetLevelResponse{
		Level:      r.Level,
		LevelIndex: r.LevelIndex,
	}, nil
}

func (this_ *Service) CheatModifyCreateTimeRequest(ctx *ctx.Context, req *servicepb.User_CheatModifyCreateTimeRequest) (*servicepb.User_CheatModifyCreateTimeResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(ctx, ctx.RoleId)
	if err != nil {
		return nil, err
	}
	r.CreateTime = req.CreateAt
	db.SaveRole(ctx, r)
	return &servicepb.User_CheatModifyCreateTimeResponse{}, nil
}

func (this_ *Service) CheatPushMsgToMobile(c *ctx.Context, req *lessservicepb.User_CheatPushToMobileRequest) (*lessservicepb.User_CheatPushToMobileResponse, *errmsg.ErrMsg) {
	if req.IggId != "" && req.GameId > 0 {
		iggsdk.GetPushIns().SendMsg(req.GameId, req.IggId, req.Context)
	}
	return &lessservicepb.User_CheatPushToMobileResponse{}, nil
}

func (this_ *Service) CheatAheadRegisterDay(c *ctx.Context, req *lessservicepb.User_CheatAheadRegisterDayRequest) (*lessservicepb.User_CheatAheadRegisterDayResponse, *errmsg.ErrMsg) {
	r, err := db.GetRole(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	r.CreateTime -= 86400000 * req.Day
	db.SaveRole(c, r)
	return &lessservicepb.User_CheatAheadRegisterDayResponse{}, nil
}

func (this_ *Service) md5Base32(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
