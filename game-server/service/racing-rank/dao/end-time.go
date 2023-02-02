package dao

import (
	"database/sql"

	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/values"
)

func GetEndTime(roleId values.RoleId) (values.Integer, *errmsg.ErrMsg) {
	query := `SELECT end_time FROM racing_rank WHERE role_id = ?`
	var endTime int64
	err := orm.GetMySQL().Get(&endTime, query, roleId)
	if err != nil {
		if err == sql.ErrNoRows {
			return endTime, nil
		}
		return 0, errmsg.NewErrorDB(err)
	}
	return endTime, nil
}

func DeleteEndTime(roleId values.RoleId) *errmsg.ErrMsg {
	query := `DELETE FROM racing_rank WHERE role_id = ?`
	if _, err := orm.GetMySQL().Exec(query, roleId); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}
