package dao

import (
	"context"

	"coin-server/common/logger"
	"github.com/go-sql-driver/mysql"

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

func (d *Dao) BatchSave(ctx context.Context, data []Role) error {

	query := `INSERT INTO roles (role_id, igg_id, nickname, level, avatar, avatar_frame, power, highest_power, title, language, login_time,logout_time, create_time)
VALUES (:role_id,:igg_id, :nickname, :level, :avatar, :avatar_frame, :power, :highest_power, :title, :language, :login_time, :logout_time, :create_time)
ON DUPLICATE KEY UPDATE igg_id = VALUES(igg_id), nickname = VALUES(nickname), LEVEL = VALUES(LEVEL), avatar = VALUES(avatar), avatar_frame = VALUES(avatar_frame),
                        power = VALUES(power), highest_power = VALUES(highest_power), title = VALUES(title), language = VALUES(language), login_time = VALUES(login_time), logout_time = VALUES(logout_time);`

	_, err := d.db.NamedExec(query, data)
	if err != nil {
		d.log.Warn("db error", zap.Error(err))
		if errMySQL, ok := err.(*mysql.MySQLError); ok {
			if errMySQL.Number == 1062 {
				return nil
			}
		}
	}

	return err
}
