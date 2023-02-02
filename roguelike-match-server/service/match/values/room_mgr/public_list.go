package room_mgr

import (
	"math/rand"
)

const (
	initListLen = 2048

	maxScanCnt   = 20
	maxSearchCnt = 200
)

type publicList struct {
	// *sync.RWMutex
	rooms []*Room
}

func newPublicList() *publicList {
	l := &publicList{
		// &sync.RWMutex{},
		make([]*Room, 0, initListLen),
	}
	return l
}

// 发布
func (pl *publicList) pub(r *Room) {
	pl.rooms = append(pl.rooms, r)
	r.pubIdx = int64(len(pl.rooms) - 1)
}

// 下架
func (pl *publicList) off(r *Room) {
	pl.rooms[r.pubIdx], pl.rooms[len(pl.rooms)-1] = pl.rooms[len(pl.rooms)-1], pl.rooms[r.pubIdx]
	pl.rooms[r.pubIdx].pubIdx = r.pubIdx
	r.pubIdx = -1
	pl.rooms = pl.rooms[:len(pl.rooms)-1]
}

func (pl *publicList) getRandomRoom(roomTime int64) *Room {
	if len(pl.rooms) == 0 {
		return nil
	}
	searchCnt := 0
	for searchCnt < maxScanCnt {
		idx := rand.Intn(len(pl.rooms))
		r := pl.rooms[idx]
		if r.IsOpen() && !r.IsStarted() && !r.isFull() && !r.CheckClose(roomTime) {
			return r
		}
		searchCnt++
	}
	return nil
}

func (pl *publicList) len() int {
	return len(pl.rooms)
}
