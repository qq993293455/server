package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

type Pay struct {
	// 必须字段
	Id   int64     `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time time.Time `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`

	// 其它字段
	Xid        string          `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	Sn         string          `json:"sn" gorm:"column:sn;type:varchar(64);not null" desc:"sn"`
	RoleId     string          `json:"role_id" gorm:"column:role_id;type:varchar(64);not null" desc:"role_id"`
	ServerId   values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`
	PcId       int64           `json:"pc_id" gorm:"column:pc_id;type:int(11);not null" desc:"pc_id"`
	PaidTime   int64           `json:"paid_time" gorm:"column:paid_time;type:int(11);not null" desc:"支付时间"`
	ExpireTime int64           `json:"expire_time" gorm:"column:expire_time;type:int(11);not null" desc:"过期时间"` // 仅订阅功能有效
}

func (i *Pay) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *Pay) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *Pay) Topic() string {
	return PayTopic
}

func (i *Pay) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Pay) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,

		i.Xid,
		i.Sn,
		i.RoleId,
		i.ServerId,
		i.PcId,
		i.PaidTime,
		i.ExpireTime,
	}
}

func (i *Pay) NewModel() Model {
	return &Pay{}
}

func (i *Pay) Desc() string {
	return "支付成功记录"
}
