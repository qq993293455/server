package tendissync

import (
	"context"
	"strconv"
	"time"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/utils"
	"coin-server/common/values"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var isOpen bool

type TendisSync struct {
	log      *logger.Logger
	kw       *kafka.Writer
	serverId string
}

type Config struct {
	Addr   []string `json:"addr"`
	Topic  string   `json:"topic"`
	IsOpen bool     `json:"is_open"`
}

var syncer *TendisSync

func Init(cfg *consulkv.Config, id values.ServerId) {
	conf := &Config{}
	utils.Must(cfg.Unmarshal("tendissync/kafka", conf))
	isOpen = conf.IsOpen
	syncer = NewSync(logger.DefaultLogger, conf, id)
}

func IsOpen() bool {
	return isOpen
}

func Close() {
	if syncer != nil {
		syncer.kw.Close()
	}
}

func Sync(ctx context.Context, cmd *models.TendisCmd) error {
	return syncer.Sync(ctx, cmd)
}

func NewSync(log *logger.Logger, cfg *Config, id values.ServerId) *TendisSync {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Addr...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Murmur2Balancer{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  3,
		BatchSize:    200,
		BatchTimeout: 100 * time.Millisecond,
	}
	return &TendisSync{log: log, kw: w, serverId: strconv.Itoa(int(id))}
}

func (s *TendisSync) Sync(ctx context.Context, cmd *models.TendisCmd) error {
	value, _ := cmd.Marshal()
	err := s.kw.WriteMessages(ctx, kafka.Message{
		Key:   []byte(s.serverId),
		Value: value,
	})
	if err != nil {
		s.log.Error("TendisSync.Sync: call kafka.WriteMessages error", zap.Error(err))
		return err
	}
	return nil
}
