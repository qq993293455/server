package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetMainTask(ctx *ctx.Context, roleId values.RoleId) (*dao.MainTask, *errmsg.ErrMsg) {
	ret := &dao.MainTask{RoleId: roleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ret)
	if !ok {
		return nil, nil
	}
	return ret, err
}

func SaveMainTask(ctx *ctx.Context, task *dao.MainTask) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), task)
}

//func GetMainTaskFinish(ctx *ctx.Context, roleId values.RoleId) (*dao.MainTaskFinish, *errmsg.ErrMsg) {
//	ret := &dao.MainTaskFinish{RoleId: roleId}
//	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), ret)
//	if !ok {
//		return nil, nil
//	}
//	if ret.ChapterFinish == nil {
//		ret.ChapterFinish = map[int64]*dao.MainTaskChapterFinish{}
//	}
//	return ret, err
//}
//
//func SaveMainTaskFinish(ctx *ctx.Context, task *dao.MainTaskFinish) {
//	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), task)
//}
