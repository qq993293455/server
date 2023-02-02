package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	"fmt"
	"strconv"
)

var (
	ArenaRankIdBase      int64 = 100000000
	ArenaRankIdTypeStart int64 = 1000000
)

type ArenaI interface {
	Save()
	GetData() *dao.ArenaData
	UseFreeChallengeTimes(atype models.ArenaType) bool
	RefreshChallengeTimes(reset_times int64)
	SetFightHero(atype models.ArenaType, fight_hero *models.Assemble) bool
	SetleftTicketPurchasesNumber(atype models.ArenaType, num int64) bool
}

func NewArena(ctx *ctx.Context, values *dao.ArenaData) ArenaI {
	return &ArenaData{
		arena_data: values,
		ctx:        ctx,
	}
}

type ArenaData struct {
	arena_data *dao.ArenaData
	ctx        *ctx.Context
}

func CreateArenaData(roleId values.RoleId, ctype models.ArenaType) *dao.ArenaData {
	ret := &dao.ArenaData{
		RoleId: roleId,
		Data: map[int32]*models.ArenaData{
			int32(ctype): {
				FightHero: &models.Assemble{
					HeroOrigin_0: 0,
					HeroOrigin_1: 0,
				},
				RankingIndex:       "",
				LastRefreshTime:    timer.Now().Unix(),
				FreeChallengeTimes: 0,
				SeasonRewardIndex:  0,
				DayRewardIndex:     0,
			},
		},
	}
	return ret
}

func GetPlayerArenaData(ctx *ctx.Context, role_id values.RoleId) (ArenaI, bool) {
	ret := &dao.ArenaData{RoleId: role_id}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	var isNew bool = false
	if !has {
		ret = CreateArenaData(role_id, models.ArenaType_ArenaType_Default)
		isNew = true
	}

	return NewArena(ctx, ret), isNew
}

func (this_ *ArenaData) Save() {
	this_.ctx.NewOrm().SetPB(redisclient.GetUserRedis(), this_.arena_data)
	this_.ctx.NewOrm().Do()
}

func (this_ *ArenaData) GetData() *dao.ArenaData {
	return this_.arena_data
}

func (this_ *ArenaData) UseFreeChallengeTimes(atype models.ArenaType) bool {
	data, ok := this_.arena_data.Data[int32(atype)]
	if !ok {
		return false
	}
	if data.FreeChallengeTimes <= 0 {
		return false
	}
	data.FreeChallengeTimes--
	this_.Save()
	return true
}

func (this_ *ArenaData) RefreshChallengeTimes(reset_times int64) {
	for _, data := range this_.arena_data.Data {
		data.FreeChallengeTimes = reset_times
		data.LastRefreshTime = timer.Now().Unix()
	}
	this_.Save()
}

func (this_ *ArenaData) SetFightHero(atype models.ArenaType, fight_hero *models.Assemble) bool {
	data, ok := this_.arena_data.Data[int32(atype)]
	if !ok {
		return false
	}
	data.FightHero = fight_hero
	this_.Save()
	return true
}

func (this_ *ArenaData) SetleftTicketPurchasesNumber(atype models.ArenaType, num int64) bool {
	data, ok := this_.arena_data.Data[int32(atype)]
	if !ok {
		return false
	}
	data.LeftTicketPurchasesNum = num
	this_.Save()
	return true
}

func GetAllFightLog(ctx *ctx.Context, aType models.ArenaType, roleId values.RoleId) []*dao.ArenaFightLogs {
	fightLogs := make([]*dao.ArenaFightLogs, 0)
	ctx.NewOrm().HScanGetAll(redisclient.GetUserRedis(), getFightKey(aType, roleId), &fightLogs)
	return fightLogs
}

func GetFightLog(ctx *ctx.Context, aType models.ArenaType, roleId values.RoleId) (*dao.ArenaFightLogs, *errmsg.ErrMsg) {
	fightLogs := &dao.ArenaFightLogs{FightIndex: GetFightIndex(aType, roleId)}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getFightKey(aType, roleId), fightLogs)
	if err != nil {
		return nil, err
	}
	if !has {
		return &dao.ArenaFightLogs{
			FightDayBegin: timer.BeginOfDay(timer.Now()).Unix(),
		}, nil
	}
	return fightLogs, nil
}

func DelFightLog(ctx *ctx.Context, aType models.ArenaType, roleId values.RoleId, delFightLogs []*dao.ArenaFightLogs) {
	var indexs []string
	for _, data := range delFightLogs {
		indexs = append(indexs, data.PK())
	}
	ctx.NewOrm().HDel(redisclient.GetUserRedis(), getFightKey(aType, roleId), indexs...)
}

func SetFightHero(ctx *ctx.Context, aType models.ArenaType, roleId values.RoleId, fightLogs *dao.ArenaFightLogs) {
	if fightLogs.FightIndex == "" {
		fightLogs.FightIndex = GetFightIndex(aType, roleId)
	}
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getFightKey(aType, roleId), fightLogs)
}

func getFightKey(aType models.ArenaType, roleId values.RoleId) string {
	return utils.GenRedisKey(values.Arena, values.Hash, roleId, fmt.Sprintf("%d:FightLogs", aType))
}

func GetFightIndex(aType models.ArenaType, roleId string) string {
	y, m, d := timer.Now().Date()
	return fmt.Sprintf("%d:%s:%02d%02d%02d", aType, roleId, y, m, d)
}

func LoadArenaDayRewardIndex(ctx *ctx.Context, aType models.ArenaType) []*models.ArenaSendRewardTime {
	ret := &dao.ArenaDayRewardIndex{
		Type: aType,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret.Info
}

func LoadArenaSeasonRewardIndex(ctx *ctx.Context, aType models.ArenaType) []*models.ArenaSendRewardTime {
	ret := &dao.ArenaSeasonRewardIndex{
		Type: aType,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret.Info
}

func getRewardKey(roleId values.RoleId, isSeansonOver bool, aType int64) string {
	if isSeansonOver {
		return utils.GenRedisKey(values.Arena, values.Hash, roleId, "season:"+strconv.FormatInt(aType, 10))
	}
	return utils.GenRedisKey(values.Arena, values.Hash, roleId, "day:"+strconv.FormatInt(aType, 10))
}

func GetRankingChangeInfo(ctx *ctx.Context, roleId string, isSeanson bool, aType int64) []*dao.ArenaRoleRankingChangeInfo {
	changeInfo := make([]*dao.ArenaRoleRankingChangeInfo, 0)
	ctx.NewOrm().HScanGetAll(redisclient.GetUserRedis(), getRewardKey(roleId, isSeanson, aType), &changeInfo)
	return changeInfo
}

func DelRankingChangeInfo(ctx *ctx.Context, roleId string, isSeanson bool, aType int64, pk string) {
	ctx.NewOrm().HDel(redisclient.GetUserRedis(), getRewardKey(roleId, isSeanson, aType), pk)
}
