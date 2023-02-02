package consumer

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"coin-server/common/logger"
	"coin-server/common/statistical2/models"
	"coin-server/statistical-server/database"
	"coin-server/statistical-server/env"
	kafka2 "coin-server/statistical-server/kafka"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	closeChan = make(chan struct{})
	closed    int32
)

func Close(ctx context.Context) error {
	close(closeChan)
	kafka2.Reader.Close()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if atomic.LoadInt32(&closed) == 1 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func Start() {
	go func() {
		logger.DefaultLogger.Info("start consumer")
		consuming(kafka2.Reader)
	}()
	go func() {
		logger.DefaultLogger.Info("start login consumer")
		loginConsuming(kafka2.LoginReader)
	}()
}

func consuming(reader *kafka.Reader) {
	ctx := context.Background()
	debug := env.IsDebug()
	count := 0
	prevTime := time.Now()
	data := make(map[string][]models.Model, 0)
	msgMap := make(map[string]*kafka.Message)
	var msg *kafka.Message
	var model models.Model
	var temp models.Model
	var ok bool

	for {
		select {
		case <-closeChan:
			atomic.StoreInt32(&closed, 1)
			return
		default:
		}

		fetchCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
		tempMsg, err := reader.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				goto Sync
			}
			logger.DefaultLogger.Error("kafka fetch message error: ", zap.Error(err))
			time.Sleep(time.Millisecond * 200)
			continue
		}

		msg = &tempMsg
		if debug {
			logger.DefaultLogger.Debug("kafka message: ", zap.String("topic", msg.Topic), zap.Int64("offset", msg.Offset), zap.String("value", string(msg.Value)))
		}
		model, ok = models.TopicModelMap[msg.Topic]
		if !ok {
			logger.DefaultLogger.Error("kafka message topic model not found: ", zap.String("topic", msg.Topic), zap.Int64("offset", msg.Offset))
			continue
		}
		temp = model.NewModel()
		if err = json.Unmarshal(msg.Value, &temp); err != nil {
			logger.DefaultLogger.Error("kafka message json unmarshal error: ", zap.Error(err), zap.String("topic", msg.Topic), zap.String("value", string(msg.Value)))
			continue
		}

		msgMap[msg.Topic] = msg
		data[msg.Topic] = append(data[msg.Topic], temp)
		count++

	Sync:
		if count == 0 {
			continue
		}
		if count >= 200 || time.Now().Sub(prevTime).Milliseconds() >= 500 {
			if debug {
				logger.DefaultLogger.Debug("kafka message save to mysql: ", zap.Int("count", count))
			}
			for err = database.SaveToMysql(data); err != nil; {
				logger.DefaultLogger.Warn("call SaveToMysql failed", zap.Error(err))
				time.Sleep(200 * time.Millisecond)
			}
			count = 0

			for _, m := range msgMap {
				if debug {
					logger.DefaultLogger.Debug("kafka commit message: ", zap.String("topic", m.Topic), zap.Int64("offset", m.Offset))
				}
				if err = reader.CommitMessages(context.Background(), *m); err != nil {
					logger.DefaultLogger.Error("kafka commit message error: ", zap.Error(err), zap.String("topic", m.Topic), zap.String("value", string(m.Value)))
					continue
				}
			}
			msgMap = map[string]*kafka.Message{}
			prevTime = time.Now()
		}
	}
}

func loginConsuming(reader *kafka.Reader) {
	ctx := context.Background()
	debug := env.IsDebug()
	count := 0
	prevTime := time.Now()
	data := make(map[string][]models.Model, 0)
	msgMap := make(map[string]*kafka.Message)
	var msg *kafka.Message
	var model models.Model
	var temp models.Model
	var ok bool

	for {
		select {
		case <-closeChan:
			atomic.StoreInt32(&closed, 1)
			return
		default:
		}

		fetchCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
		tempMsg, err := reader.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				goto Sync
			}
			logger.DefaultLogger.Error("kafka fetch message error: ", zap.Error(err))
			time.Sleep(time.Millisecond * 200)
			continue
		}

		msg = &tempMsg
		if debug {
			logger.DefaultLogger.Debug("kafka message: ", zap.String("topic", msg.Topic), zap.Int64("offset", msg.Offset), zap.String("value", string(msg.Value)))
		}
		model, ok = models.TopicModelMap[msg.Topic]
		if !ok {
			logger.DefaultLogger.Error("kafka message topic model not found: ", zap.String("topic", msg.Topic), zap.Int64("offset", msg.Offset))
			continue
		}
		temp = model.NewModel()
		if err = json.Unmarshal(msg.Value, &temp); err != nil {
			logger.DefaultLogger.Error("kafka message json unmarshal error: ", zap.Error(err), zap.String("topic", msg.Topic), zap.String("value", string(msg.Value)))
			continue
		}

		msgMap[msg.Topic] = msg
		data[msg.Topic] = append(data[msg.Topic], temp)
		count++

	Sync:
		if count == 0 {
			continue
		}
		if count >= 200 || time.Now().Sub(prevTime).Milliseconds() >= 500 {
			if debug {
				logger.DefaultLogger.Debug("kafka message save to mysql: ", zap.Int("count", count))
			}
			for err = database.ExecProcedure(data); err != nil; {
				logger.DefaultLogger.Warn("call ExecProcedure failed", zap.Error(err))
				time.Sleep(200 * time.Millisecond)
			}
			count = 0

			for _, m := range msgMap {
				if debug {
					logger.DefaultLogger.Debug("kafka commit message: ", zap.String("topic", m.Topic), zap.Int64("offset", m.Offset))
				}
				if err = reader.CommitMessages(context.Background(), *m); err != nil {
					logger.DefaultLogger.Error("kafka commit message error: ", zap.Error(err), zap.String("topic", m.Topic), zap.String("value", string(m.Value)))
					continue
				}
			}
			msgMap = map[string]*kafka.Message{}
			prevTime = time.Now()
		}
	}
}
