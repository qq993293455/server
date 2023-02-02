package models

import (
	"time"

	"coin-server/common/utils"
	"coin-server/common/values"

	json "github.com/json-iterator/go"
)

// 自增主键不能指定类型

type Game struct {
	// 必须字段
	Id       int64           `json:"id" gorm:"column:id;not null;primary_key;AUTO_INCREMENT" desc:"自增id"`
	Time     time.Time       `json:"time" gorm:"column:time;type:datetime;not null" desc:"时间"`
	IggId    string          `json:"igg_id" gorm:"column:igg_id;type:varchar(20);not null" desc:"iggId"`
	ServerId values.ServerId `json:"server_id" gorm:"column:server_id;type:int(11);not null" desc:"服务器id"`

	// 其它字段
	Xid              string `json:"xid" gorm:"column:xid;type:varchar(20);not null;uniqueIndex" desc:"唯一id 用于保证幂等"`
	RoleId           string `json:"role_id" gorm:"column:role_id;type:varchar(20);not null" desc:"角色id"`
	Memory           string `json:"memory" gorm:"column:memory;type:varchar(20);not null" desc:"设备内存大小"`
	StoryChapter     string `json:"story_chapter" gorm:"column:story_chapter;type:varchar(20);not null" desc:"对话所属章节"`
	SkipChapter      string `json:"skip_chapter" gorm:"column:skip_chapter;type:varchar(20);not null" desc:"对话所属章节跳过"`
	TutorialStep     string `json:"tutorial_step" gorm:"column:tutorial_step;type:varchar(20);not null" desc:"教程成功操作"`
	TutorialSkipStep string `json:"tutorial_skip_step" gorm:"column:tutorial_skip_step;type:varchar(20);not null" desc:"教程跳过操作"`
}

func (i *Game) TableName() string {
	t := i.Time.UTC().Format("20060102")
	return i.Topic() + "_" + t
}

func (i *Game) GetRoleId() []byte {
	return utils.StringToBytes(i.RoleId)
}

func (i *Game) Topic() string {
	return GameTopic
}

func (i *Game) ToJson() ([]byte, error) {
	return json.Marshal(i)
}

func (i *Game) ToArgs() []interface{} {
	// 这里必须和结构体定义的顺序一致 自增id不填
	return []interface{}{
		i.Time,
		i.IggId,
		i.ServerId,

		i.Xid,
		i.RoleId,
		i.Memory,
		i.StoryChapter,
		i.SkipChapter,
		i.TutorialStep,
		i.TutorialSkipStep,
	}
}

func (i *Game) NewModel() Model {
	return &Game{}
}

func (i *Game) Desc() string {
	return "游戏事件"
}
