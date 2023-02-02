package map_event

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	servicepb "coin-server/common/proto/service"
	"coin-server/game-server/event"
)

func (s *Service) HandleEventFinished(ctx *ctx.Context, d *event.MapEventFinished) *errmsg.ErrMsg {
	ctx.PushMessageToRole(ctx.RoleId, &servicepb.MapEvent_EventFinishPush{
		EventId:   d.EventId,
		IsSuccess: d.IsSuccess,
		Rewards:   d.Rewards,
		StoryId:   d.StoryId,
		Piece:     d.Piece,
		MapId:     d.MapId,
		Ratio:     d.Ratio,
	})
	return nil
}

func (s *Service) HandleRoleLoginEvent(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	err := s.unlockMapEvent(ctx)
	if err != nil {
		return err
	}
	return nil
}
