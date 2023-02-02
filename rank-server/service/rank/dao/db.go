package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/values/enum"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Dao struct {
	db  *sqlx.DB
	log *logger.Logger
}

func NewDao(db *sqlx.DB) *Dao {
	return &Dao{db: db, log: logger.DefaultLogger}
}

func (d *Dao) BatchSave(ctx *ctx.Context, data []RankValue) *errmsg.ErrMsg {
	query := "INSERT INTO `rank`(rank_id,rank_type,owner_id,value1,value2,value3,value4,extra,shard,deleted_at)"
	query += "VALUES(:rank_id,:rank_type,:owner_id,:value1,:value2,:value3,:value4,:extra,:shard,:deleted_at) ON DUPLICATE KEY "
	query += "UPDATE value1=VALUES(value1),value2=VALUES(value2),value3=VALUES(value3) ,value4=VALUES(value4),extra=VALUES(extra),shard=VALUES(shard),deleted_at=VALUES(deleted_at);"

	_, err := d.db.NamedExec(query, data)
	if err != nil {
		d.log.Error("update rank err", zap.Error(err))
		return errmsg.NewErrorDB(err)
	}

	return nil
}

func (d *Dao) Get(rankType enum.RankType, start, end int) ([]*RankValue, *errmsg.ErrMsg) {
	query := "SELECT rank_id,rank_type,owner_id,value1,value2,value3,value4,extra,shard,deleted_at FROM `rank` WHERE rank_type=? AND deleted_at=0 LIMIT ?,?"
	data := make([]*RankValue, 0)
	err := d.db.Select(&data, query, rankType, start, end)
	if err != nil {
		d.log.Error("get rank err", zap.Error(err))
		return nil, errmsg.NewErrorDB(err)
	}
	return data, nil
}
