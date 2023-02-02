package raft

import (
	"fmt"
	"net"
	"os"
	"time"

	"coin-server/role-state-server/service/state/cache"
	raft_badger "coin-server/role-state-server/service/state/raft-badger"
	"github.com/hashicorp/raft"
)

type Options struct {
	ServerId  string
	TcpAddr   string
	Bootstrap bool
	Join      bool
	DBPath    string
}

func Init(opt *Options, leaderNotifyCh chan bool, slot *cache.Mgr) (has bool, raftC *raft.Raft) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(opt.ServerId)
	raftConfig.SnapshotInterval = 20 * time.Second
	raftConfig.SnapshotThreshold = 2
	raftConfig.NotifyCh = leaderNotifyCh
	logStore, err := raft_badger.NewBadgerStore(fmt.Sprintf("%s/badgerDB/raft-log%s", opt.DBPath, opt.ServerId))
	if err != nil {
		panic(err)
	}
	stableStore, err := raft_badger.NewBadgerStore(fmt.Sprintf("%s/badgerDB/stable-log%s", opt.DBPath, opt.ServerId))
	if err != nil {
		panic(err)
	}
	snapshotStore, err := raft.NewFileSnapshotStore(fmt.Sprintf("%s/ssDB%s", opt.DBPath, opt.ServerId), 1, os.Stderr)
	if err != nil {
		panic(err)
	}
	transport, err := newRaftTransport(opt)
	if err != nil {
		panic(err)
	}
	raftC, err = raft.NewRaft(raftConfig, &FSM{slot: slot}, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		panic(err)
	}
	has, err = raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		panic(err)
	}
	if !has && opt.Bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		f := raftC.BootstrapCluster(cfg)
		if err := f.Error(); err != nil {
			panic(err)
		}
	}
	return has, raftC
}

func newRaftTransport(opt *Options) (*raft.NetworkTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", opt.TcpAddr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(address.String(), address, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	return transport, nil
}
