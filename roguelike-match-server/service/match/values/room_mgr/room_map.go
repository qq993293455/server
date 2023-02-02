package room_mgr

import (
	"sync"

	"coin-server/common/errmsg"
	"coin-server/common/values"
)

type roomSlot struct {
	*sync.RWMutex
	roomMap map[values.MatchRoomId]*Room
	pubList *publicList
}

func newRoomSlot() *roomSlot {
	return &roomSlot{
		&sync.RWMutex{},
		map[values.MatchRoomId]*Room{},
		newPublicList(),
	}
}

func (rs *roomSlot) get(id values.MatchRoomId) *Room {
	rs.RLock()
	defer rs.RUnlock()
	return rs.roomMap[id]
}

func (rs *roomSlot) set(room *Room) {
	rs.Lock()
	defer rs.Unlock()
	rs.roomMap[room.roomId] = room
	if room.IsOpen() {
		rs.pubList.pub(room)
	}
}

func (rs *roomSlot) join(roomId values.MatchRoomId, roleId values.RoleId) *errmsg.ErrMsg {
	rs.Lock()
	defer rs.Unlock()
	if r, exist := rs.roomMap[roomId]; exist {
		return r.join(roleId)
	}
	return errmsg.NewErrRoomNotExist()
}

func (rs *roomSlot) joinRobot(roomId values.MatchRoomId, roleId values.RoleId, configId values.Integer) *errmsg.ErrMsg {
	rs.Lock()
	defer rs.Unlock()
	if r, exist := rs.roomMap[roomId]; exist {
		return r.joinRobot(roleId, configId)
	}
	return errmsg.NewErrRoomNotExist()
}

func (rs *roomSlot) leave(roomId values.MatchRoomId, roleId values.RoleId) *errmsg.ErrMsg {
	rs.Lock()
	defer rs.Unlock()
	if r, exist := rs.roomMap[roomId]; exist {
		return r.leave(roleId)
	}
	return errmsg.NewErrRoomNotExist()
}

func (rs *roomSlot) delete(id values.MatchRoomId) *errmsg.ErrMsg {
	rs.Lock()
	defer rs.Unlock()
	if r, exist := rs.roomMap[id]; exist {
		delete(rs.roomMap, id)
		if r.IsOpen() {
			rs.pubList.off(r)
		}
		r.reset()
		roomPool.Put(r)
		return nil
	}
	return errmsg.NewErrRoomNotExist()
}

func (rs *roomSlot) setPub(id values.MatchRoomId, isOpen bool) {
	rs.Lock()
	defer rs.Unlock()
	if r, exist := rs.roomMap[id]; exist {
		oldPub := r.IsOpen()
		if isOpen {
			r.Open()
		} else {
			r.Close()
		}
		if oldPub && !r.IsOpen() {
			rs.pubList.off(r)
		}
		if !oldPub && r.IsOpen() {
			rs.pubList.pub(r)
		}
	}
}

func (rs *roomSlot) getRandomRoom(roomTime int64) *Room {
	rs.RLock()
	defer rs.RUnlock()
	return rs.pubList.getRandomRoom(roomTime)
}

const (
	roomIdx = iota
	roguelikeIdx
	serverIdx
)

type roleSlot struct {
	*sync.RWMutex
	roleMap map[values.RoleId]*[3]uint64
}

func newRoleSlot() *roleSlot {
	return &roleSlot{
		&sync.RWMutex{},
		map[values.RoleId]*[3]uint64{},
	}
}

func (rs *roleSlot) exist(roleId values.RoleId) bool {
	rs.RLock()
	defer rs.RUnlock()
	_, exist := rs.roleMap[roleId]
	return exist
}

func (rs *roleSlot) get(roleId values.RoleId) *[3]uint64 {
	rs.RLock()
	defer rs.RUnlock()
	return rs.roleMap[roleId]
}

func (rs *roleSlot) set(roleId values.RoleId, roomId values.MatchRoomId, roguelikeId values.RoguelikeId, serverId values.ServerId) {
	rs.Lock()
	defer rs.Unlock()
	rs.roleMap[roleId] = &[3]uint64{roomId, roguelikeId, uint64(serverId)}
}

func (rs *roleSlot) delete(id values.RoleId) {
	rs.Lock()
	defer rs.Unlock()
	delete(rs.roleMap, id)
}
