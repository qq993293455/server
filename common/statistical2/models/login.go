package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

type Login struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid           string        `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId        values.RoleId `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	UserId        string        `json:"user_id" gorm:"column:user_id;type:varchar(20);not null" desc:"用户id"`
	DeviceId      string        `json:"device_id" gorm:"column:device_id;type:varchar(128);not null" desc:"设备id"`
	IP            string        `json:"ip" gorm:"column:ip;type:varchar(128);not null" desc:"用户ip地址"`
	RuleVersion   string        `json:"rule_version" gorm:"column:rule_version;type:varchar(20);not null" desc:"规则版本"`
	GameId        string        `json:"game_id"  gorm:"column:game_id;type:varchar(20);not null" desc:"游戏版本"`
	ClientVersion string        `json:"client_version" gorm:"column:client_version;type:varchar(20);not null" desc:"客户端版本"`
}

func (l *Login) TableName() string {
	t := l.Time.UTC().Format("20060102")
	return l.Topic() + "_" + t
}

func (l *Login) GetRoleId() []byte {
	return utils.StringToBytes(l.RoleId)
}

func (l *Login) Topic() string {
	return LoginTopic
}

func (l *Login) ToJson() ([]byte, error) {
	return json.Marshal(l)
}

func (l *Login) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		l.Time,
		l.IggId,
		l.ServerId,

		l.Xid,
		l.RoleId,
		l.UserId,
		l.DeviceId,
		l.IP,
		l.RuleVersion,
		l.GameId,
		l.ClientVersion,
	}
}

func (l *Login) NewModel() Model {
	return &Login{}
}

func (l *Login) Desc() string {
	return "登录"
}
