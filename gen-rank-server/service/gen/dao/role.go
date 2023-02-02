package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetRole(c *ctx.Context, roleId values.RoleId) (*dao.Role, *errmsg.ErrMsg) {
	r := &dao.Role{RoleId: roleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), r)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return r, nil
}

func GetMultiRole(c *ctx.Context, roles []*dao.Role) *errmsg.ErrMsg {
	ri := make([]orm.RedisInterface, len(roles))
	for idx := range ri {
		ri[idx] = roles[idx]
	}
	_, err := c.NewOrm().MGetPBInSlot(redisclient.GetUserRedis(), ri...)
	if err != nil {
		return err
	}
	return nil
}

func SaveRoleRank(ctx *ctx.Context, rank *dao.RoleRank) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), rank)
}
