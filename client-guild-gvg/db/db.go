package db

import (
	"coin-server/common/logger"
	"coin-server/common/safego"

	"go.uber.org/zap"

	"github.com/tidwall/buntdb"
)

func Open(filenames ...string) (*buntdb.DB, error) {
	fn := "client-guild-gvg.db"
	if len(filenames) != 0 {
		fn = filenames[0]
	}
	return buntdb.Open(fn)
}

type KV struct {
	Key   string
	Value string
}

type DB struct {
	db        *buntdb.DB
	log       *logger.Logger
	saveQueue chan *KV
	close     chan struct{}
}

func NewDB(log *logger.Logger, filenames ...string) *DB {
	db, err := Open(filenames...)
	if err != nil {
		panic(err)
	}
	return &DB{db: db, log: log, saveQueue: make(chan *KV, 100000)}
}

func (this_ *DB) save(kv *KV) {
	err := this_.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(kv.Key, kv.Value, nil)
		return err
	})
	if err != nil {
		this_.log.Warn("DB SAVE ERROR", zap.Error(err), zap.Any("data", kv))
	}
}

func (this_ *DB) Save(kv *KV) {
	if kv == nil || kv.Key == "" {
		return
	}
	select {
	case <-this_.close:
		return
	default:
		this_.saveQueue <- kv
	}
}

func (this_ *DB) Run() {
	safego.GOWithLogger(this_.log, func() {
		for kv := range this_.saveQueue {
			this_.save(kv)
		}
	})
}
