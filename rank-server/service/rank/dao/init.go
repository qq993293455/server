package dao

import (
	"coin-server/common/consulkv"
	"coin-server/common/logger"
	"coin-server/common/mysqlclient"
	"coin-server/common/utils"
)

var dao *Dao

func Init(config *consulkv.Config) {
	cfg := &mysqlclient.Config{}
	utils.Must(config.Unmarshal("rank/mysql", cfg))
	mc, err := mysqlclient.NewClient(cfg, logger.DefaultLogger)
	utils.Must(err)
	dao = NewDao(mc.Unsafe())
}

func GetDao() *Dao {
	return dao
}
