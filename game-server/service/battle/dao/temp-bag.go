package dao

import (
	"errors"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	mapdata "coin-server/common/map-data"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/common/timer"
	"coin-server/common/values"
	"coin-server/rule"
)

func GetTempBag(ctx *ctx.Context, id values.RoleId) (*dao.RoleTempBag, *errmsg.ErrMsg) {
	data := &dao.RoleTempBag{RoleId: id}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		initTempBag(data)
		SaveTempBag(ctx, data)
		return data, nil
	}
	if data.TempBag.Items == nil {
		data.TempBag.Items = map[int64]int64{}
	}
	return data, nil
}

func SaveTempBag(ctx *ctx.Context, data *dao.RoleTempBag) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func BatchGetTempBag(ctx *ctx.Context, ids []values.RoleId) ([]*dao.RoleTempBag, *errmsg.ErrMsg) {
	data := make([]*dao.RoleTempBag, 0, len(ids))
	out := make([]orm.RedisInterface, 0, len(ids))
	for _, id := range ids {
		tb := &dao.RoleTempBag{RoleId: id}
		data = append(data, tb)
		out = append(out, tb)
	}
	notFoundIdx, err := ctx.NewOrm().MGetPB(redisclient.GetUserRedis(), out...)
	if err != nil {
		return nil, err
	}

	if len(notFoundIdx) > 0 {
		init := make([]orm.RedisInterface, 0, len(notFoundIdx))
		for _, i := range notFoundIdx {
			initTempBag(data[i])
			init = append(init, data[i])
		}
		ctx.NewOrm().MSetPB(redisclient.GetUserRedis(), init)
	}

	return data, nil
}

func BatchSaveTempBag(ctx *ctx.Context, data []*dao.RoleTempBag) {
	set := make([]orm.RedisInterface, 0, len(data))
	for _, d := range data {
		set = append(set, d)
	}
	ctx.NewOrm().MSetPB(redisclient.GetUserRedis(), set)
}

func initTempBag(bag *dao.RoleTempBag) {
	reader := rule.MustGetReader(nil)
	bag.MapId = mapdata.GetDefaultMapId(nil)
	bag.ProfitUpper = reader.TempBag.List()[0].ProfitUpper * 3600
	bag.LastCalcTime = timer.Unix()
	bag.TempBag = &models.TempBag{
		StartTime: timer.Now().Unix(),
		Items:     map[int64]int64{},
		ExpProfit: 0,
	}
	bag.ExpProfitAdd = 0     // 初始加成
	bag.ExpProfitPercent = 0 // 初始百分比加成
	bag.ExpProfitBase = 0
	cap_, ok := rule.MustGetReader(nil).KeyValue.GetInt64("TempBagCap")
	if !ok {
		panic(errors.New("KeyValue TempBagCap not found"))
	}
	bag.CapLimit = cap_
	bag.BagSize = 0
}
