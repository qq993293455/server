package activity_ranking

import (
	"sync"
)

type LocalMutex struct {
	IsRLock bool
	IsWLock bool
	Lock    *sync.RWMutex
}

func (this_ *LocalMutex) RLock() bool {
	if this_.IsWLock || this_.IsRLock {
		return true
	}
	if this_.Lock == nil {
		return false
	}
	this_.Lock.RLock()
	this_.IsRLock = true
	return true
}

func (this_ *LocalMutex) WLock() bool {
	if this_.IsWLock {
		return true
	}
	if this_.IsRLock {
		this_.Lock.RUnlock()
		this_.IsRLock = false
	}
	if this_.Lock == nil {
		return false
	}
	this_.Lock.Lock()
	this_.IsWLock = true
	return true
}

func (this_ *LocalMutex) UnLock() {
	if this_.IsWLock {
		this_.Lock.Unlock()
		this_.IsWLock = false
	}
	if this_.IsRLock {
		this_.Lock.RUnlock()
		this_.IsRLock = false
	}
}

func GetLocalLock(lock *sync.RWMutex) *LocalMutex {
	return &LocalMutex{
		Lock: lock,
	}
}
