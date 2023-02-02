package statistical2

import (
	"context"

	"coin-server/common/logger"
	"coin-server/common/statistical2/models"

	kafkago "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("statistical")

func Save(ls *LogServer, model models.Model) {
	ms, ok := ls.cache[model.Topic()]
	if !ok {
		ls.cache[model.Topic()] = make([]models.Model, 0)
	}
	ms = append(ms, model)
	ls.cache[model.Topic()] = ms
}

func Disabled(topic string) bool {
	v, ok := disable[topic]
	if !ok {
		return false
	}
	return v
}

func Flush(ls *LogServer) error {
	if Disabled("all") {
		return nil
	}

	if len(ls.GetCache()) == 0 {
		return nil
	}
	msgs := make([]kafkago.Message, 0)

	for topic, model := range ls.GetCache() {
		if Disabled(topic) {
			continue
		}
		for _, m := range model {
			value, err1 := m.ToJson()
			if err1 != nil {
				logger.DefaultLogger.Warn("kafka model to json failed", zap.String("topic", topic))
				continue
			}
			msgs = append(msgs, kafkago.Message{
				Key:   m.GetRoleId(), // 根据RoleId作为key，保证同一RoleId的消息被放到一个partition
				Value: value,
			})
		}
		w, ok := GetWriter(topic)
		if !ok {
			if logger.DefaultLogger != nil {
				logger.DefaultLogger.Warn("kafka writer not found", zap.String("topic", topic))
			}
			continue
		}
		if err := w.WriteMessages(context.Background(), msgs...); err != nil && logger.DefaultLogger != nil {
			switch errs := err.(type) {
			case nil:
			case kafkago.WriteErrors:
				if errs[0] != nil {
					logger.DefaultLogger.Error("call kafka.WriteMessages error", zap.Errors("errors", errs), zap.Any("data", model))
				}
			default:
				logger.DefaultLogger.Error("call kafka.WriteMessages error", zap.Error(err), zap.Any("data", model))
			}
			return err
		}
		msgs = msgs[:0]
	}

	return nil
}
