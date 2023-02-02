package model

import (
	"context"
	"fmt"

	"coin-server/common/ctx"
	"coin-server/common/orm"
	"coin-server/common/pay"
	"coin-server/common/proto/dao"
	utils2 "coin-server/pikaviewer/utils"
)

type Pay struct {
	Id        int64  `db:"id" json:"id"`
	RoleId    string `db:"role_id" json:"role_id"`
	SN        string `db:"sn" json:"sn"`
	PcId      int64  `db:"pc_id" json:"pc_id"`
	IggId     string `db:"igg_id" json:"igg_id"`
	PaidTime  int64  `db:"paid_time" json:"paid_time"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

func NewPay() *Pay {
	return &Pay{}
}

func (p *Pay) Save() error {
	query := `INSERT INTO pay (role_id,sn,pc_id,igg_id,paid_time,created_at)
			VALUES (:role_id,:sn,:pc_id,:igg_id,:paid_time,:created_at);`

	if _, err := orm.GetMySQL().NamedExec(query, p); err != nil {
		return err
	}
	return nil
}

func (p *Pay) SnExist(roleId, sn string) (bool, error) {
	// query := `SELECT COUNT(*) num FROM pay where sn = ? LIMIT 1`
	// var count int
	// if err := orm.GetMySQL().Get(&count, query, sn); err != nil {
	// 	return false, err
	// }
	// return count > 0, nil
	out := &dao.Pay{Sn: sn}
	ok, err := orm.GetOrm(ctx.GetContext()).HGetPB(pay.GetPayRedis(), pay.GetKey(roleId), out)
	if err != nil {
		return false, utils2.NewDefaultErrorWithMsg(err.Error())
	}
	return ok, nil
}

func (p *Pay) Save2Queue(data *dao.PayQueue) error {
	v := data.ToSave()
	fmt.Println(string(v))
	return pay.GetPayQueueRedis().RPush(context.Background(), pay.PayQueueKey, string(data.ToSave())).Err()
}
