package statistical

import (
	"context"

	"coin-server/common/logger"
	"coin-server/common/statistical/models"

	kafkago "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("statistical")

func Save(ls *LogServer, model models.Model) {
	model.Preset()
	ls.cache = append(ls.cache, model)
}

func Disabled(eventType string) bool {
	v, ok := disable[eventType]
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

	for _, model := range ls.GetCache() {
		topic := model.GetEventType()
		if Disabled(topic) {
			continue
		}
		msgs = append(msgs, kafkago.Message{
			Key:   model.HashKey(),
			Value: model.ToJson(),
		})
		w := GetWriter()
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
