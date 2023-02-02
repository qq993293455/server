package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetEnchant(ctx *ctx.Context) (*dao.Enchant, *errmsg.ErrMsg) {
	enchant := &dao.Enchant{RoleId: ctx.RoleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), enchant)
	if err != nil {
		return nil, err
	}
	return enchant, nil
}

func SaveEnchant(ctx *ctx.Context, enchant *dao.Enchant) *errmsg.ErrMsg {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), enchant)
	return nil
}

func NewEnchant(roleId values.RoleId, heroId values.HeroId, slotId values.Integer, affix *models.Affix) *dao.Enchant {
	return &dao.Enchant{
		RoleId: roleId,
		HeroId: heroId,
		SlotId: slotId,
		Affix:  affix,
	}
}

func ResetEnchant(roleId values.RoleId) *dao.Enchant {
	return &dao.Enchant{
		RoleId: roleId,
	}
}
