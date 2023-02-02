package mysqlclient

import (
	"database/sql"
	"time"

	"coin-server/common/errmsg"
	"coin-server/common/logger"

	"go.uber.org/zap"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Client struct {
	db  *sqlx.DB
	log *logger.Logger
}

type Config struct {
	User     string `json:"user"`
	Addr     string `json:"addr"`
	Pwd      string `json:"pwd"`
	Database string `json:"database"`
}

func NewClient(c *Config, log *logger.Logger) (*Client, error) {
	cnf := &mysql.Config{
		User:                 c.User,
		Passwd:               c.Pwd,
		Net:                  "tcp",
		Addr:                 c.Addr,
		DBName:               c.Database,
		Params:               nil,
		Loc:                  time.Local,
		Timeout:              time.Second * 3,
		ParseTime:            true,
		AllowNativePasswords: true,
		AllowOldPasswords:    true,
	}
	dsn := cnf.FormatDSN()
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(10)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(time.Second * 60)
	db.SetMaxOpenConns(125)
	db.SetConnMaxLifetime(time.Second * 300)
	sqlLog := log.With(zap.Fields(zap.String("addr", c.Addr), zap.String("database", c.Database)))
	return &Client{db: db, log: sqlLog}, nil
}

func (this_ *Client) Unsafe() *sqlx.DB {
	return this_.db
}

func (this_ *Client) WithTxx(f func(tx *sqlx.Tx) error) (outE *errmsg.ErrMsg) {
	var tx *sqlx.Tx
	var err error
	tx, err = this_.db.Beginx()
	if err != nil {
		this_.log.Error("WithTxx: mysql beginx error", zap.Error(err))
		return errmsg.NewErrorDB(err)
	}
	defer func() {
		if outE != nil {
			_ = tx.Rollback()
			this_.log.Error("WithTxx: error", zap.Error(outE))
		}
	}()
	err = f(tx)
	if err != nil {
		if e, ok := err.(*errmsg.ErrMsg); ok {
			if e != nil {
				return e
			}
			return errmsg.NewErrorDB(tx.Commit())
		}
		return errmsg.NewErrorDB(err)
	}
	return errmsg.NewErrorDB(tx.Commit())
}

func (this_ *Client) QueryRow(f func(row *sqlx.Row) error, query string, args ...interface{}) (*errmsg.ErrMsg, bool) {
	row := this_.db.QueryRowx(query, args...)
	err := f(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		this_.log.Error("query error", zap.Error(err), zap.String("query", query), zap.Any("args", args))
		return errmsg.NewErrorDB(err), false
	}
	return nil, true
}

func (this_ *Client) Query(f func(rows *sqlx.Rows) error, query string, args ...interface{}) *errmsg.ErrMsg {
	rows, err := this_.db.Queryx(query, args...)
	if err != nil {
		this_.log.Error("query error", zap.Error(err), zap.String("query", query), zap.Any("args", args))
		return errmsg.NewErrorDB(err)
	}
	for rows.Next() {
		err = f(rows)
		if err != nil {
			this_.log.Error("query scan error", zap.Error(err), zap.String("query", query), zap.Any("args", args))
			return errmsg.NewErrorDB(err)
		}
	}
	return nil
}

func (this_ *Client) Exec(query string, args ...interface{}) (int64, int64, *errmsg.ErrMsg) {
	r, err := this_.db.Exec(query, args...)
	if err != nil {
		this_.log.Error("exec error", zap.Error(err), zap.String("query", query), zap.Any("args", args))
		return 0, 0, errmsg.NewErrorDB(err)
	}
	affectRows, _ := r.RowsAffected()
	lastInsertId, _ := r.LastInsertId()
	return affectRows, lastInsertId, nil
}
