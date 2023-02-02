package room_mgr

import (
	"sync"
	"time"

	"coin-server/common/proto/models"
	"coin-server/common/timer"
	"coin-server/common/values"
)

const (
	clearTime int64 = 10
)

var roleSimpleInfoPool *sync.Pool

func init() {
	roleSimpleInfoPool = &sync.Pool{
		New: func() interface{} {
			return &RoleSimpleInfo{}
		},
	}
}

// TODO: 后面用lru缓存库来做
type RoleInfoMgr struct {
	infoList [roleSlotNum]*roleInfoMap
}

func NewRoleInfoMgr() *RoleInfoMgr {
	mgr := &RoleInfoMgr{
		infoList: [roleSlotNum]*roleInfoMap{},
	}
	for idx := range mgr.infoList {
		mgr.infoList[idx] = newRoleInfoMap()
	}
	go func() {
		for {
			time.Sleep(time.Minute)
			for idx := range mgr.infoList {
				mgr.infoList[idx].check()
			}
		}
	}()
	return mgr
}

func (rm *RoleInfoMgr) Get(roleId values.RoleId) *RoleSimpleInfo {
	return rm.infoList[hashRole(roleId)].get(roleId)
}

func (rm *RoleInfoMgr) Set(roleId values.RoleId, info *models.UserSimpleInfo) *RoleSimpleInfo {
	return rm.infoList[hashRole(roleId)].set(roleId, info)
}

type roleInfoMap struct {
	*sync.RWMutex
	m map[values.RoleId]*RoleSimpleInfo
}

func newRoleInfoMap() *roleInfoMap {
	return &roleInfoMap{
		&sync.RWMutex{},
		map[values.RoleId]*RoleSimpleInfo{},
	}
}

func (rim *roleInfoMap) set(roleId values.RoleId, info *models.UserSimpleInfo) *RoleSimpleInfo {
	rim.Lock()
	defer rim.Unlock()
	if v, exist := rim.m[roleId]; !exist || v == nil {
		data := newRoleSimpleInfo(info)
		rim.m[roleId] = data
		return data
	}
	return rim.m[roleId].set(info)
}

func (rim *roleInfoMap) get(roleId values.RoleId) *RoleSimpleInfo {
	rim.RLock()
	defer rim.RUnlock()
	rsl, exist := rim.m[roleId]
	if !exist {
		return nil
	}
	if timer.Unix()-rsl.updateAt > clearTime {
		return nil
	}
	return rsl
}

func (rim *roleInfoMap) delete(roleId values.RoleId) {
	if v, exist := rim.m[roleId]; !exist || v == nil {
		return
	}
	role := rim.m[roleId]
	delete(rim.m, roleId)
	role.reset()
	roleSimpleInfoPool.Put(role)
}

func (rim *roleInfoMap) check() {
	rim.Lock()
	defer rim.Unlock()
	now := timer.Unix()
	for roleId, v := range rim.m {
		if now-v.updateAt > clearTime {
			rim.delete(roleId)
		}
	}
}

type RoleSimpleInfo struct {
	*models.UserSimpleInfo
	updateAt int64
}

func newRoleSimpleInfo(info *models.UserSimpleInfo) *RoleSimpleInfo {
	rsi := roleSimpleInfoPool.Get().(*RoleSimpleInfo)
	rsi.set(info)
	return rsi
}

func (rsi *RoleSimpleInfo) set(info *models.UserSimpleInfo) *RoleSimpleInfo {
	rsi.UserSimpleInfo = info
	rsi.updateAt = timer.Unix()
	return rsi
}

func (rsi *RoleSimpleInfo) reset() {
	rsi.UserSimpleInfo = nil
	rsi.updateAt = 0
}
