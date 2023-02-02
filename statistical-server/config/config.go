package config

import (
	"coin-server/common/consulkv"
	"coin-server/common/utils"
)

var (
	Kafka      *KafkaConfig
	Mysql      *MysqlConfig
	LoginMysql *MysqlConfig
)

type KafkaConfig struct {
	Addr     []string `json:"addr"`
	MinBytes int      `json:"min_bytes"`
	MaxBytes int      `json:"max_bytes"`
}

type MysqlConfig struct {
	Addr     string `json:"addr" mapstructure:"addr"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
	Database string `json:"database" mapstructure:"database"`
}

func Init(config *consulkv.Config) {
	Kafka = &KafkaConfig{}
	Mysql = &MysqlConfig{}
	LoginMysql = &MysqlConfig{}

	utils.Must(config.Unmarshal("statistical2/kafka", Kafka))
	utils.Must(config.Unmarshal("statistical2/mysql", Mysql))
	utils.Must(config.Unmarshal("statistical2/login_mysql", LoginMysql))
}
