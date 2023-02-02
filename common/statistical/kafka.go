package statistical

import (
	"time"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/utils"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	writer  *kafka.Writer
	disable map[string]bool
	gwId    int64
)

type Config struct {
	Addr     []string `json:"addr"`
	Topic    string   `json:"topic"`
	MinBytes int      `json:"min_bytes"`
	MaxBytes int      `json:"max_bytes"`
	GwId     int64    `json:"gw_id"` // 当前采集服务所属大世界ID
}

func Init(cfg *consulkv.Config) {
	kc := &Config{}
	utils.Must(cfg.Unmarshal("statistical/kafka", kc))
	gwId = kc.GwId
	utils.Must(cfg.Unmarshal("statistical/disabled", &disable))
	writer = &kafka.Writer{
		Addr:         kafka.TCP(kc.Addr...),
		Topic:        kc.Topic,
		RequiredAcks: kafka.RequireOne,
		MaxAttempts:  3,
		BatchSize:    300,
		Async:        true,
		BatchTimeout: 20 * time.Millisecond,
	}
}

func Close(logger *logger.Logger) {
	if err := writer.Close(); err != nil {
		logger.Error("close kafka writer error", zap.String("topic", writer.Topic), zap.Error(err))
	} else {
		logger.Info("close kafka writer success", zap.String("topic", writer.Topic))
	}
}

func GetWriter() *kafka.Writer {
	return writer
}

func GwId() int64 {
	return gwId
}
