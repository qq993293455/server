package model

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

type CdKeyGen struct {
	Id        string `db:"id" json:"id"`
	BatchId   int64  `db:"batch_id" json:"batch_id"`
	BeginAt   int64  `db:"begin_at" json:"begin_at"`
	EndAt     int64  `db:"end_at" json:"end_at"`
	LimitCnt  int64  `db:"limit_cnt" json:"limit_cnt"`
	LimitTyp  int64  `db:"limit_typ" json:"limit_typ"`
	Reward    string `db:"reward" json:"reward"`
	UseCnt    int64  `db:"use_cnt" json:"use_cnt"`
	IsActive  bool   `db:"is_active" json:"is_active"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

const BatchKey = "cdkey_batch_key"

func NewCdKeyGen() *CdKeyGen {
	return &CdKeyGen{}
}

func (c *CdKeyGen) Save(data *CdKeyGen) error {
	query := `INSERT INTO cd_key_gen (id,batch_id,begin_at,end_at,limit_cnt,limit_typ,reward,use_cnt,is_active,created_at)
			VALUES (:id,:batch_id,:begin_at,:end_at,:limit_cnt,:limit_typ,:reward,:use_cnt,:is_active,:created_at);`

	if _, err := orm.GetMySQL().NamedExec(query, data); err != nil {
		return err
	}
	return nil
}

func (c *CdKeyGen) Saves(data []*CdKeyGen) error {
	// TODO: mysql batch
	for _, r := range data {
		if err := c.Save(r); err != nil {
			return err
		}
	}
	return nil
}

func (c *CdKeyGen) DeActive(batchId int64) error {
	query := `UPDATE cd_key_gen SET is_active=false WHERE batch_id = ?`
	if _, err := orm.GetMySQL().Exec(query, batchId); err != nil {
		return err
	}
	return nil
}

func (c *CdKeyGen) KeyExist(key string) (bool, error) {
	query := `SELECT COUNT(*) num FROM cd_key_gen where id = ?`
	var count int
	if err := orm.GetMySQL().Get(&count, query, key); err != nil {
		return false, err
	}
	return count > 0, nil
}

func GetBatchPB() (*dao.CdKeyBatch, *errmsg.ErrMsg) {
	data := &dao.CdKeyBatch{Id: BatchKey}
	db := orm.GetOrm(ctx.GetContext())
	ok, err := db.GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	if !ok {
		data = &dao.CdKeyBatch{Id: BatchKey, BatchId: 1}
		db.SetPB(redisclient.GetDefaultRedis(), data)
		return data, nil
	}
	return data, nil
}

func SaveBatchPB(data *dao.CdKeyBatch) *errmsg.ErrMsg {
	db := orm.GetOrm(ctx.GetContext())
	db.SetPB(redisclient.GetDefaultRedis(), data)
	return db.Do()
}
