package raft

import (
	"io"
	"strings"

	"coin-server/role-state-server/service/state/cache"
	"github.com/hashicorp/raft"
)

type FSM struct {
	slot *cache.Mgr
}

func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	logSlice := strings.Split(string(logEntry.Data), ":")
	if len(logSlice) != 2 {
		return nil
	}
	if logSlice[1] == "-1" {
		f.slot.Del(logSlice[0])
		return nil
	}
	f.slot.Set(logSlice[0], logSlice[1])
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{slot: f.slot}, nil
}

func (f *FSM) Restore(closer io.ReadCloser) error {
	defer closer.Close()
	return f.slot.FromIO(closer)
}

type snapshot struct {
	slot *cache.Mgr
}

func (ss *snapshot) Persist(sink raft.SnapshotSink) error {
	err := ss.slot.ToIO(sink)
	if err != nil {
		_ = sink.Cancel()
		return err
	}
	if err := sink.Close(); err != nil {
		_ = sink.Cancel()
		return err
	}
	return nil
}

func (ss *snapshot) Release() {

}
