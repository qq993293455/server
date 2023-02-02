package dao

import (
	"context"

	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/orm"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"

	"github.com/jmoiron/sqlx"
)

type Dao struct {
	db  *sqlx.DB
	log *logger.Logger
}

func NewDao(db *sqlx.DB) *Dao {
	return &Dao{db: db, log: logger.DefaultLogger}
}

func (d *Dao) Find(combatValue values.Integer, lt bool) ([]*Data, error) {
	var where string
	if lt {
		where = ` > ? ORDER BY highest_power ASC`
	} else {
		where = ` <= ? ORDER BY highest_power DESC`
	}
	where += ` LIMIT 100`
	query := `SELECT role_id,highest_power,login_time FROM roles WHERE highest_power ` + where
	dest := make([]*Data, 0)
	if err := d.db.Select(&dest, query, combatValue); err != nil {
		return nil, err
	}
	return dest, nil
}

// func (d *Dao) SaveEndTime(data EndTime) {
// 	query := `INSERT INTO racing_rank (role_id, end_time)
// 			VALUES (:role_id, :end_time)
// 			ON DUPLICATE KEY UPDATE end_time = VALUES(end_time);`
//
// 	if _, err := d.db.NamedExec(query, data); err != nil {
// 		d.log.Warn("SaveEndTime error", zap.Error(err), zap.String("role_id", data.RoleId), zap.Int64("end_time", data.EndTime))
// 	}
// }
//
// func (d *Dao) DeleteEndTime(roleId values.RoleId) {
// 	query := `DELETE FROM racing_rank WHERE role_id=?`
// 	if _, err := d.db.Exec(query, roleId); err != nil {
// 		d.log.Warn("DeleteEndTime error", zap.Error(err), zap.String("role_id", roleId))
// 	}
// }
//
// func (d *Dao) BatchGetEndTime(start, end int) ([]*EndTime, error) {
// 	query := `SELECT role_id,end_time FROM racing_rank ORDER BY id LIMIT ?,?;`
// 	dest := make([]*EndTime, 0)
// 	if err := d.db.Select(&dest, query, start, end); err != nil {
// 		return nil, err
// 	}
// 	return dest, nil
// }

func (d *Dao) SaveData(data *pbdao.RacingRankData) *errmsg.ErrMsg {
	o := orm.GetOrm(context.Background())
	o.SetPB(redisclient.GetDefaultRedis(), data)
	return o.Do()
}

func (d *Dao) GetRacingRankData(roleId values.RoleId) (*pbdao.RacingRankData, *errmsg.ErrMsg) {
	data := &pbdao.RacingRankData{RoleId: roleId}
	_, err := orm.GetOrm(context.Background()).GetPB(redisclient.GetDefaultRedis(), data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func (d *Dao) BatchGetRole(roleIds []values.RoleId) (map[values.RoleId]*pbdao.Role, *errmsg.ErrMsg) {
	ri := make([]orm.RedisInterface, len(roleIds))
	for idx := range ri {
		ri[idx] = &pbdao.Role{RoleId: roleIds[idx]}
	}
	notFound, err := orm.GetOrm(context.Background()).MGetPB(redisclient.GetDefaultRedis())
	if err != nil {
		return nil, err
	}
	notFoundMap := make(map[int]struct{}, len(notFound))
	for _, i := range notFound {
		notFoundMap[i] = struct{}{}
	}
	data := make(map[values.RoleId]*pbdao.Role, 0)
	for i := range ri {
		if _, ok := notFoundMap[i]; ok {
			continue
		}
		role := ri[i].(*pbdao.Role)
		data[role.RoleId] = role
	}
	return data, nil
}
