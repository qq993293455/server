package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetTitleRewards(c *ctx.Context) (*dao.TitleRewards, *errmsg.ErrMsg) {
	d := &dao.TitleRewards{RoleId: c.RoleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), d)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return d, nil
}

func SaveTitleRewards(c *ctx.Context, d *dao.TitleRewards) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), d)
}
