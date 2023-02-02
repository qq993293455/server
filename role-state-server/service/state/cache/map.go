package cache

import (
	"bufio"
	"io"
	"sync"
	"sync/atomic"

	readpb "coin-server/common/proto/role_state_read"
	"coin-server/common/utils"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/raft"
)

const (
	slotMask = 0b1111 // 15
	slotNum  = slotMask + 1
)

type Mgr struct {
	slots [slotNum]*readpb.RoleStateROnly_CacheSlot
	locks [slotNum]sync.RWMutex
	cnt   int64
}

func NewSlot() *Mgr {
	mgr := &Mgr{}
	for idx := range mgr.slots {
		mgr.slots[idx] = &readpb.RoleStateROnly_CacheSlot{
			Data: map[string]string{},
		}
	}
	return mgr
}

func (s *Mgr) Get(key string) (bool, string) {
	h := hashString(key)
	s.locks[h].RLock()
	defer s.locks[h].RUnlock()
	val, ok := s.slots[h].Data[key]
	return ok, val
}

func (s *Mgr) GetSlotNum() int {
	return slotNum
}

func (s *Mgr) GetOneSlot(idx int) map[string][]string {
	if idx < 0 || idx >= slotNum {
		return nil
	}
	s.locks[idx].RLock()
	defer s.locks[idx].RUnlock()
	res := map[string][]string{}
	for roleId, gateId := range s.slots[idx].Data {
		if _, exist := res[gateId]; !exist {
			res[gateId] = make([]string, 0, 64)
		}
		res[gateId] = append(res[gateId], roleId)
	}
	return res
}

func (s *Mgr) Set(key, val string) {
	h := hashString(key)
	s.locks[h].Lock()
	defer s.locks[h].Unlock()
	if _, exist := s.slots[h].Data[key]; !exist {
		atomic.AddInt64(&s.cnt, 1)
	}
	s.slots[h].Data[key] = val
}

func (s *Mgr) Del(key string) {
	h := hashString(key)
	s.locks[h].Lock()
	defer s.locks[h].Unlock()
	if _, exist := s.slots[h].Data[key]; exist {
		delete(s.slots[h].Data, key)
		atomic.AddInt64(&s.cnt, -1)
	}
}

func (s *Mgr) ToIO(sink raft.SnapshotSink) error {
	for idx := range s.slots {
		s.locks[idx].RLock()
		size := s.slots[idx].Size()
		d := make([]byte, size, size+1)
		n, err := s.slots[idx].MarshalTo(d)
		s.locks[idx].RUnlock()
		if err != nil {
			return err
		}
		d = append(d[:n], '|')
		if _, err = sink.Write(d); err != nil {
			return err
		}
	}
	return nil
}

func (s *Mgr) FromIO(serialized io.ReadCloser) error {
	buf := bufio.NewReader(serialized)
	slotIdx := 0
	for {
		slot, err := buf.ReadBytes('|')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		slot = slot[:len(slot)-1]
		tmp := &readpb.RoleStateROnly_CacheSlot{}
		if err = proto.Unmarshal(slot, tmp); err != nil {
			return err
		}
		s.locks[slotIdx].Lock()
		if tmp.Data != nil {
			s.slots[slotIdx] = tmp
		}
		s.locks[slotIdx].Unlock()
		slotIdx++
	}
}

func hashString(id string) int {
	return int(utils.Base34DecodeString(id) & slotMask)
}
