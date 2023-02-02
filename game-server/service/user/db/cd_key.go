package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/timer"
)

type CdKeyGen struct {
	Id        string `db:"id" json:"id"`
	BatchId   int64  `db:"batch_id" json:"batch_id"`
	BeginAt   int64  `db:"begin_at" json:"begin_at"`
	EndAt     int64  `db:"end_at" json:"end_at"`
	LimitCnt  int64  `db:"limit_cnt" json:"limit_cnt"`
	LimitTyp  int64  `db:"limit_typ" json:"limit_typ"`
	Reward    string `db:"reward" json:"reward"`
	IsActive  bool   `db:"is_active" json:"is_active"`
	UseCnt    int64  `db:"use_cnt" json:"use_cnt"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

type CdKeyUse struct {
	Id        int64  `db:"id" json:"id"`
	RoleId    string `db:"role_id" json:"role_id"`
	BatchId   int64  `db:"batch_id" json:"batch_id"`
	KeyId     string `db:"key_id" json:"key_id"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

func GetKeyGen(c *ctx.Context, key string) (*CdKeyGen, *errmsg.ErrMsg) {
	data := &CdKeyGen{}
	t := timer.StartTime(c.StartTime).UnixMilli()
	query := "SELECT * FROM cd_key_gen WHERE id=? and begin_at <= ? and end_at >= ?"
	if err := orm.GetMySQL().Get(data, query, key, t, t); err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	return data, nil
}

func HasKeyUse(roleId string, batchId int64) (bool, *errmsg.ErrMsg) {
	query := "SELECT COUNT(*) num FROM cd_key_use where role_id = ? and batch_id = ?"
	var count int
	if err := orm.GetMySQL().Get(&count, query, roleId, batchId); err != nil {
		return false, errmsg.NewErrorDB(err)
	}
	return count > 0, nil
}

// KeyUseRecord 多人单次,供所有人使用，每个玩家限用一次
func KeyUseRecord(c *ctx.Context, use *CdKeyUse) *errmsg.ErrMsg {
	tx, err := orm.GetMySQL().BeginTx(c, nil)
	if err != nil {
		return errmsg.NewErrorDB(err)
	}
	q1 := "UPDATE cd_key_gen SET use_cnt=use_cnt+1 WHERE id = ? and is_active = true and limit_cnt > 0"
	if _, err = tx.Exec(q1, use.KeyId); err != nil {
		return errmsg.NewErrCdKeyIsUsed()
	}
	q2 := `INSERT INTO cd_key_use (role_id,batch_id,key_id,created_at) VALUES (?, ?, ?, ?)`
	if _, err = tx.Exec(q2, use.RoleId, use.BatchId, use.KeyId, use.CreatedAt); err != nil {
		return errmsg.NewErrorDB(err)
	}
	if err = tx.Commit(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

// KeyUseDel 单人单次,一个人用一次
func KeyUseDel(c *ctx.Context, use *CdKeyUse) *errmsg.ErrMsg {
	tx, err := orm.GetMySQL().BeginTx(c, nil)
	if err != nil {
		return errmsg.NewErrorDB(err)
	}
	q1 := "UPDATE cd_key_gen SET limit_cnt=limit_cnt-1, use_cnt=use_cnt+1, is_active=false WHERE id = ? and is_active = true and limit_cnt > 0"
	if _, err = tx.Exec(q1, use.KeyId); err != nil {
		return errmsg.NewErrCdKeyIsUsed()
	}
	q2 := `INSERT INTO cd_key_use (role_id,batch_id,key_id,created_at) VALUES (?, ?, ?, ?)`
	if _, err = tx.Exec(q2, use.RoleId, use.BatchId, use.KeyId, use.CreatedAt); err != nil {
		return errmsg.NewErrorDB(err)
	}
	if err = tx.Commit(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

// KeyUseSub 多人单次且限定次数
func KeyUseSub(c *ctx.Context, use *CdKeyUse) *errmsg.ErrMsg {
	tx, err := orm.GetMySQL().BeginTx(c, nil)
	if err != nil {
		return errmsg.NewErrorDB(err)
	}
	q1 := "UPDATE cd_key_gen SET limit_cnt=limit_cnt-1, use_cnt=use_cnt+1 WHERE id = ? and is_active = true and limit_cnt > 0"
	if _, err = tx.Exec(q1, use.KeyId); err != nil {
		return errmsg.NewErrCdKeyIsUsed()
	}
	q2 := `INSERT INTO cd_key_use (role_id,batch_id,key_id,created_at) VALUES (?, ?, ?, ?)`
	if _, err = tx.Exec(q2, use.RoleId, use.BatchId, use.KeyId, use.CreatedAt); err != nil {
		return errmsg.NewErrorDB(err)
	}
	if err = tx.Commit(); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}
