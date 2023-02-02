package dao

import (
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/values"
)

type Model struct {
	Id   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func NameExist(name string) (bool, *errmsg.ErrMsg) {
	query := `SELECT COUNT(*) num FROM guild WHERE name=?;`
	var count int
	if err := orm.GetMySQL().Get(&count, query, name); err != nil {
		return false, errmsg.NewErrorDB(err)
	}
	return count > 0, nil
}

func SaveName(data *Model) *errmsg.ErrMsg {
	query := `INSERT INTO guild (id,name) VALUES (:id, :name)
	ON DUPLICATE KEY UPDATE name = VALUES(name);`
	if _, err := orm.GetMySQL().NamedExec(query, data); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}

func DeleteName(id values.GuildId) *errmsg.ErrMsg {
	query := `DELETE FROM guild WHERE id = ?`
	if _, err := orm.GetMySQL().Exec(query, id); err != nil {
		return errmsg.NewErrorDB(err)
	}
	return nil
}
