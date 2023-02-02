package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetPreDownloadInfo(ctx *ctx.Context, roleId values.RoleId) *dao.PreDownloadData {
	ret := &dao.PreDownloadData{
		RoleId: roleId,
	}
	has, _ := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), ret)
	if !has {
		return nil
	}
	return ret
}

func SavePreDownloadInfo(ctx *ctx.Context, info *dao.PreDownloadData) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), info)
}
