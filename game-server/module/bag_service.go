package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/game-server/event"
	rulemodel "coin-server/rule/rule-model"
)

type BagService interface {
	InitBagConfig(ctx *ctx.Context)
	GetBagConfig(ctx *ctx.Context) (*pbdao.BagConfig, *errmsg.ErrMsg)
	SaveBagConfig(ctx *ctx.Context, data *pbdao.BagConfig)
	IsBagEnough(ctx *ctx.Context, items map[values.ItemId]values.Integer) (bool, *errmsg.ErrMsg)

	// 获取道具
	GetItem(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (count values.Integer, err *errmsg.ErrMsg)
	GetManyItem(ctx *ctx.Context, roleId values.RoleId, itemsId []values.ItemId) (map[values.ItemId]values.Integer, *errmsg.ErrMsg)
	GetItemPb(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (item *models.Item, err *errmsg.ErrMsg)
	// 添加道具（可添加普通道具、装备、遗物）
	AddItem(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg
	AddManyItem(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) ([]*models.Equipment, *errmsg.ErrMsg)
	AddManyItemPb(ctx *ctx.Context, roleId values.RoleId, items ...*models.Item) *errmsg.ErrMsg
	// 减少道具(只能扣普通道具)
	SubItem(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg
	SubManyItem(ctx *ctx.Context, roleId values.RoleId, items map[values.ItemId]values.Integer) *errmsg.ErrMsg
	SubManyItemPb(ctx *ctx.Context, roleId values.RoleId, items ...*models.Item) *errmsg.ErrMsg
	// 交换道具(只发一个事件)
	ExchangeManyItem(ctx *ctx.Context, roleId values.RoleId, add map[values.ItemId]values.Integer, sub map[values.ItemId]values.Integer) *errmsg.ErrMsg
	// ExchangeManyItemPb(ctx *ctx.Context, roleId values.RoleId, add []*models.Item, sub []*models.Item) *errmsg.ErrMsg
	/*
	 * 开宝箱（不负责扣除及添加开出的道具）
	 * 交换类型是自选宝箱的需要传choose，其他传nil即可
	 */
	UseItemCase(ctx *ctx.Context, itemId values.ItemId, count values.Integer, choose map[values.ItemId]values.Integer) (map[values.ItemId]values.Integer, *errmsg.ErrMsg)
	// 道具更新器、查询器
	RegisterUpdaterById(id values.ItemId, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg)
	RegisterUpdaterByType(typ values.ItemType, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.Integer) *errmsg.ErrMsg)
	RegisterQuerierById(id values.ItemId, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) (count values.Integer, err *errmsg.ErrMsg))
	RegisterExchangerById(id values.ItemId, f func(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId, count values.ItemId) *errmsg.ErrMsg)
	// 装备相关
	// GetEquipBag(ctx *ctx.Context, roleId values.RoleId) ([]*models.Equipment, *errmsg.ErrMsg)
	GenEquip(ctx *ctx.Context, itemId values.ItemId) (*models.Equipment, values.Integer, *errmsg.ErrMsg)
	GetManyEquipBagMap(ctx *ctx.Context, roleId values.RoleId, equipId ...values.EquipId) (map[values.EquipId]*models.Equipment, *errmsg.ErrMsg)
	GetEquipById(ctx *ctx.Context, roleId values.RoleId, equipId values.EquipId) (*models.Equipment, *errmsg.ErrMsg)
	// GetEquipById(ctx *ctx.Context, roleId values.RoleId, equipId values.EquipId) (*models.Equipment, bool, *errmsg.ErrMsg)
	SaveEquipment(ctx *ctx.Context, roleId values.RoleId, equip *models.Equipment) *errmsg.ErrMsg
	SaveManyEquipment(ctx *ctx.Context, roleId values.RoleId, equips []*models.Equipment)
	DelEquipment(ctx *ctx.Context, equipId ...values.EquipId) *errmsg.ErrMsg
	GenOneAffix(entryCfg []rulemodel.EquipEntry, minQuality, quality values.Integer) (*models.Affix, []rulemodel.EquipEntry, *errmsg.ErrMsg)
	GenAffixQuality(qualityWeight map[values.Integer]values.Integer, minQuality values.Integer) (values.Integer, *errmsg.ErrMsg)
	UnlockEquipAffix(ctx *ctx.Context, star values.Integer, equip *models.Equipment) bool
	GetEquipTalentBonus(ctx *ctx.Context, equip *models.Equipment, onlyActive, takeDown bool) map[values.HeroSkillId]values.Level
	CalEquipScore(ctx *ctx.Context, equip *models.Equipment, star, roleLevel values.Integer, equipCfg *rulemodel.Equip, heroId values.HeroId) values.Integer
	// SaveEquipmentBrief(ctx *ctx.Context, roleId values.RoleId, equips ...*models.Equipment)
	GetEquipId(ctx *ctx.Context) (*pbdao.EquipId, *errmsg.ErrMsg)
	SaveEquipId(ctx *ctx.Context, id *pbdao.EquipId)

	// 遗物相关
	GetRelicsById(ctx *ctx.Context, roleId values.RoleId, relicsId values.ItemId) (*models.Relics, *errmsg.ErrMsg)
	UpdateRelics(ctx *ctx.Context, roleId values.RoleId, relics *models.Relics)
	GetRelics(ctx *ctx.Context, roleId values.RoleId) ([]*models.Relics, *errmsg.ErrMsg)
	UpdateMultiRelics(ctx *ctx.Context, roleId values.RoleId, data []*models.Relics)
	GetManyRelics(ctx *ctx.Context, roleId values.RoleId, relicsIds []values.ItemId) ([]*models.Relics, *errmsg.ErrMsg)

	// 吃药
	AutoTakeMedicine(ctx *ctx.Context, roleId values.RoleId, typ values.Integer, mapId values.MapId) *errmsg.ErrMsg
	GetMedicineInfo(ctx *ctx.Context, roleId values.RoleId) (*pbdao.MedicineInfo, *errmsg.ErrMsg)
	GetMedicineMsg(ctx *ctx.Context, roleId values.RoleId, battleMapId values.Integer) (map[values.Integer]*models.MedicineInfo, *errmsg.ErrMsg)
	SaveMedicineInfo(ctx *ctx.Context, info *pbdao.MedicineInfo)

	// 锁
	Lock(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) *errmsg.ErrMsg
	Unlock(ctx *ctx.Context, roleId values.RoleId, itemId values.ItemId) *errmsg.ErrMsg

	// 天赋符文
	GetManyRunes(ctx *ctx.Context, roleId values.RoleId, runeIds []values.RuneId) (map[values.RuneId]*models.TalentRune, *errmsg.ErrMsg)
	SaveManyRunes(ctx *ctx.Context, roleId values.RoleId, runes []*models.TalentRune) (*event.TalentRuneUpdate, *errmsg.ErrMsg)
	GetRuneById(ctx *ctx.Context, roleId values.RoleId, runeId values.RuneId) (*models.TalentRune, *errmsg.ErrMsg)
	SaveRune(ctx *ctx.Context, roleId values.RoleId, rune *models.TalentRune) (*event.TalentRuneUpdate, *errmsg.ErrMsg)
	DelManyRunes(ctx *ctx.Context, roleId values.RoleId, runeIds []values.RuneId) (*event.TalentRuneDestroyed, *errmsg.ErrMsg)
}
