package syncrole

import (
	"context"
	"time"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/proto/dao"
	"coin-server/common/utils"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type SyncRole struct {
	log *logger.Logger
	kw  *kafka.Writer
}

type Config struct {
	Addr  []string `json:"addr"`
	Topic string   `json:"topic"`
}

var syncer *SyncRole

func Init(cfg *consulkv.Config) {
	conf := &Config{}
	utils.Must(cfg.Unmarshal("syncrole/kafka", conf))
	syncer = NewSync(logger.DefaultLogger, conf)
}

func Close() {
	if syncer != nil {
		syncer.kw.Close()
	}
}

func Create(ctx context.Context, role *dao.Role) error {
	return syncer.Create(ctx, role)
}

func Update(ctx context.Context, role *dao.Role) error {
	return syncer.Update(ctx, role)
}

func NewSync(log *logger.Logger, cfg *Config) *SyncRole {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Addr...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.Murmur2Balancer{},
		RequiredAcks: kafka.RequireNone,
		MaxAttempts:  3,
		BatchSize:    200,
		Async:        true,
		BatchTimeout: 100 * time.Millisecond,
	}
	return &SyncRole{log: log, kw: w}
}

func (s *SyncRole) Create(ctx context.Context, role *dao.Role) error {
	value, _ := (&dao.SyncRole{Op: dao.SyncOp_CREATE, Role: role}).Marshal()
	err := s.kw.WriteMessages(ctx, kafka.Message{
		Key:   []byte(role.RoleId),
		Value: value,
	})
	if err != nil {
		s.log.Error("SyncRole.Create: call kafka.WriteMessages error", zap.Error(err))
		return err
	}
	return nil
}

func (s *SyncRole) Update(ctx context.Context, role *dao.Role) error {
	value, _ := (&dao.SyncRole{Op: dao.SyncOp_UPDATE, Role: role}).Marshal()
	err := s.kw.WriteMessages(ctx, kafka.Message{
		Key:   []byte(role.RoleId),
		Value: value,
	})
	if err != nil {
		s.log.Error("SyncRole.Update: call kafka.WriteMessages error", zap.Error(err))
		return err
	}
	return nil
}
