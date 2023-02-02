package room_mgr

import (
	"sync"

	"coin-server/common/values"
)

var roomPool *sync.Pool
var publicPool *sync.Pool
var randomRoomPool *sync.Pool
var roleInRoomPool *sync.Pool

func init() {
	roomPool = &sync.Pool{
		New: func() interface{} {
			return &Room{}
		},
	}
	publicPool = &sync.Pool{
		New: func() interface{} {
			return map[values.MatchRoomId]bool{}
		},
	}

	randomRoomPool = &sync.Pool{
		New: func() interface{} {
			return make([]*Room, 0, RoomMaxParty)
		},
	}

	roleInRoomPool = &sync.Pool{
		New: func() interface{} {
			return make([]RoleInRoom, 0, RoomMaxParty)
		},
	}
}
