package room_mgr

import (
	"sync"

	"coin-server/common/proto/models"
	"coin-server/common/values"
)

var robotInfoPool *sync.Pool

func init() {
	robotInfoPool = &sync.Pool{
		New: func() interface{} {
			return &models.Bot{}
		},
	}
}

type RobotInfoMgr struct {
	infoList [roleSlotNum]*robotInfoMap
}

func NewRobotInfoMgr() *RobotInfoMgr {
	mgr := &RobotInfoMgr{
		infoList: [roleSlotNum]*robotInfoMap{},
	}
	for idx := range mgr.infoList {
		mgr.infoList[idx] = newRobotInfoMap()
	}
	return mgr
}

func (rm *RobotInfoMgr) Get(roleId values.RoleId) *models.Bot {
	return rm.infoList[hashRole(roleId)].get(roleId)
}

func (rm *RobotInfoMgr) Set(roleId values.RoleId, robotId values.Integer, nickName string) *models.Bot {
	return rm.infoList[hashRole(roleId)].set(roleId, robotId, nickName)
}

func (rm *RobotInfoMgr) Del(roleId values.RoleId) {
	rm.infoList[hashRole(roleId)].delete(roleId)
}

type robotInfoMap struct {
	*sync.RWMutex
	m map[values.RoleId]*models.Bot
}

func newRobotInfoMap() *robotInfoMap {
	return &robotInfoMap{
		&sync.RWMutex{},
		map[values.RoleId]*models.Bot{},
	}
}

func (rim *robotInfoMap) set(roleId values.RoleId, robotId values.Integer, nickName string) *models.Bot {
	rim.Lock()
	defer rim.Unlock()
	if v, exist := rim.m[roleId]; !exist || v == nil {
		data := newRobotImpleInfo(roleId, robotId, nickName)
		rim.m[roleId] = data
		return data
	}
	v := rim.m[roleId]
	v.RoleId = roleId
	v.RobotId = robotId
	v.Nickname = nickName
	return v
}

func (rim *robotInfoMap) get(roleId values.RoleId) *models.Bot {
	rim.RLock()
	defer rim.RUnlock()
	return rim.m[roleId]
}

func (rim *robotInfoMap) delete(roleId values.RoleId) {
	if v, exist := rim.m[roleId]; !exist || v == nil {
		return
	}
	role := rim.m[roleId]
	delete(rim.m, roleId)
	role.RoleId = ""
	role.RobotId = 0
	role.Nickname = ""
	robotInfoPool.Put(role)
}

func newRobotImpleInfo(roleId values.RoleId, robotId values.Integer, nickName string) *models.Bot {
	rsi := robotInfoPool.Get().(*models.Bot)
	rsi.RoleId = roleId
	rsi.RobotId = robotId
	rsi.Nickname = nickName
	return rsi
}
