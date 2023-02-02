package db

import (
	"strconv"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

var serverUser = "server.user:"
var serverRole = "server.role:"

func GetUser(c *ctx.Context, userId string) (*dao.User, *errmsg.ErrMsg) {
	u := &dao.User{UserId: userId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), u)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return u, nil
}

func GetMultiUser(c *ctx.Context, users []*dao.User) *errmsg.ErrMsg {
	ri := make([]orm.RedisInterface, len(users))
	for idx := range ri {
		ri[idx] = users[idx]
	}
	_, err := c.NewOrm().MGetPB(redisclient.GetUserRedis(), ri...)
	if err != nil {
		return err
	}
	return nil
}

func GetUsers(c *ctx.Context, users []string) (map[string]*dao.User, *errmsg.ErrMsg) {
	out := map[string]*dao.User{}
	ri := make([]orm.RedisInterface, 0, len(users))
	for _, v := range users {
		du := &dao.User{UserId: v}
		out[v] = du
		ri = append(ri, du)
	}
	_, err := c.NewOrm().MGetPB(redisclient.GetUserRedis(), ri...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func SaveUser(c *ctx.Context, u *dao.User) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
}

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

func GetRoles(c *ctx.Context, roles []string) (map[string]*dao.Role, *errmsg.ErrMsg) {
	out := map[string]*dao.Role{}
	if len(roles) == 0 {
		return out, nil
	}
	daoRoles := make([]*dao.Role, 0, len(roles))
	for _, v := range roles {
		daoRoles = append(daoRoles, &dao.Role{RoleId: v})
	}
	resp, err := GetMultiRole(c, daoRoles)
	if err != nil {
		return nil, err
	}
	for _, v := range resp {
		out[v.RoleId] = v
	}
	return out, nil
}

func GetMultiRole(c *ctx.Context, roles []*dao.Role) ([]*dao.Role, *errmsg.ErrMsg) {
	ri := make([]orm.RedisInterface, len(roles))
	for idx := range ri {
		ri[idx] = roles[idx]
	}
	notFound, err := c.NewOrm().MGetPB(redisclient.GetUserRedis(), ri...)
	if err != nil {
		return nil, err
	}
	newList := make([]*dao.Role, 0)
	notFoundMap := make(map[int]struct{}, len(notFound))
	for _, i := range notFound {
		notFoundMap[i] = struct{}{}
	}
	for i := range ri {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		newList = append(newList, ri[i].(*dao.Role))
	}
	return newList, nil
}

func SaveRole(c *ctx.Context, u *dao.Role) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
}

func getServerUserKey(serverId values.ServerId) string {
	bs := make([]byte, 0, len(serverUser)+20)
	bs = append(bs, serverUser...)
	bs = strconv.AppendInt(bs, serverId, 10)
	return utils.BytesToString(bs)
}

func GetRoleAttr(ctx *ctx.Context, roleId values.RoleId) ([]*dao.RoleAttr, *errmsg.ErrMsg) {
	attrs := make([]*dao.RoleAttr, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getRoleAttrKey(roleId), &attrs)
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

func GetRoleAttrByType(ctx *ctx.Context, roleId values.RoleId, typ models.AttrBonusType) (*dao.RoleAttr, *errmsg.ErrMsg) {
	attr := &dao.RoleAttr{Typ: typ}
	_, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getRoleAttrKey(roleId), attr)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

func SaveRoleAttr(ctx *ctx.Context, roleId values.RoleId, attr *dao.RoleAttr) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getRoleAttrKey(roleId), attr)
}

func GetRoleSkill(ctx *ctx.Context, roleId values.RoleId) (*dao.RoleSkill, *errmsg.ErrMsg) {
	r := &dao.RoleSkill{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), r)
	if err != nil {
		return nil, err
	}
	if !ok {
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), r)
	}
	return r, nil
}

func SaveRoleSkill(ctx *ctx.Context, skill *dao.RoleSkill) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), skill)
}

func getRoleAttrKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.RoleAttr, values.Hash, roleId)
}

func NameExist(name string) (bool, *errmsg.ErrMsg) {
	query := `SELECT COUNT(*) num FROM roles WHERE nickname=?;`
	var count int
	if err := orm.GetMySQL().Get(&count, query, name); err != nil {
		return false, errmsg.NewErrorDB(err)
	}
	return count > 0, nil
}
