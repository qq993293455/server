package match_joiner

import (
	pbSvr "coin-server/common/proto/dungeon_match"
	_ "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/load-test/assert"
	"coin-server/load-test/core"
)

type joinerService struct {
	joinUser map[*testUser]values.MatchRoomId
}

func NewJoiner(ctx *core.RoleContext) core.ILoadTestModule {
	s := &joinerService{
		joinUser: map[*testUser]values.MatchRoomId{},
	}
	for i := 0; i < 3; i++ {
		s.joinUser[newTestUser()] = 0
	}

	return s
}

func (s *joinerService) Process(ctx *core.RoleContext) {
	s.joinRequest(ctx)
	s.leaveRequest(ctx)
}

func (s *joinerService) joinRequest(ctx *core.RoleContext) {
	for role := range s.joinUser {
		req := &pbSvr.DungeonMatch_GetRandomRoomRequest{DungeonId: 1, Num: 5}
		_, res, err := role.ctx.Request(req)
		assert.Nil(role.ctx, err)
		assert.NotNil(role.ctx, res)
		rooms := res.(*pbSvr.DungeonMatch_GetRandomRoomResponse)

		if len(rooms.Rooms) == 0 {
			continue
		}
		req1 := &pbSvr.DungeonMatch_JoinRoomRequest{
			RoomId:    rooms.Rooms[0].RoomId,
			DungeonId: rooms.Rooms[0].DungeonId,
		}
		s.joinUser[role] = values.MatchRoomId(rooms.Rooms[0].RoomId)
		_, res1, err1 := role.ctx.Request(req1)
		if err1 != nil {
			// fmt.Println(err1)
		} else {
			_ = res1.(*pbSvr.DungeonMatch_JoinRoomResponse)
		}
	}
}

func (s *joinerService) leaveRequest(ctx *core.RoleContext) {
	for role, roomId := range s.joinUser {
		if roomId != 0 {
			req := &pbSvr.DungeonMatch_LeaveRoomRequest{
				RoomId:    int64(roomId),
				DungeonId: 1,
			}
			_, res, err := role.ctx.Request(req)
			if err != nil {
				// fmt.Println(err)
			} else {
				_ = res.(*pbSvr.DungeonMatch_LeaveRoomResponse)
			}
		}
	}
}
