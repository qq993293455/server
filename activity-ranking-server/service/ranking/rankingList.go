package activity_ranking

import (
	"sync"
)

type DataMapSt struct {
	dataMap *DataMap
	lock    *LocalMutex
}

func (this_ *DataMapSt) GetReadMap() *DataMap {
	this_.lock.RLock()
	return this_.dataMap
}

func (this_ *DataMapSt) GetClearWriteMap() *DataMap {
	this_.ClearWriteMap()
	return this_.GetWriteMap()
}

func (this_ *DataMapSt) GetWriteMap() *DataMap {
	this_.lock.WLock()
	return this_.dataMap
}

func (this_ *DataMapSt) ClearWriteMap() {
	this_.lock.WLock()
	this_.dataMap.Clear()
}

func (this_ *DataMapSt) UnLock() {
	this_.lock.UnLock()
}

type DataMap struct {
	RoleIdTorankingId map[string]int64
	RankingIdToRoleId map[int64]string
}

func (this_ *DataMap) Init() {
	this_.RoleIdTorankingId = make(map[string]int64)
	this_.RankingIdToRoleId = make(map[int64]string)
}

func (this_ *DataMap) Clear() {
	this_.RoleIdTorankingId = map[string]int64{}
	this_.RankingIdToRoleId = map[int64]string{}
}

type RankingList struct {
	Dirty
	readLock  sync.RWMutex
	readMap   *DataMap
	writeLock sync.RWMutex
	writeMap  *DataMap
}

func (this_ *RankingList) Init() {
	this_.readMap = &DataMap{
		RoleIdTorankingId: make(map[string]int64),
		RankingIdToRoleId: make(map[int64]string),
	}
	this_.writeMap = &DataMap{
		RoleIdTorankingId: make(map[string]int64),
		RankingIdToRoleId: make(map[int64]string),
	}
	this_.isDirty = true
}

func (this_ *RankingList) GetReadData() *DataMapSt {
	return &DataMapSt{
		lock:    GetLocalLock(&this_.readLock),
		dataMap: this_.readMap,
	}
}

func (this_ *RankingList) GetWriteData() *DataMapSt {
	return &DataMapSt{
		lock:    GetLocalLock(&this_.writeLock),
		dataMap: this_.writeMap,
	}
}

func Swap(data *DataMapSt, free *DataMapSt) {
	data.lock.WLock()
	free.lock.WLock()
	*data.dataMap, *free.dataMap = *free.dataMap, *data.dataMap
}
