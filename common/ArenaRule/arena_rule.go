package ArenaRule

import (
	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/utils"
)

type ArenaConfig struct {
	ArenaType map[int64]int64 `json:"ArenaType"`
	PprofAddr string
}

type ArenaRule struct {
	config *ArenaConfig
	log    *logger.Logger
}

var arenaRule *ArenaRule

func Init(cfg *consulkv.Config) {
	conf := &ArenaConfig{}
	utils.Must(cfg.Unmarshal("Arena/ArenaRule", conf))
	arenaRule = NewArenaRule(logger.DefaultLogger, conf)
}

func NewArenaRule(log *logger.Logger, cfg *ArenaConfig) *ArenaRule {
	return &ArenaRule{
		log:    log,
		config: cfg,
	}
}

func GetArenaServer(aType models.ArenaType) int64 {
	serverId, ok := arenaRule.config.ArenaType[int64(aType)]
	if !ok {
		return 1
	}
	return serverId
}

func GetPprofAddr() string {
	return arenaRule.config.PprofAddr
}
