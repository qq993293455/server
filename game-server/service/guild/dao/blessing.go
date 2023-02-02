package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetBlessing(ctx *ctx.Context) (*dao.Blessing, *errmsg.ErrMsg) {
	data := &dao.Blessing{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		data = &dao.Blessing{
			RoleId:    ctx.RoleId,
			Stage:     1,
			Page:      1,
			Activated: make([]values.Integer, 0),
			Queue:     make([]*models.BlessingQueue, 0),
		}
	}
	if data.Activated == nil {
		data.Activated = make([]values.Integer, 0)
	}
	if data.Queue == nil {
		data.Queue = make([]*models.BlessingQueue, 0)
	}
	return data, nil
}

func SaveBlessing(ctx *ctx.Context, data *dao.Blessing) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}

func GetBlessingEffic(ctx *ctx.Context, guildId values.GuildId) (*dao.BlessingEffic, *errmsg.ErrMsg) {
	data := &dao.BlessingEffic{GuildId: guildId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if data.Effic == nil {
		data.Effic = []*models.BlessingEfficItem{}
	}
	return data, nil
}

func SaveBlessingEffic(ctx *ctx.Context, data *dao.BlessingEffic) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}
