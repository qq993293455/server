package ActivityRankingRule

import (
	"coin-server/common/consulkv"
)

const (
	ActivityRankingServerId_default = 1000
)

type ActivityRankingConfig struct {
}

type ActivityRankingRule struct {
}

func Init(cfg *consulkv.Config) {
}

func GetActivityRankingServer(activityId int64) int64 {
	return ActivityRankingServerId_default
}
