package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

type Hero struct {
	values.RoleId
}

func NewHero(id values.RoleId) *Hero {
	return &Hero{
		RoleId: id,
	}
}

func (h *Hero) Get(ctx *ctx.Context) ([]*dao.Hero, *errmsg.ErrMsg) {
	list := make([]*dao.Hero, 0)
	if err := ctx.NewOrm().HGetAll(redisclient.GetDefaultRedis(), h.getKey(h.RoleId), &list); err != nil {
		return nil, err
	}
	for i := range list {
		NilInit(list[i])
	}
	return list, nil
}

func (h *Hero) GetOne(ctx *ctx.Context, id values.HeroId) (*dao.Hero, bool, *errmsg.ErrMsg) {
	hero := &dao.Hero{
		Id: id,
	}
	ok, err := ctx.NewOrm().HGetPB(redisclient.GetDefaultRedis(), h.getKey(h.RoleId), hero)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, err
	}
	NilInit(hero)
	return hero, true, nil
}

func (h *Hero) GetSome(ctx *ctx.Context, ids []int64) ([]*dao.Hero, *errmsg.ErrMsg) {
	list := make([]*dao.Hero, len(ids))
	for i, v := range ids {
		list[i] = &dao.Hero{Id: v}
	}
	in := make([]orm.RedisInterface, 0, len(ids))
	for _, v := range list {
		in = append(in, v)
	}

	notFound, err := ctx.NewOrm().HMGetPB(redisclient.GetDefaultRedis(), h.getKey(h.RoleId), in)
	if err != nil {
		return nil, err
	}
	for i := range list {
		NilInit(list[i])
	}
	if len(notFound) == 0 {
		return list, nil
	}
	return nil, errmsg.NewErrHeroNotFound()
}

func (h *Hero) Save(ctx *ctx.Context, heroes ...*dao.Hero) *errmsg.ErrMsg {
	if len(heroes) == 0 {
		return nil
	}
	ins := make([]orm.RedisInterface, 0, len(heroes))
	for _, hero := range heroes {
		ins = append(ins, hero)
	}
	ctx.NewOrm().HMSetPB(redisclient.GetDefaultRedis(), h.getKey(h.RoleId), ins)
	return nil
}

func (h *Hero) getKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.Hero, values.Hash, roleId)
}

func NilInit(hero *dao.Hero) {
	if hero.Skill == nil {
		hero.Skill = map[int64]*dao.SkillLevel{}
	}
	if hero.Attrs == nil {
		hero.Attrs = map[int64]*models.HeroAttr{}
	}
	if hero.EquipSlot == nil {
		hero.EquipSlot = map[int64]*models.HeroEquipSlot{}
	}
	if hero.CombatValue == nil {
		hero.CombatValue = &models.CombatValue{}
	}
	if hero.CombatValue.Details == nil {
		hero.CombatValue.Details = map[int64]int64{}
	}
	if hero.Buff == nil {
		hero.Buff = map[int64]*models.HeroBuffItem{}
	}
	for i := range hero.Buff {
		if hero.Buff[i] == nil {
			hero.Buff[i] = &models.HeroBuffItem{}
		}
	}
	if hero.SoulContract == nil {
		hero.SoulContract = &models.SoulContract{}
	}
	if hero.Resonance == nil {
		hero.Resonance = map[int64]models.ResonanceStatus{}
	}
	if hero.Fashion == nil {
		hero.Fashion = &models.HeroFashion{}
	}
	if hero.Fashion.Data == nil {
		hero.Fashion.Data = map[int64]int64{}
	}
}
