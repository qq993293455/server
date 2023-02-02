package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetRecentChat(ctx *ctx.Context, roleId values.RoleId) (*dao.RecentChatTars, *errmsg.ErrMsg) {
	res := &dao.RecentChatTars{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		data := &dao.RecentChatTars{
			RoleId:     roleId,
			TarRoleIds: map[string]int64{},
		}
		ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
		return data, nil
	}
	return handleNilMap(res), nil
}

func SaveRecentChat(ctx *ctx.Context, data *dao.RecentChatTars) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), data)
}

func handleNilMap(data *dao.RecentChatTars) *dao.RecentChatTars {
	if data.TarRoleIds == nil {
		data.TarRoleIds = map[string]int64{}
	}
	return data
}
