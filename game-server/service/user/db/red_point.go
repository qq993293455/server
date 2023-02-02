package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetReadPoint(c *ctx.Context, roleId values.RoleId) (*dao.ReadPoint, *errmsg.ErrMsg) {
	u := &dao.ReadPoint{RoleId: roleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), u)
	if err != nil {
		return nil, err
	}
	if !ok {
		u.RedPoints = map[string]int64{}
		c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
		return u, nil
	}
	if u.RedPoints == nil {
		u.RedPoints = map[string]int64{}
	}
	return u, nil
}

func SaveRedPoint(c *ctx.Context, u *dao.ReadPoint) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
}
