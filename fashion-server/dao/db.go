package dao

import (
	"coin-server/common/logger"
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

func (d *Dao) Exist(roleId values.RoleId, fashionId values.FashionId) (bool, error) {
	query := `SELECT COUNT(*) num FROM fashion WHERE role_id = ? AND fashion_id = ?`
	var count int
	err := d.db.Get(&count, query, roleId, fashionId)
	return count > 0, err
}

func (d *Dao) Save(f *Fashion) error {
	query := `INSERT INTO fashion (role_id,user_id,server_id, hero_id,fashion_id,expired_at,created_at)
			VALUES (:role_id,:user_id,:server_id, :hero_id,:fashion_id,:expired_at,:created_at)`

	_, err := d.db.NamedExec(query, f)
	return err
}

func (d *Dao) UpdateExpire(roleId values.RoleId, fashionId values.FashionId, expire values.Integer) error {
	query := `UPDATE fashion SET expired_at = ? WHERE role_id = ? AND fashion_id = ?`
	_, err := d.db.Exec(query, expire, roleId, fashionId)
	return err
}

func (d *Dao) Delete(roleId values.RoleId, fashionId values.FashionId) error {
	query := `DELETE FROM fashion WHERE role_id=? AND fashion_id = ?`
	_, err := d.db.Exec(query, roleId, fashionId)
	return err
}

func (d *Dao) BatchGet(start, end int) ([]*Fashion, error) {
	query := `SELECT role_id,user_id,server_id, hero_id,fashion_id,expired_at FROM fashion ORDER BY id LIMIT ?,?;`
	dest := make([]*Fashion, 0)
	if err := d.db.Select(&dest, query, start, end); err != nil {
		return nil, err
	}
	return dest, nil
}
