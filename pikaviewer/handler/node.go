package handler

import redis "coin-server/common/redisclient"

type Nodes struct {
}

func NewNodes() *Nodes {
	return &Nodes{}
}

func (n *Nodes) ModuleNodes() []string {
	//return redis.GetAllRouterHashKey()
	return nil
}

func (n *Nodes) PikaNodes() []string {
	return redis.GetPikaNodes()
}
