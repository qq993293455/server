package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"coin-server/common/consulkv"
	"coin-server/common/logger"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/utils"
	"coin-server/sync-role-worker/dao"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Config struct {
	Addr  []string `json:"addr"`
	Topic string   `json:"topic"`
	Group string   `json:"group"`
}

var dw *DBWorker

func Init(config *consulkv.Config) {
	cfg := &Config{}
	utils.Must(config.Unmarshal("syncrole/kafka", cfg))
	dw = NewDBWorker(cfg, dao.GetDao(), logger.DefaultLogger)
}

func Start() {
	dw.Start()
}

func Close(ctx context.Context) error {
	return dw.Close(ctx)
}

type DBWorker struct {
	conf    *Config
	dao     *dao.Dao
	log     *logger.Logger
	kr      *kafka.Reader
	close   chan struct{}
	closed1 int32
	closed2 int32

	batchData []dao.Role
}

func NewDBWorker(conf *Config, _dao *dao.Dao, log *logger.Logger) *DBWorker {
	kr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        conf.Addr,
		GroupID:        conf.Group,
		Topic:          conf.Topic,
		MinBytes:       1,
		MaxBytes:       50e6, // 50MB
		CommitInterval: 100 * time.Millisecond,
	})

	return &DBWorker{
		conf:      conf,
		dao:       _dao,
		log:       log,
		kr:        kr,
		close:     make(chan struct{}),
		batchData: make([]dao.Role, 0, 200),
	}
}

func (d *DBWorker) Start() {
	go d.KafkaConsume2MySQL()
}

var (
	protoPool = sync.Pool{New: func() interface{} { return new(daopb.SyncRole) }}
	// dbMsgPool = sync.Pool{New: func() interface{} { return new(dao.Role) }}
)

func protoPoolPut(p *daopb.SyncRole) {
	p.Reset()
	protoPool.Put(p)
}

// func dbMsgPoolPut(p *dao.Role) {
//	p.Reset()
//	dbMsgPool.Put(p)
// }

func (d *DBWorker) KafkaConsume2MySQL() {
	ctx := context.Background()
	prevTime := time.Now()
	var proto *daopb.SyncRole
	var mcount int
	var msg *kafka.Message

	for {
		select {
		case <-d.close:
			atomic.StoreInt32(&d.closed1, 1)
			return
		default:
		}
		fetchCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
		tempMsg, err := d.kr.FetchMessage(fetchCtx)
		cancel()
		if err != nil {
			if err == context.DeadlineExceeded {
				goto Sync
			}
			d.log.Warn("call kafka.FetchMessage failed", zap.Error(err))
			continue
		}

		msg = &tempMsg
		d.log.Info("[mysql] received message.", zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))

		proto = protoPool.Get().(*daopb.SyncRole)
		err = proto.Unmarshal(msg.Value)
		if err != nil {
			d.log.Warn("call proto.Unmarshal failed", zap.Error(err))
			protoPoolPut(proto)
			continue
		}

		mcount = d.mysqlProcess(ctx, proto)
		protoPoolPut(proto)

	Sync:
		if mcount >= 200 || time.Now().Sub(prevTime).Milliseconds() >= 200 {
			if len(d.batchData) == 0 {
				continue
			}
			for {
				err = d.mysqlBatchSave(ctx)
				if err != nil {
					d.log.Warn("call mysqlBatchSave failed", zap.Error(err))
					time.Sleep(100 * time.Millisecond)
				} else {
					break
				}
			}
			if err = d.kr.CommitMessages(ctx, *msg); err != nil {
				d.log.Warn("call kafka.CommitMessages failed", zap.Error(err))
			}
			d.log.Info("[mysql] commit message.", zap.String("topic", msg.Topic), zap.Int("partition", msg.Partition), zap.Int64("offset", msg.Offset))
			prevTime = time.Now()
		}
	}
}

func (d *DBWorker) mysqlProcess(ctx context.Context, body *daopb.SyncRole) (count int) {
	role := body.Role
	d.batchData = append(d.batchData, dao.Role{
		RoleId:       role.RoleId,
		IggId:        role.UserId,
		Nickname:     role.Nickname,
		Level:        role.Level,
		Avatar:       role.AvatarId,
		AvatarFrame:  role.AvatarFrame,
		Power:        role.Power,
		HighestPower: role.HighestPower,
		Title:        role.Title,
		Language:     role.Language,
		LoginTime:    role.Login,
		LogoutTime:   role.Logout,
		CreateTime:   role.CreateTime,
	})
	return len(d.batchData)
}

func (d *DBWorker) mysqlBatchSave(ctx context.Context) (err error) {
	if len(d.batchData) == 0 {
		return nil
	}
	d.log.Info("mysqlBatchSave", zap.Int("count", len(d.batchData)))
	err = d.dao.BatchSave(ctx, d.batchData)
	if err != nil {
		return err
	}
	d.batchData = d.batchData[:0]
	return nil
}

func (d *DBWorker) Close(ctx context.Context) error {
	close(d.close)
	d.kr.Close()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if atomic.LoadInt32(&d.closed1) == 1 && atomic.LoadInt32(&d.closed2) == 1 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
