package dao

import (
	"time"

	"coin-server/common/logger"
	"github.com/jmoiron/sqlx"
)

type Dao struct {
	db  *sqlx.DB
	log *logger.Logger
}

func NewDao(db *sqlx.DB) *Dao {
	return &Dao{db: db, log: logger.DefaultLogger}
}

func (d *Dao) Find() ([]*Notice, error) {
	query := `SELECT id,title,content,reward_content,is_custom,begin_at,expired_at,created_at FROM notice WHERE expired_at >=?`
	dest := make([]*Notice, 0)
	if err := d.db.Select(&dest, query, time.Now().Unix()); err != nil {
		return nil, err
	}
	return dest, nil
}

func (d *Dao) Save(data *Notice) error {
	query := `INSERT INTO notice (id,title,content,reward_content,is_custom,begin_at,expired_at,created_at)
			VALUES (:id,:title,:content,:reward_content,:is_custom,:begin_at,:expired_at,:created_at)
			ON DUPLICATE KEY UPDATE title = VALUES(title),content = VALUES(content),reward_content = (reward_content),
			is_custom = VALUES(is_custom),begin_at = VALUES(begin_at),expired_at = VALUES(expired_at),created_at = VALUES(created_at);`

	_, err := d.db.NamedExec(query, data)
	return err
}

func (d *Dao) Delete(id string) error {
	query := `DELETE FROM notice WHERE id=?`
	_, err := d.db.Exec(query, id)
	return err
}
