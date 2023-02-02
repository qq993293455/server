package npc

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/game-server/event"
)

func (s *Service) HandleTalkEvent(ctx *ctx.Context, d *event.Talk) *errmsg.ErrMsg {
	return nil
}
