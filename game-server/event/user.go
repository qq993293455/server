package event

import (
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/common/values/enum/AdditionType"
)

type Login struct {
	IsRegister  bool
	UserId      string
	ServerId    int64
	RuleVersion string
	RoleId      string
}

// 接收其他系统发出的增加全局加成事件
type AttrUpdateToRole struct {
	Typ         models.AttrBonusType
	AttrFixed   map[values.AttrId]values.Integer
	AttrPercent map[values.AttrId]values.Integer
	IsCover     bool // 是否覆盖
	HeroId      values.HeroId
}

// 自己发出的属性改变事件
type RoleAttrUpdate struct {
	Typ             models.AttrBonusType
	AttrFixed       []*models.AttrBonus
	AttrPercent     []*models.AttrBonus
	HeroAttrFixed   []*models.HeroAttrBonus
	HeroAttrPercent []*models.HeroAttrBonus
}

type RoleSkillUpdate struct {
	OldSkill values.Integer
	NewSkill values.Integer
}

type RoleSkillUpdateFinish struct {
	Skill []values.HeroSkillId
}

type Logout struct {
}

type UserLevelChange struct {
	Level          values.Level   // 当前服务端实际等级
	Incr           values.Integer // 等级增量
	IsAdvance      bool           // 是否是突破
	LevelIndex     values.Integer // role_lv表的id列
	LevelIndexIncr values.Integer
}

type BattleMapChange struct {
	MapId int64
}

type HeroAttrUpdate struct {
	Data []*HeroAttrUpdateItem
}

type HeroAttrUpdateItem struct {
	IsSkillChange bool
	Hero          *models.Hero
}

type UserCombatValueChange struct {
	RoleId values.RoleId
	Value  values.Integer // 最新的战斗力
}

type BattleSettingChange struct {
	Setting *models.BattleSettingData
}

type RedPointAdd struct {
	RoleId values.RoleId
	Key    string
	Val    values.Integer // 增量
}

type RedPointChange struct {
	RoleId values.RoleId
	Key    string
	Val    values.Integer // 全量
}

type UserTitleChange struct {
	RoleId       values.RoleId
	LastTitle    values.Integer
	CurrentTitle values.Integer
	CombatValue  values.Integer
}

type UserRecentChatAdd struct {
	MyRoleId  values.RoleId
	TarRoleId values.RoleId
}

type ExtraSkillTypAdd struct {
	TypId    models.EntrySkillType // relics_skilltype 表的ID
	LogicId  int64                 // 业务区分Id
	Cnt      int64                 // 增量
	ValueTyp AdditionType.Enum     // 1为正常数值，2为万分比
}

type ExtraSkillTypTotal struct {
	TypId    models.EntrySkillType // relics_skilltype 表的ID
	LogicId  int64                 // 业务区分Id
	TotalCnt int64                 // 全量
	ThisAdd  int64                 // 本次改变的增量
	ValueTyp AdditionType.Enum     // 1为正常数值，2为万分比
}

type BaseAttrAdditionData struct {
	Fixed   map[values.AttrId]values.Integer
	Percent map[values.AttrId]values.Integer
}

type AttrAdditionData struct {
	Base BaseAttrAdditionData
	Hero map[values.HeroId]BaseAttrAdditionData
}

// 充值事件
type RechargeAmountEvt struct {
	Amount int64
}

// 充值成功后写入role后的事件
type RechargeSuccEvt struct {
	Old int64
	New int64
}

func NewAttrAdditionData() *AttrAdditionData {
	return &AttrAdditionData{
		Base: BaseAttrAdditionData{
			Fixed:   make(map[values.AttrId]values.Integer),
			Percent: make(map[values.AttrId]values.Integer),
		},
		Hero: make(map[values.HeroId]BaseAttrAdditionData),
	}
}

func (d *AttrAdditionData) AddFixed(attrId values.AttrId, val values.Integer) {
	d.Base.Fixed[attrId] += val
}

func (d *AttrAdditionData) AddPercent(attrId values.AttrId, val values.Integer) {
	d.Base.Percent[attrId] += val
}

func (d *AttrAdditionData) AddHeroFixed(heroId values.HeroId, attrId values.AttrId, val values.Integer) {
	if _, ok := d.Hero[heroId]; !ok {
		d.Hero[heroId].Fixed[attrId] += val
	} else {
		d.Hero[heroId] = BaseAttrAdditionData{
			Fixed:   map[values.AttrId]values.Integer{attrId: val},
			Percent: map[values.AttrId]values.Integer{},
		}
	}
}

func (d *AttrAdditionData) AddHeroPercent(heroId values.HeroId, attrId values.AttrId, val values.Integer) {
	if _, ok := d.Hero[heroId]; ok {
		d.Hero[heroId].Percent[attrId] += val
	} else {
		d.Hero[heroId] = BaseAttrAdditionData{
			Fixed:   map[values.AttrId]values.Integer{},
			Percent: map[values.AttrId]values.Integer{attrId: val},
		}
	}
}
