package room_mgr

import (
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/utils"
	"coin-server/common/values"
)

const (
	RoomMaxParty = 3
)

type RoleInRoom struct {
	RoleId   values.RoleId
	ConfigId values.Integer
	IsReady  bool
}

type Room struct {
	// 房间Id
	roomId values.MatchRoomId
	// 地图Id
	roguelikeId values.RoguelikeId
	// 该副本最大队员数
	maxParty int32
	// 当前队员数
	currParty int32
	// 开放设置
	isOpen bool
	// 公开序列中的idx
	pubIdx int64
	// 队员
	party [RoomMaxParty]values.RoleId
	// 准备状态
	pStatus [RoomMaxParty]partyStatus
	// 英雄
	heroes [RoomMaxParty]values.Integer
	// 卡组
	cards [RoomMaxParty]values.Integer
	// 状态
	status models.RoguelikeStatus
	// 战斗服务id
	battleId values.Integer
	// 开始时间
	startTime int64
	// 更新时间，用来自动关闭房间
	updateAt int64
	// 战斗力需要
	combatNeed values.Integer
}

func newRoom(ownerId values.RoleId, roguelikeId values.RoguelikeId, partyCnt int32) (*Room, *errmsg.ErrMsg) {
	if partyCnt > RoomMaxParty {
		partyCnt = RoomMaxParty
	}
	r := roomPool.Get().(*Room)
	now := uint64(timer.Now().Unix())
	r.roomId = utils.Base34DecodeString(ownerId)*10000 + now - now/10000*10000
	r.roguelikeId = roguelikeId
	r.isOpen = true
	r.currParty = 1
	r.maxParty = partyCnt
	r.pubIdx = -1
	r.party[0] = ownerId
	r.status = models.RoguelikeStatus_RoguelikeStatusStatusMatch
	r.updateAt = int64(now)
	r.combatNeed = 0
	return r, nil
}

func (r *Room) RoomId() values.MatchRoomId {
	return r.roomId
}

func (r *Room) RoomLen() int {
	return int(r.currParty)
}

func (r *Room) RoguelikeId() values.RoguelikeId {
	return r.roguelikeId
}

func (r *Room) OwnerId() values.RoleId {
	return r.party[0]
}

func (r *Room) CombatNeed() values.Integer {
	return r.combatNeed
}

func (r *Room) SetCombatNeed(combat values.Integer) {
	r.combatNeed = combat
}

func (r *Room) ChangeOwner(roleId values.RoleId) *errmsg.ErrMsg {
	for idx := int32(1); idx < r.currParty; idx++ {
		if r.party[idx] == roleId {
			if idx == 0 {
				return errmsg.NewErrRLAlreadyOwner()
			}
			r.party[idx], r.party[0] = r.party[0], r.party[idx]
			r.pStatus[idx], r.pStatus[0] = r.pStatus[0], r.pStatus[idx]
			r.heroes[idx], r.heroes[0] = r.heroes[0], r.heroes[idx]
			r.cards[idx], r.cards[0] = r.cards[0], r.cards[idx]
			return nil
		}
	}
	return errmsg.NewErrRLNotIn()
}

func (r *Room) CheckClose(after values.Integer) bool {
	return timer.Now().Unix()-r.updateAt > after
}

func (r *Room) BattleId() values.Integer {
	return r.battleId
}

func (r *Room) IsOpen() bool {
	return r.isOpen
}

func (r *Room) CloseAt(after int64) int64 {
	return r.updateAt + after
}

func (r *Room) GetParty() []values.RoleId {
	res := make([]values.RoleId, 0, r.currParty)
	for idx := 0; idx < int(r.currParty); idx++ {
		if r.party[idx] != "" && !r.pStatus[idx].IsRobot() {
			res = append(res, r.party[idx])
		}
	}
	return res
}

func (r *Room) GetRobots() []values.RoleId {
	res := make([]values.RoleId, 0, r.currParty)
	for idx := 0; idx < int(r.currParty); idx++ {
		if r.party[idx] != "" && r.pStatus[idx].IsRobot() {
			res = append(res, r.party[idx])
		}
	}
	return res
}

func (r *Room) GetPartyInfo() []*models.RoguelikeRoleInfo {
	res := make([]*models.RoguelikeRoleInfo, 0, r.currParty)
	for idx := int32(0); idx < r.currParty; idx++ {
		res = append(res, &models.RoguelikeRoleInfo{
			RoleId:   r.party[idx],
			ConfigId: r.heroes[idx],
			CardId:   r.cards[idx],
			IsReady:  r.pStatus[idx].IsReady(),
			IsRobot:  r.pStatus[idx].IsRobot(),
		})
	}
	return res
}

func (r *Room) IsAllReady() bool {
	for idx := int32(1); idx < r.currParty; idx++ {
		if !r.pStatus[idx].IsReady() {
			return false
		}
	}
	return true
}

func (r *Room) Ready(roleId values.RoleId, b bool) *errmsg.ErrMsg {
	for idx, rid := range r.party {
		if rid == roleId {
			if b && !r.pStatus[idx].IsReady() {
				r.pStatus[idx] += IsPartyReady
			}
			if !b && r.pStatus[idx].IsReady() {
				r.pStatus[idx] -= IsPartyReady
			}
			return nil
		}
	}
	return errmsg.NewErrRLNotIn()
}

func (r *Room) ChooseHero(roleId values.RoleId, configId, cardId values.Integer) *errmsg.ErrMsg {
	for idx, rid := range r.party {
		if rid == roleId {
			r.heroes[idx] = configId
			r.cards[idx] = cardId
			return nil
		}
	}
	return errmsg.NewErrRLNotIn()
}

func (r *Room) CardMap() map[values.RoleId]values.Integer {
	res := map[values.RoleId]values.Integer{}
	for i := int32(0); i < r.currParty; i++ {
		res[r.party[i]] = r.cards[i]
	}
	return res
}

func (r *Room) IsStarted() bool {
	return r.status == models.RoguelikeStatus_RoguelikeStatusStatusFighting
}

func (r *Room) Status() models.RoguelikeStatus {
	return r.status
}

func (r *Room) StartFight(battleId values.Integer) {
	r.status = models.RoguelikeStatus_RoguelikeStatusStatusFighting
	r.battleId = battleId
	now := timer.Unix()
	r.startTime = now
	r.updateAt = now
}

func (r *Room) StartTime() int64 {
	return r.startTime
}

func (r *Room) StopFight() {
	r.status = models.RoguelikeStatus_RoguelikeStatusStatusMatch
	r.battleId = 0
	r.startTime = 0
	r.updateAt = timer.Unix()
	for i := 0; i < int(r.currParty); i++ {
		if !r.pStatus[i].IsRobot() {
			// 玩家
			if r.pStatus[i].IsReady() {
				r.pStatus[i] -= IsPartyReady
			}
		} else {
			// 机器人
			for curr := i; curr < int(r.currParty)-1; curr++ {
				r.party[curr] = r.party[curr+1]
				r.pStatus[curr] = r.pStatus[curr+1]
				r.heroes[curr] = r.heroes[curr+1]
				r.cards[curr] = r.cards[curr+1]
			}
			r.party[r.currParty-1] = ""
			r.pStatus[r.currParty-1] = 0
			r.heroes[r.currParty-1] = 0
			r.cards[r.currParty-1] = 0
			r.currParty--
			i--
		}
	}
}

func (r *Room) Open() {
	r.isOpen = true
}

func (r *Room) Close() {
	r.isOpen = false
}

func (r *Room) isFull() bool {
	return r.currParty >= r.maxParty
}

func (r *Room) Remain() int32 {
	return r.maxParty - r.currParty
}

func (r *Room) join(roleId values.RoleId) *errmsg.ErrMsg {
	if r.isFull() {
		return errmsg.NewErrRoomFull()
	}
	r.party[r.currParty] = roleId
	r.heroes[r.currParty] = 0
	r.cards[r.currParty] = 0
	r.pStatus[r.currParty] = 0
	r.currParty++
	return nil
}

func (r *Room) joinRobot(roleId values.RoleId, configId values.Integer) *errmsg.ErrMsg {
	if r.isFull() {
		return errmsg.NewErrRoomFull()
	}
	r.party[r.currParty] = roleId
	r.heroes[r.currParty] = configId
	r.pStatus[r.currParty] = IsPartyRobot + IsPartyReady
	r.cards[r.currParty] = 1
	r.currParty++
	return nil
}

func (r *Room) leave(roleId values.RoleId) *errmsg.ErrMsg {
	for i := 0; i < int(r.currParty); i++ {
		if roleId == r.party[i] {
			for curr := i; curr < int(r.currParty)-1; curr++ {
				r.party[curr] = r.party[curr+1]
				r.pStatus[curr] = r.pStatus[curr+1]
				r.heroes[curr] = r.heroes[curr+1]
				r.cards[curr] = r.cards[curr+1]
			}
			r.party[r.currParty-1] = ""
			r.pStatus[r.currParty-1] = 0
			r.heroes[r.currParty-1] = 0
			r.cards[r.currParty-1] = 0
			r.currParty--
			return nil
		}
	}
	return errmsg.NewErrNotIn()
}

func (r *Room) KickRobot(roleId values.RoleId) *errmsg.ErrMsg {
	if r.IsRobot(roleId) {
		r.leave(roleId)
	}
	return nil
}

func (r *Room) IsRobot(roleId values.RoleId) bool {
	for idx, val := range r.party {
		if val == roleId {
			return r.pStatus[idx].IsRobot()
		}
	}
	return false
}

func (r *Room) IsEmpty() bool {
	if r.currParty == 0 {
		return true
	}
	for idx := 0; idx < int(r.currParty); idx++ {
		if !r.pStatus[idx].IsRobot() {
			return false
		}
	}
	return true
}

func (r *Room) reset() {
	r.roomId = 0
	r.roguelikeId = 0
	r.isOpen = false
	r.currParty = 0
	r.maxParty = 0
	r.pubIdx = -1
	r.startTime = 0
	r.updateAt = 0
	r.combatNeed = 0
	for idx := range r.party {
		r.party[idx] = ""
		r.pStatus[idx] = 0
		r.heroes[idx] = 0
		r.cards[idx] = 0
	}
	r.status = models.RoguelikeStatus_RoguelikeStatusStatusMatch
}

const (
	IsPartyRobot = 1 << iota
	IsPartyReady
)

type partyStatus int64

func (ps partyStatus) IsRobot() bool {
	return ps&IsPartyRobot > 0
}

func (ps partyStatus) IsReady() bool {
	return ps&IsPartyReady > 0
}
