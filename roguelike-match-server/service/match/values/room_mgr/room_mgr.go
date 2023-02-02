package room_mgr

import (
	"math"
	"math/rand"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/rule"
)

const (
	roomMask    = 0b111 // 255
	roomSlotNum = roomMask + 1

	roleMask    = 0b11111 // 2047
	roleSlotNum = roleMask + 1

	UnExistRoom      values.MatchRoomId = math.MaxUint64
	UnExistRoguelike values.RoguelikeId = math.MaxUint64
)

type RoomMgr struct {
	roomSlot         [][roomSlotNum]*roomSlot // 创建出来的房间
	roleSlot         [roleSlotNum]*roleSlot   // 参与到匹配系统中的角色
	CloseAfter       int64
	OneDayStartLimit int64
}

func NewRoomMgr() *RoomMgr {
	reader := rule.MustGetReader(&ctx.Context{})
	roguelikeLen := reader.RoguelikeDungeon.Len()
	rm := &RoomMgr{
		roomSlot: make([][roomSlotNum]*roomSlot, roguelikeLen),
		roleSlot: [roleSlotNum]*roleSlot{},
	}
	for idx := 0; idx < roleSlotNum; idx++ {
		rm.roleSlot[idx] = newRoleSlot()
	}
	for idx := range rm.roomSlot {
		for i := 0; i < roomSlotNum; i++ {
			rm.roomSlot[idx][i] = newRoomSlot()
		}
	}
	rm.CloseAfter = 1800
	rm.OneDayStartLimit = 5
	t, ok := reader.KeyValue.GetInt64("RougueDungeonsTeamTime")
	if ok {
		rm.CloseAfter = t
	}
	o, ok := reader.KeyValue.GetInt64("RougueDungeonsNum")
	if ok {
		rm.OneDayStartLimit = o
	}
	return rm
}

func hashRole(id string) int {
	return int(utils.Base34DecodeString(id) & roleMask)
}

func hashInt(id uint64) int {
	return int(id & roomMask)
}

func (rm *RoomMgr) CreateRoom(ctx *ctx.Context, ownerId values.RoleId, roguelikeId values.RoguelikeId, serverId values.ServerId) (*Room, *errmsg.ErrMsg) {
	roguelikeRule, ok := rule.MustGetReader(ctx).RoguelikeDungeon.GetRoguelikeDungeonById(int64(roguelikeId))
	if !ok {
		return nil, errmsg.NewErrRoguelikeNotExist()
	}
	if !rm.isLegal(roguelikeId) {
		return nil, errmsg.NewErrRoguelikeNotExist()
	}
	roleHash := hashRole(ownerId)
	if rm.roleSlot[roleHash].exist(ownerId) {
		return nil, errmsg.NewErrRLOnlyCanJoinOneRoom()
	}
	r, err := newRoom(ownerId, roguelikeId, int32(roguelikeRule.DungeonPlayerNum))
	if err != nil {
		return nil, err
	}
	rm.roleSlot[hashRole(ownerId)].set(ownerId, r.roomId, r.roguelikeId, serverId)
	rm.roomSlot[roguelikeId-1][hashInt(r.RoomId())].set(r)
	return r, nil
}

func (rm *RoomMgr) GetRoom(roomId values.MatchRoomId, roguelikeId values.RoguelikeId) *Room {
	if !rm.isLegal(roguelikeId) {
		return nil
	}
	return rm.roomSlot[roguelikeId-1][hashInt(roomId)].get(roomId)
}

func (rm *RoomMgr) GetRoleRoom(roleId values.RoleId) (values.MatchRoomId, values.RoguelikeId) {
	hKey := hashRole(roleId)
	if rm.roleSlot[hKey].exist(roleId) {
		r := rm.roleSlot[hKey].get(roleId)
		return r[roomIdx], r[roguelikeIdx]
	}
	return UnExistRoom, UnExistRoguelike
}

func (rm *RoomMgr) GetRoleSvcId(roleId values.RoleId) values.ServerId {
	hKey := hashRole(roleId)
	if rm.roleSlot[hKey].exist(roleId) {
		return values.ServerId(rm.roleSlot[hKey].get(roleId)[serverIdx])
	}
	return -1
}

func (rm *RoomMgr) Join(roomId values.MatchRoomId, roguelikeId values.RoguelikeId, roleId values.RoleId, serverId values.ServerId) *errmsg.ErrMsg {
	if !rm.isLegal(roguelikeId) {
		return errmsg.NewErrRoguelikeNotExist()
	}
	roleHash := hashRole(roleId)
	if rm.roleSlot[roleHash].exist(roleId) {
		return errmsg.NewErrOnlyCanJoinOneRoom()
	}
	if err := rm.roomSlot[roguelikeId-1][hashInt(roomId)].join(roomId, roleId); err != nil {
		return err
	}
	rm.roleSlot[roleHash].set(roleId, roomId, roguelikeId, serverId)
	return nil
}

func (rm *RoomMgr) JoinRobot(roomId values.MatchRoomId, roguelikeId values.RoguelikeId, roleId values.RoleId, configId values.Integer) *errmsg.ErrMsg {
	if !rm.isLegal(roguelikeId) {
		return errmsg.NewErrRoguelikeNotExist()
	}
	if err := rm.roomSlot[roguelikeId-1][hashInt(roomId)].joinRobot(roomId, roleId, configId); err != nil {
		return err
	}
	return nil
}

func (rm *RoomMgr) Leave(roomId values.MatchRoomId, roleId values.RoleId) *errmsg.ErrMsg {
	roleHash := hashRole(roleId)
	ok := rm.roleSlot[roleHash].exist(roleId)
	if !ok {
		return errmsg.NewErrNotIn()
	}
	data := rm.roleSlot[roleHash].get(roleId)
	tarR, tarRoguelike := data[roomIdx], data[roguelikeIdx]
	if tarR != roomId || tarR == 0 {
		return errmsg.NewErrTarNotInThisRoom()
	}
	if !rm.isLegal(tarRoguelike) {
		return errmsg.NewErrRoguelikeNotExist()
	}
	if err := rm.roomSlot[tarRoguelike-1][hashInt(roomId)].leave(roomId, roleId); err != nil {
		return err
	}
	rm.roleSlot[roleHash].delete(roleId)
	return nil
}

func (rm *RoomMgr) DelRoom(roomId values.MatchRoomId, roguelikeId values.RoguelikeId) *errmsg.ErrMsg {
	if !rm.isLegal(roguelikeId) {
		return errmsg.NewErrRoguelikeNotExist()
	}
	r := rm.roomSlot[roguelikeId-1][hashInt(roomId)].get(roomId)
	if r == nil {
		return errmsg.NewErrRoomNotExist()
	}
	for _, party := range r.party {
		rm.roleSlot[hashRole(party)].delete(party)
	}
	if err := rm.roomSlot[roguelikeId-1][hashInt(r.RoomId())].delete(r.RoomId()); err != nil {
		return err
	}
	return nil
}

func (rm *RoomMgr) SetPub(roomId values.MatchRoomId, roguelikeId values.RoguelikeId, isOpen bool) *errmsg.ErrMsg {
	if !rm.isLegal(roguelikeId) {
		return errmsg.NewErrRoguelikeNotExist()
	}
	rm.roomSlot[roguelikeId-1][hashInt(roomId)].setPub(roomId, isOpen)
	return nil
}

func (rm *RoomMgr) isLegal(roguelikeId values.RoguelikeId) bool {
	return roguelikeId > 0 && roguelikeId <= uint64(len(rm.roomSlot))
}

func (rm *RoomMgr) GetRandomRooms(roguelikeId values.RoguelikeId, num int, roomTime int64) ([]*Room, *errmsg.ErrMsg) {
	if !rm.isLegal(roguelikeId) {
		return nil, errmsg.NewErrRoguelikeNotExist()
	}
	res := randomRoomPool.Get().([]*Room)
	used := publicPool.Get().(map[values.MatchRoomId]bool)
	searchCnt, i := 0, 0
	for i < num && searchCnt < maxSearchCnt {
		slotIdx := rand.Intn(roomSlotNum)
		r := rm.roomSlot[roguelikeId-1][slotIdx].getRandomRoom(roomTime)
		if r != nil && !used[r.roomId] {
			res = append(res, r)
			used[r.roomId] = true
			i++
		}

		searchCnt++
	}
	for key := range used {
		delete(used, key)
	}
	publicPool.Put(used)
	return res, nil
}

func (rm *RoomMgr) PutPubRooms(rooms []*Room) {
	rooms = rooms[:0]
	randomRoomPool.Put(rooms)
}
