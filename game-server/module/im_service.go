package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/values/enum/Notice"
)

type ImService interface {
	SendNotice(ctx *ctx.Context, parseType int, noticeId Notice.Enum, args ...any) *errmsg.ErrMsg // 发跑马灯公告
}
