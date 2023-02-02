package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetAtlas(ctx *ctx.Context, roleId values.RoleId) ([]*dao.Atlas, *errmsg.ErrMsg) {
	atlas := make([]*dao.Atlas, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getAtlasKey(roleId), &atlas)
	if err != nil {
		return nil, err
	}
	return atlas, nil
}

func GetAtlasByType(ctx *ctx.Context, roleId values.RoleId, typ models.AtlasType) (*dao.Atlas, *errmsg.ErrMsg) {
	atlas := &dao.Atlas{AtlasType: typ}
	has, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getAtlasKey(roleId), atlas)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return atlas, nil
}

func SaveAtlas(ctx *ctx.Context, roleId values.RoleId, atlas *dao.Atlas) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getAtlasKey(roleId), atlas)
	return
}

func getAtlasKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Atlas, values.Hash, roleId)
}
