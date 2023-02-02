package nutsdb

import (
	"github.com/xujiajun/nutsdb"
)

type NutsDb struct {
	option *nutsdb.Options
	db     *nutsdb.DB
	bucket string
}

func NewNutsDb() *NutsDb {
	opt := nutsdb.DefaultOptions
	opt.Dir = "tcp-client/nutsdb"
	n := &NutsDb{
		option: &opt,
		db:     nil,
		bucket: "config",
	}
	Db = n
	return n
}

func (n *NutsDb) Connect() error {
	db, err := nutsdb.Open(*n.option)
	if err != nil {
		return err
	}
	n.db = db
	return nil
}

func (n *NutsDb) Get(key string) []byte {
	var value []byte
	f := func(tx *nutsdb.Tx) error {
		k := []byte(key)
		if e, err := tx.Get(n.bucket, k); err != nil {
			return err
		} else {
			value = e.Value
			return nil
		}
	}
	err := n.db.View(f)
	if err != nil {
		return nil
	}
	return value
}

func (n *NutsDb) Set(key string, value []byte) error {
	f := func(tx *nutsdb.Tx) error {
		k := []byte(key)
		v := value
		if err := tx.Put(n.bucket, k, v, 0); err != nil {
			return err
		}
		return nil
	}
	return n.db.Update(f)
}

func (n *NutsDb) Del(key string) error {
	f := func(tx *nutsdb.Tx) error {
		k := []byte(key)
		if err := tx.Delete(n.bucket, k); err != nil {
			return err
		}
		return nil
	}
	return n.db.Update(f)
}

func (n *NutsDb) Close() error {
	return n.db.Close()
}
