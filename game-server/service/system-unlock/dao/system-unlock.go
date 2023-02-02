package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetSysUnlock(ctx *ctx.Context, roleId values.RoleId) (*dao.SystemUnlock, *errmsg.ErrMsg) {
	unlock := &dao.SystemUnlock{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), unlock)
	if err != nil {
		return nil, err
	}
	if unlock.Click == nil {
		unlock.Click = map[int64]bool{}
	}
	if unlock.Unlock == nil {
		unlock.Unlock = map[int64]bool{}
	}
	return unlock, nil
}

func GetMultiSysUnlock(ctx *ctx.Context, roleIds []values.RoleId) (map[values.RoleId]*dao.SystemUnlock, *errmsg.ErrMsg) {
	data := make([]orm.RedisInterface, 0, len(roleIds))
	for _, roleId := range roleIds {
		data = append(data, &dao.SystemUnlock{RoleId: roleId})
	}
	notFound, err := orm.GetOrm(ctx).MGetPB(redisclient.GetUserRedis(), data...)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 {
		notFoundId := make(map[int]struct{})
		for _, id := range notFound {
			notFoundId[id] = struct{}{}
		}
		res := make(map[values.RoleId]*dao.SystemUnlock, 0)
		for idx, v := range data {
			if _, exist := notFoundId[idx]; !exist {
				temp := v.(*dao.SystemUnlock)
				if temp.Unlock == nil {
					temp.Unlock = map[int64]bool{}
				}
				if temp.Click == nil {
					temp.Click = map[int64]bool{}
				}
				res[temp.RoleId] = temp
			} else {
				res[roleIds[idx]] = &dao.SystemUnlock{
					RoleId: roleIds[idx],
					Unlock: map[int64]bool{},
					Click:  map[int64]bool{},
				}
			}
		}
		return res, nil
	}
	res := make(map[values.RoleId]*dao.SystemUnlock, len(data))
	for _, v := range data {
		temp := v.(*dao.SystemUnlock)
		if temp.Unlock == nil {
			temp.Unlock = map[int64]bool{}
		}
		if temp.Click == nil {
			temp.Click = map[int64]bool{}
		}
		res[temp.RoleId] = temp
	}
	return res, nil
}

func SaveSysUnlock(ctx *ctx.Context, unlock *dao.SystemUnlock) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), unlock)
	return
}
