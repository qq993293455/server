package redisclient

import (
	"context"
	"strings"

	"coin-server/common/metrics"
	"coin-server/common/proto/models"
	tendissync "coin-server/common/tendis-sync"
	"coin-server/common/utils"

	"github.com/go-redis/redis/extra/rediscmd/v8"
	"github.com/go-redis/redis/v8"
)

type MetricsHook struct{}

var _ redis.Hook = (*MetricsHook)(nil)

func NewMetricsHook() *MetricsHook {
	return new(MetricsHook)
}

func (MetricsHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (MetricsHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if !metrics.IsOpen() {
		return nil
	}
	dao := drawDao(cmd)
	if dao == "" {
		return nil
	}
	metrics.DBProcessTotal.WithLabelValues(cmd.FullName(), dao).Inc()
	metrics.DBDataSizeSummary.WithLabelValues(cmd.FullName(), dao).Observe(float64(len(utils.StringToBytes(cmd.String()))))
	return nil
}

func (MetricsHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (MetricsHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if !metrics.IsOpen() {
		return nil
	}
	for _, cmd := range cmds {
		dao := drawDao(cmd)
		if dao == "" {
			continue
		}
		metrics.DBProcessTotal.WithLabelValues(cmd.FullName(), dao).Inc()
		metrics.DBDataSizeSummary.WithLabelValues(cmd.FullName(), dao).Observe(float64(len(utils.StringToBytes(cmd.String()))))
	}
	return nil
}

func drawDao(cmd redis.Cmder) string {
	key := drawKey(cmd)
	if key == "" {
		return key
	}
	return strings.SplitN(key, ":", 2)[0]
}

func drawKey(cmd redis.Cmder) string {
	args := cmd.Args()
	if len(args) < 2 {
		return ""
	}
	var val interface{}
	switch name := cmd.Name(); name {
	case "cluster", "command":
		if len(args) < 3 {
			return ""
		}
		val = args[2]
	default:
		val = args[1]
	}
	if key, ok := val.(string); ok {
		return key
	}
	return ""
}

type TendisSyncHook struct{}

var _ redis.Hook = (*TendisSyncHook)(nil)

func NewTendisSyncHook() *TendisSyncHook {
	return new(TendisSyncHook)
}

var setMap = map[string]struct{}{
	"set":   {},
	"del":   {},
	"mset":  {},
	"hset":  {},
	"hmset": {},
	"hdel":  {},
}

func (TendisSyncHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (TendisSyncHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if !tendissync.IsOpen() {
		return nil
	}
	if _, ok := setMap[cmd.Name()]; !ok {
		return nil
	}
	return tendissync.Sync(ctx, &models.TendisCmd{Cmd: rediscmd.CmdString(cmd)})
}

func (TendisSyncHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (TendisSyncHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	if !tendissync.IsOpen() {
		return nil
	}
	do := false
	for _, cmd := range cmds {
		if _, ok := setMap[cmd.Name()]; ok {
			do = true
			break
		}
	}
	if do {
		_, cmdstring := rediscmd.CmdsString(cmds)
		return tendissync.Sync(ctx, &models.TendisCmd{Cmd: cmdstring})
	}
	return nil
}
