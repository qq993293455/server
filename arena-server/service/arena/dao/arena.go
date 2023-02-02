package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
	"strconv"
)

var (
	ArenaRankIdBase      int64 = 100000000
	ArenaRankIdTypeStart int64 = 1000000
)

func GetServerStartTime(ctx *ctx.Context) int64 {
	ret := &dao.ServerData{Identification: "ServerStart"}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		ret.ServerStartTime = timer.BeginOfDay(timer.Now()).Unix()
		ctx.NewOrm().SetPB(redisclient.GetUserRedis(), ret)
	}
	return ret.ServerStartTime
}

func GetArenaTypeInfos(ctx *ctx.Context, atype models.ArenaType) (*dao.ArenaRankingTypeInfos, bool) {
	ret := &dao.ArenaRankingTypeInfos{Type: atype}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return &dao.ArenaRankingTypeInfos{
			Type: atype,
			TypeInfos: &models.ArenaRanking_TypeInfos{
				Type:              atype,
				Index:             0,
				SeasonRewardIndex: 1,
				DayRewardIndex:    1,
			},
		}, true
	}
	return ret, false
}

func SaveArenaTypeInfos(ctx *ctx.Context, typeInfos *dao.ArenaRankingTypeInfos) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), typeInfos)
}

func GetArenaInfos(ctx *ctx.Context, rankingIndex string) *dao.ArenaRankingInfos {
	ret := &dao.ArenaRankingInfos{RankingIndex: rankingIndex}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret
}

func SaveArenaInfos(ctx *ctx.Context, infos *dao.ArenaRankingInfos) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), infos)
}

func GetNewRankingIndex(atype models.ArenaType, index int64) string {
	return strconv.FormatInt(int64(atype), 10) + ":" + strconv.FormatInt(ArenaRankIdBase+ArenaRankIdTypeStart*int64(atype)+index, 10)
}

func SaveChange(ctx *ctx.Context, aType models.ArenaType, rankingIndex string, seasonRewardIndex uint64, dayRewardIndex uint64, roleId string, rankingId int32) {
	changeInfo := &dao.ArenaRoleRankingChangeInfo{
		RewardIndex: seasonRewardIndex,
		RankingId:   rankingId,
		ChangeTime:  timer.Now().Unix(),
	}

	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getRewardKey(roleId, true, int64(aType)), changeInfo)
	changeInfo.RewardIndex = dayRewardIndex
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getRewardKey(roleId, false, int64(aType)), changeInfo)
}

func getRewardKey(roleId values.RoleId, isSeansonOver bool, aType int64) string {
	if isSeansonOver {
		return utils.GenRedisKey(values.Arena, values.Hash, roleId, "season:"+strconv.FormatInt(aType, 10))
	}
	return utils.GenRedisKey(values.Arena, values.Hash, roleId, "day:"+strconv.FormatInt(aType, 10))
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

func SaveArenaDayRewardIndex(ctx *ctx.Context, aType models.ArenaType, rewardInfo []*models.ArenaSendRewardTime) {
	rewardIndex := &dao.ArenaDayRewardIndex{
		Type: aType,
		Info: rewardInfo,
	}
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), rewardIndex)
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

func SaveArenaSeasonRewardIndex(ctx *ctx.Context, aType models.ArenaType, rewardInfo []*models.ArenaSendRewardTime) {
	rewardIndex := &dao.ArenaSeasonRewardIndex{
		Type: aType,
		Info: rewardInfo,
	}
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), rewardIndex)
}

func GetRoleKey(rankingIndex string) string {
	return utils.GenDefaultRedisKey(values.Arena, values.Hash, rankingIndex)
}

func SaveArenaRoleInfo(ctx *ctx.Context, rankingIndex string, info *models.ArenaRanking_Info) {
	infoPb := &dao.ArenaRankingRole{
		RankingRoleKey: info.RoleId,
		Info:           info,
	}
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), GetRoleKey(rankingIndex), infoPb)
}

func LoadArenaAllRoleInfo(ctx *ctx.Context, rankingIndex string) []*dao.ArenaRankingRole {
	roleInfos := make([]*dao.ArenaRankingRole, 0)
	ctx.NewOrm().HScanGetAll(redisclient.GetUserRedis(), GetRoleKey(rankingIndex), &roleInfos)
	return roleInfos
}

func SaveArenaRankingIndex(ctx *ctx.Context, aType models.ArenaType, rankingIndex string) {
	rankingIndexPb := &dao.ArenaRankingIndex{
		RankingIndex: rankingIndex,
		CreateTime:   timer.Now().Unix(),
	}
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), GetRankingIndexKey(aType), rankingIndexPb)
}

func GetRankingIndexKey(aType models.ArenaType) string {
	return utils.GenDefaultRedisKey(values.Arena, values.Hash, "rankingIndex:"+strconv.FormatInt(int64(aType), 10))
}

func LoadArenaRankingIndex(ctx *ctx.Context, aType models.ArenaType) []*dao.ArenaRankingIndex {
	rankingIndexs := make([]*dao.ArenaRankingIndex, 0)
	ctx.NewOrm().HScanGetAll(redisclient.GetUserRedis(), GetRankingIndexKey(aType), &rankingIndexs)
	return rankingIndexs
}
