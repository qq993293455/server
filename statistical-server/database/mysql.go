package database

import (
	"fmt"

	"coin-server/statistical-server/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var mysqlConn *gorm.DB
var loginMysqlConn *gorm.DB

func InitMysql() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		config.Mysql.Username, config.Mysql.Password, config.Mysql.Addr, config.Mysql.Database)
	var err error
	mysqlConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	loginDsn := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		config.LoginMysql.Username, config.LoginMysql.Password, config.LoginMysql.Addr, config.LoginMysql.Database)
	loginMysqlConn, err = gorm.Open(mysql.Open(loginDsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}
