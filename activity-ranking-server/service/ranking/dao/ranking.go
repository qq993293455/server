package dao

import (
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetActivityManagerInfo(ctx *ctx.Context, serverId int64) (*dao.ActivityManagerInfo, bool) {
	ret := &dao.ActivityManagerInfo{ServerId: strconv.FormatInt(serverId, 10)}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return &dao.ActivityManagerInfo{
			ServerId: strconv.FormatInt(serverId, 10),
			Infos:    make(map[string]*models.ActivityRanking_Info),
		}, true
	}
	return ret, false
}

func SaveActivityManagerInfo(ctx *ctx.Context, data *dao.ActivityManagerInfo) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func DelActivityManagerInfo(ctx *ctx.Context, serverId int64) {
	data := &dao.ActivityManagerInfo{ServerId: strconv.FormatInt(serverId, 10)}
	ctx.NewOrm().DelPB(redisclient.GetUserRedis(), data)
}

func GetRankingKey(rankingIndex string, version int64) string {
	return utils.GenDefaultRedisKey(values.Activity, values.Hash, rankingIndex+":"+strconv.FormatInt(version, 10))
}

func SaveActivityRankingData(ctx *ctx.Context, rankingIndex string, version int64, data *models.ActivityRanking_Data) {
	infoPb := &dao.ActivityRankingData{
		RoleId: data.RoleId,
		Data:   data,
	}
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), GetRankingKey(rankingIndex, version), infoPb)
}

func GetActivityRankingData(ctx *ctx.Context, rankingIndex string, version int64) []*dao.ActivityRankingData {
	rankingDataList := make([]*dao.ActivityRankingData, 0)
	ctx.NewOrm().HScanGetAll(redisclient.GetUserRedis(), GetRankingKey(rankingIndex, version), &rankingDataList)
	return rankingDataList
}

func DelActivityRankingData(ctx *ctx.Context, rankingIndex string, version int64, roleId string) {
	infoPb := &dao.ActivityRankingData{
		RoleId: roleId,
	}
	delInfo := []orm.RedisInterface{infoPb}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), GetRankingKey(rankingIndex, version), delInfo)
}
