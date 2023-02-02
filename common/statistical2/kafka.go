package statistical2

import (
	"time"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/statistical2/models"
	"coin-server/common/utils"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	writer  map[string]*kafka.Writer
	disable map[string]bool
)

type Config struct {
	Addr     []string `json:"addr"`
	MinBytes int      `json:"min_bytes"`
	MaxBytes int      `json:"max_bytes"`
}

func Init(cfg *consulkv.Config) {
	kc := &Config{}
	utils.Must(cfg.Unmarshal("statistical2/kafka", kc))
	utils.Must(cfg.Unmarshal("statistical2/disabled", &disable))
	writer = make(map[string]*kafka.Writer)
	for _, topic := range models.TopicList {
		writer[topic] = &kafka.Writer{
			Addr:         kafka.TCP(kc.Addr...),
			Topic:        topic,
			RequiredAcks: kafka.RequireOne,
			MaxAttempts:  3,
			BatchSize:    300,
			Async:        true,
			BatchTimeout: 100 * time.Millisecond,
		}
	}
}

func Close(logger *logger.Logger) {
	for _, w := range writer {
		if err := w.Close(); err != nil {
			logger.Error("close kafka writer error", zap.String("topic", w.Topic), zap.Error(err))
		} else {
			logger.Info("close kafka writer success", zap.String("topic", w.Topic))
		}
	}
}

func GetWriter(topic string) (*kafka.Writer, bool) {
	w, ok := writer[topic]
	return w, ok
}
