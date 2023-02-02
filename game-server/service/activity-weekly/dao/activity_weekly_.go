package dao

import (
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetActivityWeeklyData(ctx *ctx.Context, roleId string) (*dao.ActivityWeeklyData, *errmsg.ErrMsg) {
	ret := &dao.ActivityWeeklyData{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func SaveActivityWeeklyData(ctx *ctx.Context, data *dao.ActivityWeeklyData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func GetRankingKey(rankingIndex string, version int64) string {
	return utils.GenDefaultRedisKey(values.Activity, values.Hash, rankingIndex+":"+strconv.FormatInt(version, 10))
}

func GetRankingData(ctx *ctx.Context, rankingIndex string, version int64, roleId string) (bool, *dao.ActivityRankingData, *errmsg.ErrMsg) {
	rankingData := &dao.ActivityRankingData{RoleId: roleId}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), GetRankingKey(rankingIndex, version), rankingData)
	if err != nil {
		return true, nil, err
	}
	return has, rankingData, nil
}

func DelRankingData(ctx *ctx.Context, rankingIndex string, version int64, roleId string) {
	infoPb := &dao.ActivityRankingData{
		RoleId: roleId,
	}
	delInfo := []orm.RedisInterface{infoPb}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), GetRankingKey(rankingIndex, version), delInfo)
}

func genGuildChallengeKey(activityId, version values.Integer, guildKey string) string {
	return strconv.FormatInt(activityId, 10) + ":" + strconv.FormatInt(version, 10) + ":" + guildKey
}

func GetGuildChallengeInfo(ctx *ctx.Context, activityId, version values.Integer, guildId string) (*dao.ActivityWeeklyGuildData, *errmsg.ErrMsg) {
	key := genGuildChallengeKey(activityId, version, guildId)
	ret := &dao.ActivityWeeklyGuildData{GuildKey: key}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func SaveGuildChallengeInfo(ctx *ctx.Context, data *dao.ActivityWeeklyGuildData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func DelGuildChallengeInfo(ctx *ctx.Context, activityId, version values.Integer, guildId string) {
	key := genGuildChallengeKey(activityId, version, guildId)
	data := &dao.ActivityWeeklyGuildData{GuildKey: key}
	ctx.NewOrm().DelPB(redisclient.GetUserRedis(), data)
}
