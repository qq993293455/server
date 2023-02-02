package orm

import (
	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/mysqlclient"
	"coin-server/common/utils"

	"github.com/jmoiron/sqlx"
)

// TODO 临时

var mysql *sqlx.DB

func InitMySQL(config *consulkv.Config) {
	cfg := &mysqlclient.Config{}
	utils.Must(config.Unmarshal("syncrole/mysql", cfg))
	mc, err := mysqlclient.NewClient(cfg, logger.DefaultLogger)
	utils.Must(err)
	mysql = mc.Unsafe()
}

func GetMySQL() *sqlx.DB {
	return mysql
}
