package group_lock

import (
	"sync"

	"coin-server/common/utils"
)

const (
	groupMask = 0b1111111111 // 1023
	groupCnt  = groupMask + 1
)

// 分段锁，切分map避免对整个map加锁
var groupLock [groupCnt]*sync.Mutex

func init() {
	for idx := range groupLock {
		groupLock[idx] = &sync.Mutex{}
	}
}

func hashString(id string) int {
	return int(utils.Base34DecodeString(id) & groupMask)
}

func hashInt(id uint64) int {
	return int(id & groupMask)
}

func LockRole(key string) {
	groupLock[hashString(key)].Lock()
}

func LockRoom(key uint64) {
	groupLock[hashInt(key)].Lock()
}

func UnlockRole(key string) {
	groupLock[hashString(key)].Unlock()
}

func UnlockRoom(key uint64) {
	groupLock[hashInt(key)].Unlock()
}
