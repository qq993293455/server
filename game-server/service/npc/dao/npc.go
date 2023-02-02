package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetNpcTalk(ctx *ctx.Context, roleId values.RoleId) (*dao.NpcTalk, *errmsg.ErrMsg) {
	talk := &dao.NpcTalk{RoleId: roleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetUserRedis(), talk)
	if err != nil {
		return talk, err
	}
	//if !has {
	//	return nil, nil
	//}
	if talk.TalkReward == nil {
		talk.TalkReward = map[int64]bool{}
	}
	return talk, nil
}

func SaveNpcTalk(ctx *ctx.Context, talk *dao.NpcTalk) {
	ctx.NewOrm().SetPB(redisclient.GetUserRedis(), talk)
	return
}
