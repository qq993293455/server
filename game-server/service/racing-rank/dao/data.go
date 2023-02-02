package dao

import (
	"context"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetData(ctx *ctx.Context) (*dao.RacingRankData, *errmsg.ErrMsg) {
	data := &dao.RacingRankData{RoleId: ctx.RoleId}
	_, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SaveData(ctx *ctx.Context, data *dao.RacingRankData) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), data)
}

func SaveDataImmediately(data *dao.RacingRankData) *errmsg.ErrMsg {
	o := orm.GetOrm(context.Background())
	o.SetPB(redisclient.GetDefaultRedis(), data)
	return o.Do()
}

type Data struct {
	RoleId       values.RoleId  `db:"role_id" json:"role_id"`
	HighestPower values.Integer `db:"highest_power" json:"highest_power"`
	LoginTime    values.Integer `db:"login_time" json:"login_time"`
}

func Find(combatValue values.Integer, lt bool) ([]*Data, error) {
	var where string
	if lt {
		where = ` > ? ORDER BY highest_power ASC`
	} else {
		where = ` <= ? ORDER BY highest_power DESC`
	}
	where += ` LIMIT 100`
	query := `SELECT role_id,highest_power,login_time FROM roles WHERE highest_power ` + where
	dest := make([]*Data, 0)
	if err := orm.GetMySQL().Select(&dest, query, combatValue); err != nil {
		return nil, err
	}
	return dest, nil
}
