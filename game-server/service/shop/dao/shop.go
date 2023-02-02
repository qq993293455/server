package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/game-server/service/shop/values"
)

func GetShop(ctx *ctx.Context) (*values.Shop, *errmsg.ErrMsg) {
	res := &dao.Shop{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.Shop{
			RoleId: ctx.RoleId,
			Data:   &dao.ManualUpdateShop{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return values.NewShop(res), nil
}

func SaveShop(ctx *ctx.Context, i *values.Shop) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), i.ToDao())
}

func GetArenaShop(ctx *ctx.Context) (*values.ArenaShop, *errmsg.ErrMsg) {
	res := &dao.ArenaShop{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.ArenaShop{
			RoleId: ctx.RoleId,
			Data:   &dao.AutoUpdateShop{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return values.NewArenaShop(res), nil
}

func SaveArenaShop(ctx *ctx.Context, i *values.ArenaShop) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), i.ToDao())
}

func GetGuildShop(ctx *ctx.Context) (*values.GuildShop, *errmsg.ErrMsg) {
	res := &dao.GuildShop{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.GuildShop{
			RoleId: ctx.RoleId,
			Data:   &dao.AutoUpdateShop{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return values.NewGuildShop(res), nil
}

func SaveGuildShop(ctx *ctx.Context, i *values.GuildShop) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), i.ToDao())
}

func GetCampShop(ctx *ctx.Context) (*values.CampShop, *errmsg.ErrMsg) {
	res := &dao.CampShop{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.CampShop{
			RoleId: ctx.RoleId,
			Data:   &dao.ManualUpdateShop{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return values.NewCampShop(res), nil
}

func SaveCampShop(ctx *ctx.Context, i *values.CampShop) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), i.ToDao())
}
