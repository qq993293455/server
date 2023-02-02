package dao

import (
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/rule"
)

func Get(c *ctx.Context, roleId string) (*dao.Formation, bool, *errmsg.ErrMsg) {
	res := &dao.Formation{RoleId: roleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, false, errmsg.NewErrorDB(err)
	}
	if !ok {
		return nil, false, nil
	}
	return res, true, nil
}

func Create(c *ctx.Context, roleId string, heroId int64, heroOriginId int64, heroId1 int64, heroOriginId1 int64) *dao.Formation {
	res := &dao.Formation{RoleId: roleId, DefaultIndex: 0}
	formationDefault, ok := rule.MustGetReader(c).KeyValue.GetInt64("FormationDefault")
	if !ok || formationDefault <= 0 {
		formationDefault = 1
	}

	for i := int64(0); i < formationDefault; i++ {
		res.Assembles = append(res.Assembles, &models.Assemble{
			Hero_0: 0,
			Hero_1: 0,
			Name:   fmt.Sprintf("编队%d", i+1),
		})
	}
	res.Assembles[0].Hero_0 = heroId
	res.Assembles[0].Hero_1 = heroId1
	res.Assembles[0].HeroOrigin_0 = heroOriginId
	res.Assembles[0].HeroOrigin_1 = heroOriginId1

	c.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	return res
}

func GetMulti(c *ctx.Context, roleIds []string) (map[string]*dao.Formation, *errmsg.ErrMsg) {
	out := make(map[string]*dao.Formation, len(roleIds))
	if len(roleIds) == 0 {
		return out, nil
	}
	in := make([]orm.RedisInterface, 0, len(roleIds))
	for _, v := range roleIds {
		f := &dao.Formation{RoleId: v}
		out[v] = f
		in = append(in, f)
	}

	_, err := c.NewOrm().MGetPB(redisclient.GetDefaultRedis(), in...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func Save(c *ctx.Context, f *dao.Formation) {
	c.NewOrm().SetPB(redisclient.GetDefaultRedis(), f)
}
