package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetExtraSkill(c *ctx.Context, roleId values.RoleId) (*dao.ExtraSkillTypCnt, *errmsg.ErrMsg) {
	u := &dao.ExtraSkillTypCnt{RoleId: roleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), u)
	if err != nil {
		return nil, err
	}
	if !ok {
		u.Data = map[int64]*dao.ExtraSkillTypCntDetail{}
		c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
		return u, nil
	}
	if u.Data == nil {
		u.Data = map[int64]*dao.ExtraSkillTypCntDetail{}
	}
	return u, nil
}

func SaveExtraSkill(c *ctx.Context, u *dao.ExtraSkillTypCnt) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
}
