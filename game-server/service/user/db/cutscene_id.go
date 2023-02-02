package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetCutScene(c *ctx.Context, roleId values.RoleId) (*dao.CutSceneId, *errmsg.ErrMsg) {
	u := &dao.CutSceneId{RoleId: roleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), u)
	if err != nil {
		return nil, err
	}
	if !ok {
		c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
		return u, nil
	}
	return u, nil
}

func SaveCutScene(c *ctx.Context, u *dao.CutSceneId) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
}
