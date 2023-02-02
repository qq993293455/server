package models

import (
	"time"

	"coin-server/common/utils"

	json "github.com/json-iterator/go"
)

type Launch struct {
	// 必须字段
	Id    int64     `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time  time.Time `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId string    `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`

	// 其它字段
	Xid      string `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	UDId     string `json:"ud_id" gorm:"column:ud_id;type:varchar(64);not null" desc:"ud_id"`
	GameId   string `json:"game_id" gorm:"column:game_id;type:varchar(64);not null" desc:"gameid"`
	Device   string `json:"device" gorm:"column:device;type:varchar(64);not null" desc:"设备号"`
	Progress string `json:"progress" gorm:"column:progress;type:varchar(256);not null" desc:"进度"`
}

func (i *Launch) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *Launch) GetRoleId() []byte {
	return utils.StringToBytes(i.Device)
}

func (i *Launch) Topic() string {
	return LaunchTopic
}

func (i *Launch) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Launch) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,

		i.Xid,
		i.UDId,
		i.GameId,
		i.Device,
		i.Progress,
	}
}

func (i *Launch) NewModel() Model {
	return &Launch{}
}

func (i *Launch) Desc() string {
	return "登录前打点"
}
