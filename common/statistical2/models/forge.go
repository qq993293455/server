package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

type Forge struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid     string         `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId  values.RoleId  `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	Quality values.Quality `json:"quality" gorm:"column:quality;type:int(11);not null" desc:"品质"`
	EquipId values.EquipId `json:"equip_id" gorm:"column:equip_id;type:varchar(20);not null" desc:"装备唯一id"`
	ItemId  values.ItemId  `json:"item_id" gorm:"column:item_id;type:int(11);not null" desc:"装备的item id"`
}

func (f *Forge) TableName() string {
	t := f.Time.UTC().Format("20060102")
	return f.Topic() + "_" + t
}

func (f *Forge) GetRoleId() []byte {
	return utils.StringToBytes(f.RoleId)
}

func (f *Forge) Topic() string {
	return ForgeTopic
}

func (f *Forge) ToJson() ([]byte, error) {
	return json.Marshal(f)
}

func (f *Forge) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致
	return []interface{}{
		f.Time,
		f.IggId,
		f.ServerId,

		f.Xid,
		f.RoleId,
		f.Quality,
		f.EquipId,
		f.ItemId,
	}
}
func (f *Forge) NewModel() Model {
	return &Forge{}
}

func (f *Forge) Desc() string {
	return "装备打造"
}
