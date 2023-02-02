package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
)

func GetItem(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (*dao.Item, *errmsg.ErrMsg) {
	item := &dao.Item{ItemId: itemId}
	_, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getItemKey(roleId), item)
	if err != nil {
		return nil, err
	}
	return item, err
}

func GetItems(ctx *ctx.Context, roleId values.RoleId, items []*dao.Item) *errmsg.ErrMsg {
	item := make([]orm.RedisInterface, len(items))
	for idx := range item {
		item[idx] = items[idx]
	}
	_, err := ctx.NewOrm().HMGetPB(redisclient.GetUserRedis(), getItemKey(roleId), item)
	if err != nil {
		return err
	}
	return nil
}

func GetItemBag(ctx *ctx.Context, roleId values.RoleId) ([]*dao.Item, *errmsg.ErrMsg) {
	bag := make([]*dao.Item, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getItemKey(roleId), &bag)
	if err != nil {
		return nil, err
	}
	return bag, nil
}

func SaveItem(ctx *ctx.Context, roleId values.RoleId, item *dao.Item) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getItemKey(roleId), item)
	return
}

// func DelItem(ctx *ctx.Context, roleId values.RoleId, item *dao.Item) *errmsg.ErrMsg {
// 	items := []orm.RedisInterface{item}
// 	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getItemKey(roleId), items)
// 	bagLen, err := GetBagLen(ctx, roleId)
// 	if err != nil {
// 		return err
// 	}
// 	bagLen.Length--
// 	return UpdateBagLen(ctx, bagLen)
// }

func SaveManyItem(ctx *ctx.Context, roleId values.RoleId, items []*dao.Item) {
	if len(items) == 0 {
		return
	}
	add := make([]orm.RedisInterface, len(items))
	for idx := range add {
		add[idx] = items[idx]
	}
	ctx.NewOrm().HMSetPB(redisclient.GetUserRedis(), getItemKey(roleId), add)
	return
}

func DelManyItem(ctx *ctx.Context, roleId values.RoleId, items []*dao.Item) {
	if len(items) == 0 {
		return
	}
	del := make([]orm.RedisInterface, len(items))
	for idx := range del {
		del[idx] = items[idx]
	}
	ctx.NewOrm().HDelPB(redisclient.GetUserRedis(), getItemKey(roleId), del)
	// bagLen, err := GetBagLen(ctx, roleId)
	// if err != nil {
	// 	return err
	// }
	// bagLen.Length -= values.Integer(len(items))
	// return UpdateBagLen(ctx, bagLen)
}

func getItemKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.ItemBag, values.Hash, roleId)
}
