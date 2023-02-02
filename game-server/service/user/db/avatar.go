package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
)

func GetOwnAvatar(c *ctx.Context) (*dao.RoleOwnAvatar, *errmsg.ErrMsg) {
	d := &dao.RoleOwnAvatar{RoleId: c.RoleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), d)
	if err != nil {
		return nil, err
	}
	if !ok {
		d.OwnAvatar = map[int64]*models.Avatar{}
	}
	return d, nil
}

func SaveOwnAvatar(c *ctx.Context, d *dao.RoleOwnAvatar) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), d)
}
