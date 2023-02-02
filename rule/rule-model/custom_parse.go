package rule_model

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"coin-server/common/proto/models"
	"coin-server/common/utils"
	wr "coin-server/common/utils/weightedrand"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/rule/factory/condition"
	tasktarget "coin-server/rule/factory/task-target"

	"github.com/igrmk/treemap/v2"
)

type CustomParse struct {
	achievementCnt    map[values.AchievementId]map[values.Integer]*AchievementList
	gitItem           map[values.Integer][]GiftItem
	exchangeWeight    map[values.ItemId]*CustomExchangeWeight
	criticalRatio     float64
	maxGuildLevel     values.Level
	guildPositionList []values.GuildPosition
	npcTaskUnlock     map[models.TaskType]map[values.Integer]map[values.Integer][]values.TaskId // 类型->目标id->目标数量->可解锁的任务
	mainTaskChapter   map[values.Integer][]MainTaskChapterReward
	atlasByType       map[values.Integer]map[values.Integer]AtlasItem
	shopMap           map[values.Integer][]GoodslistLv
	dropListMap       map[values.Integer][]DropListsMini
	dropMap           map[values.Integer][]DropMini
	// heroLvUp          map[values.HeroId][]HeroLvup
	maxRoleLevel                      values.Level
	divination                        CustomDivination
	anecdotesPiece                    map[values.StoryId][]AnecdotesClassText
	anecdotesGame1                    map[int64]map[int64]AnecdotesGame1Option
	anecdotesGame3                    map[int64]map[int64]AnecdotesGame3Option
	anecdotesGame4                    map[int64]map[int64]AnecdotesGame4Option
	anecdotesGame5                    map[int64]map[int64]AnecdotesGame5Option
	anecdotesGame7                    map[int64]map[int64]AnecdotesGame7Option
	anecdotesGame8                    map[int64]map[int64]AnecdotesGame8Option
	attrTransGroup                    map[values.Integer][]AttrTrans
	maxEquipStar                      values.Level
	npcDialogOpt                      map[values.Integer][]DialogueOpt
	equipEntryGroup                   map[values.Integer][]EquipEntry
	maxEquipMeltLevel                 map[values.EquipSlot]values.Level
	equipRefine                       map[values.EquipSlot]map[values.Level]*EquipRefine
	maxEquipForgeLevel                values.Level
	equipForgeLeveGroup               map[values.Level][]ForgeLevel
	equipForgeMaxLevel                values.Integer // 打造装备打造等级最高可以加成的装备等级
	relicsLevelUpCost                 map[values.Integer][]RelicsLvQuality
	mapSelect                         *treemap.TreeMap[int64, *MapSelect]
	mapSelectScene                    map[int64]int64
	mapEventAutoRef                   []AnecdotesGames
	gachaUnlockCond                   map[values.Integer]map[values.Integer][]values.Integer
	gachaWeightParse                  map[values.GachaId]*CustomGachaWeight
	defaultTalent                     map[values.Integer][]Talent
	eachTalent                        map[values.Integer][]Talent
	taskTypAchievement                map[models.TaskType]map[values.AchievementId]values.Integer
	npcDialogRelation                 CustomNpcDialogRelationMap
	initRowHeroTalentMap              map[values.HeroId]values.HeroId // 这里的heroId为row_hero表里的id列
	medicineSortedMap                 map[values.Integer][]Medicament
	renameCost                        [][2]values.Integer
	originHead                        [2]values.Integer
	sysName2Type                      map[string]models.SystemType
	default_tower_cost                [][2]values.Integer
	originTalentPoint                 values.Integer
	maxTitle                          values.Integer
	robotMap                          map[values.Integer][]Robot
	deriveHeroMap                     map[values.HeroId][]values.HeroId
	roguelikeBossGroup                map[values.Integer][]values.Integer
	roguelikeEntryGroup               map[values.Integer][]values.Integer
	biographyGroupByHero              map[values.HeroId][]Biography
	biographyGroupByTaskType          map[models.TaskType][]Biography
	expeditionQuantityGroupByTaskType map[models.TaskType][]ExpeditionQuantity
	expeditionGroupByTaskType         map[models.TaskType]map[values.Quality][]Expedition
	soulContractGroupByHero           map[values.HeroId][]SoulContract
	maxSoulContract                   map[values.HeroId]MaxSoulContract
	maxExpSkipCount                   values.Integer
	shopTypMap                        map[values.Integer][]values.Integer
	allTargetTaskTypes                map[models.TaskType]struct{}
	allMainTaskTypes                  map[models.TaskType]struct{}
	allNpcTaskTypes                   map[models.TaskType]struct{}
	allLoopTaskTypes                  map[models.TaskType]struct{}
	maxSkillId                        map[values.HeroSkillId]values.HeroSkillId
	originHeroMap                     map[values.HeroId]values.Integer
	beginningHeroes                   []int64
	beginningMap                      map[int64]int64
	targetTaskStageMap                map[values.Integer][]TargetTaskStage
	allLimitedTimePackConditions      map[models.TaskType]map[values.Integer]struct{}
	limitedTimePackagePayList         map[values.Integer][]ActivityLimitedtimePackagePay
	dailySaleLevelUpdate              []int64
	dailySaleChargeUpdate             []int64
	weeklySaleLevelUpdate             []int64
	weeklySaleChargeUpdate            []int64
	equipResonanceGroupByHero         map[values.HeroId]map[values.Level]EquipResonance
	expeditionRewardGroupByParent     map[values.Integer][]ExpeditionReward
	personalBossMonsterAttr           map[int64]MonsterAttr
	personalBossLvGroupByFloor        map[int64][]PersonalBossLv
	bossHallOpenTime                  [][]int64
	defaultFashion                    map[values.HeroId]values.FashionId
	relicsFuncAttr                    map[models.TaskType][]values.Integer
	soulSkill                         map[values.HeroId]values.HeroSkillId
	equipStarCost                     []values.ItemId // 强化装备需要消耗道具（所有）
	maxRacingRanking                  values.Integer
	maxBlessingPage                   values.Integer
}

func NewCustomParse() CustomParse {
	return CustomParse{}
}

func ParseCustom() {
	r.CustomParse.gitItem = customParseGiftItem()
	r.CustomParse.exchangeWeight = customParseExchangeWeight()
	r.CustomParse.criticalRatio = customParseGlobalNeeded()
	r.CustomParse.maxGuildLevel = customParseGuildLevel()
	r.CustomParse.achievementCnt = customParseAchievementMap()
	r.CustomParse.taskTypAchievement = customParseTaskAchievement()
	r.CustomParse.npcTaskUnlock = customParseNpcTaskCondition()
	r.CustomParse.guildPositionList = customParseGuildPosition()
	r.CustomParse.mainTaskChapter = customParseMainTaskChapter()
	r.CustomParse.atlasByType = customParseAtlasType()
	r.CustomParse.shopMap = customParseShopMap()
	r.CustomParse.shopTypMap = customParseShopTyp()
	r.dropListMap = customParseDropList()
	r.dropMap = customParseDrop()
	// r.heroLvUp = customParseHeroLvUp()
	r.maxRoleLevel = customParseMaxRoleLevel()
	r.divination = customParseDivination()
	r.anecdotesPiece = customParseAnecdotes()
	r.anecdotesGame1 = customParseAnecdoteGame1()
	r.anecdotesGame3 = customParseAnecdoteGame3()
	r.anecdotesGame4 = customParseAnecdoteGame4()
	r.anecdotesGame5 = customParseAnecdoteGame5()
	r.anecdotesGame7 = customParseAnecdoteGame7()
	r.anecdotesGame8 = customParseAnecdoteGame8()
	r.maxEquipStar = customParseMaxEquipStar()
	r.attrTransGroup = customParseAttrTrans()
	r.npcDialogOpt = customParseDialog()
	r.equipEntryGroup = customParseEquipEntry()
	r.maxEquipMeltLevel, r.equipRefine = customParseMaxEquipRefine()
	r.maxEquipForgeLevel = customParseMaxEquipForgeLevel()
	r.equipForgeLeveGroup, r.equipForgeMaxLevel = customParseEquipForgeLevelGroup()
	r.CustomParse.relicsLevelUpCost = customParseRelicsCost()
	r.mapSelect, r.mapSelectScene = customParseMapSelect()
	r.mapEventAutoRef = customParseAutoRefreshGame()
	r.CustomParse.gachaUnlockCond = customParseGachaUnlock()
	r.CustomParse.gachaWeightParse = customParseGachaWeight()
	r.defaultTalent, r.eachTalent = customParseDefaultTalent()
	r.npcDialogRelation = customParseNpcDialogRelation()
	r.originHeroMap, r.initRowHeroTalentMap, r.deriveHeroMap = customParseInitRowHeroTalent()
	r.medicineSortedMap = customParseMedicineMap()
	r.renameCost = customParseRenameCost()
	r.originHead, r.originTalentPoint = customParseOriginHead()
	r.sysName2Type = customParseSysType()
	r.default_tower_cost = customParaesDefaultTowerCost()
	r.maxTitle = customParseMaxTitle()
	r.robotMap = customParseRobot()
	r.roguelikeBossGroup = customParseRoguelikeBossGroup()
	r.roguelikeEntryGroup = customParseRoguelikeEntryGroup()
	r.biographyGroupByHero, r.biographyGroupByTaskType = customParseBiography()
	r.expeditionQuantityGroupByTaskType = customParseExpeditionQuantity()
	r.expeditionGroupByTaskType = customParseExpedition()
	r.soulContractGroupByHero, r.maxSoulContract = customParseSoulContract()
	r.maxExpSkipCount = customParseExpSkip()
	r.allTargetTaskTypes = customParseAllTargetTaskTypes()
	r.allMainTaskTypes = customParseAllMainTaskTypes()
	r.allNpcTaskTypes = customParseAllNpcTaskTypes()
	r.allLoopTaskTypes = customParseAllLoopTaskTypes()
	r.allLimitedTimePackConditions = customParseAllLimitedTimePackageConditions()
	r.limitedTimePackagePayList = customParseLimitedTimePackagePayList()
	r.maxSkillId = customParseSkill()
	r.beginningHeroes = customBeginningHeroes()
	r.beginningMap = customBeginningMap()
	r.targetTaskStageMap = customParseTargetTaskStageList()
	r.dailySaleLevelUpdate, r.dailySaleChargeUpdate, r.weeklySaleLevelUpdate, r.weeklySaleChargeUpdate = customDailySaleUpdateLine()
	r.equipResonanceGroupByHero = customParseEquipResonance()
	r.personalBossMonsterAttr = customParsePersonalBossAttr()
	r.personalBossLvGroupByFloor = customParsePersonalBossLv()
	// r.expeditionRewardGroupByParent = customParseExpeditionReward()
	r.defaultFashion = customParseFashion()
	r.relicsFuncAttr = customParseRelicsAttr()
	r.soulSkill = customParseSoulSkill()
	r.expeditionRewardGroupByParent = customParseExpeditionReward()
	r.bossHallOpenTime = parseBossHallOpenTime()
	r.equipStarCost = customParseEquipStar()
	r.maxRacingRanking = customParseRacingRankReward()
	r.maxBlessingPage = customParseGuildBlessing()

	parseCounter()
	check()
}

func check() {
	equipConfigCheck()
	guildCfgCheck()
	foreFixedRecipeCheck()
	relicsCfgCheck()
	exchangeCfgCheck()
	equipCompleteCheck()
	rowHeroCheck()
	combatBattleCheck()
	expeditionCostRecoveryCheck()
}

func (c *CustomParse) GetBeginHeroes() [][]int64 {
	out := make([][]int64, 0, 2)
	for _, v := range c.beginningHeroes {
		rh, ok := GetReader().RowHero.GetRowHeroById(v)
		if !ok {
			panic("begining heroes not found")
		}
		out = append(out, []int64{v, rh.OriginId})
	}
	return out
}

func (c *CustomParse) GetBeginningMap() map[int64]int64 {
	return c.beginningMap
}

func (*TargetTaskStage) ListByParent(parentId values.Integer) []TargetTaskStage {
	return r.CustomParse.targetTaskStageMap[parentId]
}
func parseBossHallOpenTime() [][]int64 {
	str, ok := GetReader().KeyValue.GetString("DevilSecretOpenTime")
	if !ok {
		panic("KeyValue:DevilSecretOpenTime not found")
	}
	var out [][]int64
	err := json.Unmarshal([]byte(str), &out)
	if err != nil {
		panic(err)
	}
	return out
}

func (c *CustomParse) GetBossHallOpenTime() [][]int64 {
	return r.bossHallOpenTime

}

func (c *CustomParse) GitItemMap() map[values.Integer][]GiftItem {
	return r.CustomParse.gitItem
}

func (c *CustomParse) ExchangeItemWeight(itemId values.ItemId) (*CustomExchangeWeight, bool) {
	weight, ok := r.CustomParse.exchangeWeight[itemId]
	return weight, ok
}

func (c *CustomParse) CriticalRatio() float64 {
	return r.CustomParse.criticalRatio
}

func (c *CustomParse) AchievementGear() map[values.AchievementId]map[values.Integer]*AchievementList {
	return r.achievementCnt
}

func (c *CustomParse) GetShopMap() map[values.Integer][]GoodslistLv {
	return r.shopMap
}

func (i *Guild) MaxLevel() values.Level {
	return r.CustomParse.maxGuildLevel
}

func (i *GuildPosition) GuildPositionList() []values.GuildPosition {
	return r.CustomParse.guildPositionList
}

func (i *GuildPosition) GetLeader() values.GuildPosition {
	return r.CustomParse.guildPositionList[0]
}

func (i *CustomParse) RenameCost() [][2]values.Integer {
	return r.CustomParse.renameCost
}

func (i *CustomParse) DefaultTowerCost() [][2]values.Integer {
	return r.CustomParse.default_tower_cost
}

func (i *GuildPosition) GetMember() values.GuildPosition {
	return r.CustomParse.guildPositionList[len(r.CustomParse.guildPositionList)-1]
}

func (i *NpcTask) NpcTaskUnlockCond(typ models.TaskType) (map[values.Integer]map[values.Integer][]values.TaskId, bool) {
	cond, ok := r.CustomParse.npcTaskUnlock[typ]
	return cond, ok
}

func (i *MainTask) MainTaskChapter(id values.Integer) ([]MainTaskChapterReward, bool) {
	chapter, ok := r.CustomParse.mainTaskChapter[id]
	return chapter, ok
}

func (i *Relics) LevelUpCost(quality values.Integer, level values.Integer) (values.Integer, map[values.Integer]values.Integer, bool) {
	qualityCost, ok := r.CustomParse.relicsLevelUpCost[quality]
	if !ok {
		return 0, nil, false
	}
	cost := qualityCost[level]
	if len(cost.LvCost) == 0 {
		return 0, nil, false
	}
	return cost.UnlockLv, cost.LvCost, ok
}

func (i *Relics) Level(quality values.Integer) values.Integer {
	qualityCost, ok := r.CustomParse.relicsLevelUpCost[quality]
	if !ok {
		return 0
	}
	return values.Integer(len(qualityCost))
}

func (i *CustomParse) GetTaskTypAchieve() map[models.TaskType]map[values.AchievementId]values.Integer {
	return i.taskTypAchievement
}

func (i *Atlas) AtlasByType(typ values.Integer) (map[values.Integer]AtlasItem, bool) {
	a, ok := r.CustomParse.atlasByType[typ]
	return a, ok
}

func (i *Gacha) GachaUnlockCond(cond values.Integer) (map[values.Integer][]values.Integer, bool) {
	condition, ok := r.CustomParse.gachaUnlockCond[cond]
	return condition, ok
}

func (i *Gacha) GachaWeightById(gachaId values.GachaId) (*CustomGachaWeight, bool) {
	weight, ok := r.CustomParse.gachaWeightParse[gachaId]
	return weight, ok
}

func (i *CustomParse) GetRobotMap() map[values.Integer][]Robot {
	return i.robotMap
}

func (i *CustomParse) GetDropList(dropListID values.Integer) map[values.Integer]values.Integer {
	items := make(map[values.Integer]values.Integer, 0)

	list, ok := i.dropListMap[dropListID]
	if !ok {
		return items
	}

	dlms := make([]DropListsMini, 0)
	for _, v := range list {
		prob := rand.Int63n(10000) + 1
		if v.DropProb >= prob {
			dlms = append(dlms, v)
		}
	}

	for _, v := range dlms {
		drops := i.dropMap[v.DropId]
		choices := make([]*wr.Choice[DropMini, int64], 0, len(drops))
		for _, vv := range drops {
			choices = append(choices, wr.NewChoice(vv, vv.ItemWeight))
		}
		chooser, _ := wr.NewChooser(choices...)
		dm := chooser.Pick()
		items = utils.MergeMapNumber(items, dm.ItemId)
	}

	return items
}

func (i *CustomParse) GetShopTypMap(integer values.Integer) []values.Integer {
	if v, exist := i.shopTypMap[integer]; exist {
		return v
	}
	return nil
}

// func (i *RowHero) HeroLvUp() map[values.HeroId][]HeroLvup {
//	return r.CustomParse.heroLvUp
// }

func (i *RoleLv) MaxRoleLevel() values.Level {
	return r.CustomParse.maxRoleLevel
}

func (i *Divination) Data() CustomDivination {
	return r.divination
}

func (i *Anecdotes) CustomAnecdotes() map[values.StoryId][]AnecdotesClassText {
	return r.anecdotesPiece
}

func (i *EquipStar) MaxEquipStar() values.Level {
	return r.maxEquipStar
}

func (i *Anecdotes) CustomAnecdoteGame1() map[int64]map[int64]AnecdotesGame1Option {
	return r.anecdotesGame1
}

func (i *Anecdotes) CustomAnecdoteGame3() map[int64]map[int64]AnecdotesGame3Option {
	return r.anecdotesGame3
}

func (i *Anecdotes) CustomAnecdoteGame4() map[int64]map[int64]AnecdotesGame4Option {
	return r.anecdotesGame4
}

func (i *Anecdotes) CustomAnecdoteGame5() map[int64]map[int64]AnecdotesGame5Option {
	return r.anecdotesGame5
}

func (i *Anecdotes) CustomAnecdoteGame7() map[int64]map[int64]AnecdotesGame7Option {
	return r.anecdotesGame7
}

func (i *Anecdotes) CustomAnecdoteGame8() map[int64]map[int64]AnecdotesGame8Option {
	return r.anecdotesGame8
}

func (i *MapSelect) Next() *MapSelect {
	iter := r.mapSelect.LowerBound(i.Id)
	if !iter.Valid() {
		return nil
	}
	iter.Next()
	if !iter.Valid() {
		return nil
	}
	return iter.Value()
}

func (i *CustomParse) GetMapSelectIdBySceneId(sceneId int64) (int64, bool) {
	v, ok := i.mapSelectScene[sceneId]
	return v, ok
}

func (i *CustomParse) OriginTalentPoint() values.Integer {
	return i.originTalentPoint
}

func (i *Anecdotes) GetAutoRefreshEvent() []AnecdotesGames {
	return r.mapEventAutoRef
}

func (i *CustomParse) GetDefaultTalent() map[values.Integer][]Talent {
	return i.defaultTalent
}

func (i *CustomParse) GetEachTalent() map[values.Integer][]Talent {
	return i.eachTalent
}

func (i *CustomParse) GetOriginHead() [2]values.Integer {
	return i.originHead
}

func (i *AttrTrans) GetAttrTransByAttrId(attrId values.AttrId) []AttrTrans {
	return r.attrTransGroup[attrId]
}

func (i *NpcDialogue) GetNpcDialogOpts(dialogId values.Integer) []DialogueOpt {
	return r.npcDialogOpt[dialogId]
}

func (i *NpcDialogue) GetNpcDialogRelation(dialogId values.Integer) CustomDialogRelation {
	ret, ok := r.npcDialogRelation[dialogId]
	if !ok {
		ret.HeadDialogId = dialogId
		ret.IsEnd = true
	}
	return ret
}

func (i *Equip) GetEquipEntry(id values.ItemId) []EquipEntry {
	return r.equipEntryGroup[id]
}

func (i *EquipRefine) GetMaxEquipMeltLevel(slot values.EquipSlot) values.Level {
	return r.maxEquipMeltLevel[slot]
}

func (i *EquipRefine) GetEquipRefine(slot values.EquipSlot, level values.Level) (*EquipRefine, bool) {
	temp, ok := r.equipRefine[slot]
	if !ok {
		return nil, false
	}
	er, ok := temp[level]
	return er, ok
}

func (i *ForgeLevel) GetMaxForgeLevel() values.Level {
	return r.maxEquipForgeLevel
}

func (i *ForgeLevel) GetForgeLevelByEquipLevel(lv values.Level) []ForgeLevel {
	return r.equipForgeLeveGroup[lv]
}

func (i *ForgeLevel) GetEquipForgeMaxLevel() values.Integer {
	return r.equipForgeMaxLevel
}

func (i *RowHero) InitTalent(id values.HeroId) values.HeroId {
	return r.initRowHeroTalentMap[id]
}

func (i *CustomParse) OriginHero(id values.HeroId) values.Integer {
	return r.originHeroMap[id]
}

func (i *CustomParse) DeriveHeroMap(id values.HeroId) []values.HeroId {
	return r.deriveHeroMap[id]
}

func (i *CustomParse) GenRandomRobotNickname() string {
	l := r.RobotName.Len()
	fn, _ := r.RobotName.GetRobotNameById(values.Integer(rand.Intn(l)) + 1)
	sn, _ := r.RobotName.GetRobotNameById(values.Integer(rand.Intn(l)) + 1)
	return fmt.Sprintf("%s.%s", fn.FirstName, sn.SecondName)
}

func (i *CustomParse) GetDailySaleLine() ([]int64, []int64) {
	return r.dailySaleLevelUpdate, r.dailySaleChargeUpdate
}

func (i *CustomParse) GetWeeklySaleLine() ([]int64, []int64) {
	return r.weeklySaleLevelUpdate, r.weeklySaleChargeUpdate
}

func (i *CustomParse) GenRoguelikeBossGroup() map[values.Integer][]values.Integer {
	return i.roguelikeBossGroup
}

func (i *CustomParse) GenRoguelikeEntryGroup() map[values.Integer][]values.Integer {
	return i.roguelikeEntryGroup
}

func (i *Medicament) GetMedicineByType(typ values.Integer) ([]Medicament, bool) {
	s, ok := r.medicineSortedMap[typ]
	return s, ok
}

func (i *Medicament) GetMedicine() map[values.Integer][]Medicament {
	return r.medicineSortedMap
}

func (i *System) GetSysTypeByName(name string) (models.SystemType, bool) {
	typ, ok := r.sysName2Type[name]
	return typ, ok
}

func (i *LanguageBackend) GetContext(lag string) string {
	switch lag {
	case "en":
		return i.En
	case "cn":
		return i.Cn
	case "hk":
		return i.Hk
	default:
		return ""
	}
}

func (i *RoleLvTitle) GetMaxTitle() values.Integer {
	return r.maxTitle
}

func (i *Biography) GetBiographyByHero(heroId values.HeroId) []Biography {
	return r.biographyGroupByHero[heroId]
}

func (i *Biography) GetBiographyByTaskType(typ models.TaskType) []Biography {
	return r.biographyGroupByTaskType[typ]
}

func (i *ExpeditionQuantity) GetExpeditionQuantityByTaskType(typ models.TaskType) []ExpeditionQuantity {
	return r.expeditionQuantityGroupByTaskType[typ]
}

func (i *ExpeditionQuantity) GetAllExpeditionQuantityGroupByTaskType() map[models.TaskType][]ExpeditionQuantity {
	return r.expeditionQuantityGroupByTaskType
}

func (i *Expedition) GetAllExpedition() map[models.TaskType]map[values.Quality][]Expedition {
	return r.expeditionGroupByTaskType
}

func (i *ExpeditionReward) GetExpeditionRewardByExpeditionId(id values.Integer) []ExpeditionReward {
	return r.expeditionRewardGroupByParent[id]
}

func (i *SoulContract) GetSoulContractByHero(heroId values.HeroId) []SoulContract {
	return r.soulContractGroupByHero[heroId]
}

func (i *SoulContract) GetMaxSoulContractByHero(heroId values.HeroId) (MaxSoulContract, bool) {
	msc, ok := r.maxSoulContract[heroId]
	return msc, ok
}

func (i *ExpSkip) GetMaxExpSkipCount() values.Integer {
	return r.maxExpSkipCount
}

func (i *TargetTask) GetAllTargetTaskTypes() map[models.TaskType]struct{} {
	return r.allTargetTaskTypes
}

func (i *MainTask) GetAllMainTaskTypes() map[models.TaskType]struct{} {
	return r.allMainTaskTypes
}

func (i *NpcTask) GetAllNpcTaskTypes() map[models.TaskType]struct{} {
	return r.allNpcTaskTypes
}

func (i *Looptask) GetAllLoopTaskTypes() map[models.TaskType]struct{} {
	return r.allLoopTaskTypes
}

func (i *ActivityLimitedtimePackage) GetAllUnlockConditions() map[models.TaskType]map[values.Integer]struct{} {
	return r.allLimitedTimePackConditions
}

func (i *ActivityLimitedtimePackagePay) ListByParentId(id values.Integer) []ActivityLimitedtimePackagePay {
	return r.limitedTimePackagePayList[id]
}

func (i *Skill) GetMaxSkillId(id values.HeroSkillId) values.HeroSkillId {
	skillId, ok := r.maxSkillId[id]
	if !ok {
		skillId = id
	}
	return skillId
}

func (i *EquipResonance) GetEquipResonanceByHero(id values.HeroId) map[values.Level]EquipResonance {
	return r.equipResonanceGroupByHero[id]
}

func (i *Fashion) GetDefaultFashion(id values.HeroId) values.FashionId {
	// 服务启动的时候检查了，每个英雄必须配置默认皮肤，所以这里一定会取到值
	return r.defaultFashion[id]
}

func (i *MonsterAttr) GetPersonalBossMonsterAttr() map[values.Level]MonsterAttr {
	return r.personalBossMonsterAttr
}

func (i *PersonalBossLv) GetPersonalBossLvByFloor(floor int64) []PersonalBossLv {
	return r.personalBossLvGroupByFloor[floor]
}

func (i *EquipStar) GetAllCostItemId() []values.ItemId {
	return r.equipStarCost
}

func (i *CombatBattle) GetMaxRacingRanking() values.Integer {
	return r.maxRacingRanking
}

func (i *GuildBlessing) GetMaxGuildBlessingPage() values.Integer {
	return r.maxBlessingPage
}

// 自定义解析礼包数据
func customParseGiftItem() map[values.Integer][]GiftItem {
	dataMap := make(map[values.Integer][]GiftItem)
	for _, v := range h.giftItem {
		dataMap[v.GiftId] = append(dataMap[v.GiftId], v)
	}
	return dataMap
}

func customParseExchangeItem() map[values.ItemId][]ExchangeItem {
	dataMap := make(map[values.ItemId][]ExchangeItem)
	for _, v := range h.exchangeItem {
		dataMap[v.ExchangeId] = append(dataMap[v.ExchangeId], v)
	}
	return dataMap
}

func customParseExchangeWeight() map[values.ItemId]*CustomExchangeWeight {
	dataMap := make(map[values.ItemId][]ExchangeItem)
	for _, v := range h.exchangeItem {
		dataMap[v.ExchangeId] = append(dataMap[v.ExchangeId], v)
	}

	ret := map[values.ItemId]*CustomExchangeWeight{}
	for itemId, exchanges := range dataMap {
		var weight values.Integer
		exItem := make([]ExchangeItem, len(exchanges))
		for idx, v := range exchanges {
			weight += v.ItemWeight
			v.ItemWeight = weight
			exItem[idx] = v
		}
		ret[itemId] = &CustomExchangeWeight{
			TotalWeight:     weight,
			ExchangeWeights: exItem,
		}
	}
	return ret
}

func customBeginningHeroes() []int64 {
	out := make([]int64, 0, 2)
	for i := range h.begining {
		if h.begining[i].Typ == 2 {
			out = append(out, h.begining[i].Rewardid)
		}
	}
	return out
}

func customBeginningMap() map[int64]int64 {
	out := make(map[int64]int64, len(h.begining))
	for _, cfg := range h.begining {
		out[cfg.Rewardid] = cfg.Count
	}
	return out
}

func customDailySaleUpdateLine() ([]int64, []int64, []int64, []int64) {
	dailySaleLevelUpdate := make([]int64, 0, 8)
	dailySaleChargeUpdate := make([]int64, 0, 8)
	weeklySaleLevelUpdate := make([]int64, 0, 8)
	weeklySaleChargeUpdate := make([]int64, 0, 8)
	dailySaleLevelUpdateM := map[int64]bool{}
	dailySaleChargeUpdateM := map[int64]bool{}
	weeklySaleLevelUpdateM := map[int64]bool{}
	weeklySaleChargeUpdateM := map[int64]bool{}
	for _, v := range h.activityDailygiftBuy {
		if v.TypId == enum.DailySale {
			if _, exist := dailySaleLevelUpdateM[v.GradeRange[0]]; !exist {
				dailySaleLevelUpdateM[v.GradeRange[0]] = true
				dailySaleLevelUpdate = append(dailySaleLevelUpdate, v.GradeRange[0])
			}
			if _, exist := dailySaleChargeUpdateM[v.RechargeInterval[0]]; !exist {
				dailySaleChargeUpdateM[v.RechargeInterval[0]] = true
				dailySaleChargeUpdate = append(dailySaleChargeUpdate, v.RechargeInterval[0])
			}
		} else if v.TypId == enum.WeeklySale {
			if _, exist := weeklySaleLevelUpdateM[v.GradeRange[0]]; !exist {
				weeklySaleLevelUpdateM[v.GradeRange[0]] = true
				weeklySaleLevelUpdate = append(weeklySaleLevelUpdate, v.GradeRange[0])
			}
			if _, exist := weeklySaleChargeUpdateM[v.RechargeInterval[0]]; !exist {
				weeklySaleChargeUpdateM[v.RechargeInterval[0]] = true
				weeklySaleChargeUpdate = append(weeklySaleChargeUpdate, v.RechargeInterval[0])
			}
		}
	}
	return dailySaleLevelUpdate, dailySaleChargeUpdate, weeklySaleLevelUpdate, weeklySaleChargeUpdate
}

func customParseMap(value string) [][2]values.Integer {
	data := strings.Split(value, ",")
	res := make([][2]values.Integer, 0, len(data))
	for _, itemKV := range data {
		itemList := strings.Split(itemKV, ":")
		if len(itemList) != 2 {
			continue
		}
		kInt, err := strconv.ParseInt(itemList[0], 10, 64)
		if err != nil {
			continue
		}
		vInt, err := strconv.ParseInt(itemList[1], 10, 64)
		if err != nil {
			continue
		}
		res = append(res, [2]values.Integer{kInt, vInt})
	}
	return res
}

func customParseRenameCost() [][2]values.Integer {
	for _, el := range h.keyValue {
		if el.Key == "RenameConsumeItem" {
			return customParseMap(el.Value.(string))
		}
	}
	return make([][2]values.Integer, 0)
}

func customParaesDefaultTowerCost() [][2]values.Integer {
	for _, el := range h.keyValue {
		if el.Key == "DefaultTowerCost" {
			return customParseMap(el.Value.(string))
		}
	}
	return make([][2]values.Integer, 0)
}

func customParseOriginHead() (res [2]values.Integer, talentPoint values.Integer) {
	for _, el := range h.begining {
		if el.Typ == 3 {
			for _, hs := range h.headSculpture {
				if hs.Id == el.Rewardid {
					if hs.HeadSculptureType == 1 {
						res[0] = hs.Id
					}
					if hs.HeadSculptureType == 2 {
						res[1] = hs.Id
					}
				}
			}
		}
		if el.Typ == 4 {
			talentPoint = el.Count
		}
	}
	return
}

func customParseDefaultTalent() (map[values.Integer][]Talent, map[values.Integer][]Talent) {
	res := map[values.Integer][]Talent{}
	all := map[values.Integer][]Talent{}
	for _, el := range h.talent {
		if _, exist := res[el.BuildId]; !exist {
			res[el.BuildId] = make([]Talent, 0, 4)
		}
		if _, exist := all[el.BuildId]; !exist {
			all[el.BuildId] = make([]Talent, 0, 32)
		}
		if el.DefaultActivation == 1 {
			res[el.BuildId] = append(res[el.BuildId], el)
		}
		all[el.BuildId] = append(all[el.BuildId], el)
	}
	return res, all
}

func customParseGlobalNeeded() float64 {
	for _, el := range h.keyValue {
		if el.Key == "CriticalRatio" {
			temp, ok := el.Value.(string)
			if !ok {
				return 1
			}
			critical, err := strconv.Atoi(temp)
			if err != nil {
				return 1
			}
			return float64(critical) / 10000
		}
	}
	return 1
}

func customParseGuildLevel() values.Level {
	var maxLevel values.Level
	for _, el := range h.guild {
		if el.Id > maxLevel {
			maxLevel = el.Id
		}
	}
	return maxLevel
}

func customParseRobot() map[values.Integer][]Robot {
	robotMap := map[values.Integer][]Robot{}
	for _, el := range h.robot {
		if _, exist := robotMap[el.ConfigId]; !exist {
			robotMap[el.ConfigId] = make([]Robot, 0, 8)
		}
		robotMap[el.ConfigId] = append(robotMap[el.ConfigId], el)
	}
	return robotMap
}

func customParseRoguelikeBossGroup() map[values.Integer][]values.Integer {
	res := map[values.Integer][]values.Integer{}
	for _, el := range h.roguelikeDungeon {
		if _, exist := res[el.DungeonDay]; !exist {
			res[el.DungeonDay] = make([]values.Integer, 0, 8)
		}
		res[el.DungeonDay] = append(res[el.DungeonDay], el.Id)
	}
	return res
}

func customParseRoguelikeEntryGroup() map[values.Integer][]values.Integer {
	res := map[values.Integer][]values.Integer{}
	for _, el := range h.roguelikeEntry {
		if _, exist := res[el.Group]; !exist {
			res[el.Group] = make([]values.Integer, 0, 8)
		}
		res[el.Group] = append(res[el.Group], el.Id)
	}
	return res
}

func customParseAchievementMap() map[values.AchievementId]map[values.Integer]*AchievementList {
	res := map[values.AchievementId]map[values.Integer]*AchievementList{}
	for i, el := range h.achievementList {
		if _, exist := res[el.AchievementId]; !exist {
			res[el.AchievementId] = map[values.Integer]*AchievementList{}
		}
		res[el.AchievementId][el.Id] = &h.achievementList[i]
	}
	return res
}

func customParseTaskAchievement() map[models.TaskType]map[values.AchievementId]values.Integer {
	res := map[models.TaskType]map[values.AchievementId]values.Integer{}
	for _, el := range h.achievementList {
		ttp := el.TaskTypeParam
		if len(ttp) >= 3 {
			typ := models.TaskType(ttp[0])
			if _, exist := res[typ]; !exist {
				res[typ] = map[values.AchievementId]values.Integer{}
			}
			if _, exist := res[typ][el.AchievementId]; !exist {
				res[typ][el.AchievementId] = ttp[1]
			}
		}
	}
	return res
}

func customParseGuildPosition() []values.GuildPosition {
	var res []values.GuildPosition
	for _, el := range h.guildPosition {
		res = append(res, el.Id)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

func customParseRelicsCost() map[values.Integer][]RelicsLvQuality {
	dataMap := map[values.Integer][]RelicsLvQuality{}
	for _, v := range h.relicsLvQuality {
		dataMap[v.RelicsLvId] = append(dataMap[v.RelicsLvId], v)
	}
	return dataMap
}

func customParseAutoRefreshGame() []AnecdotesGames {
	res := make([]AnecdotesGames, 0)
	for _, v := range h.anecdotesGames {
		if v.Typ == 1 {
			res = append(res, v)
		}
	}
	return res
}

func customParseGachaUnlock() map[values.Integer]map[values.Integer][]values.Integer {
	dataMap := map[values.Integer]map[values.Integer][]values.Integer{}
	for _, v := range h.gacha {
		for cond, count := range v.Unlock {
			if dataMap[cond] == nil {
				dataMap[cond] = map[values.Integer][]values.Integer{}
			}
			if dataMap[cond][count] == nil {
				dataMap[cond][count] = make([]values.Integer, 0)
			}
			dataMap[cond][count] = append(dataMap[cond][count], v.Id)
		}
	}
	return dataMap
}

func customParseNpcTaskCondition() map[models.TaskType]map[values.Integer]map[values.Integer][]values.TaskId {
	dataMap := map[models.TaskType]map[values.Integer]map[values.Integer][]values.TaskId{}
	for _, v := range h.npcTask {
		for _, cond := range v.AcceptTaskParam {
			if len(cond) == 0 {
				continue
			}
			if dataMap[models.TaskType(cond[0])] == nil {
				dataMap[models.TaskType(cond[0])] = map[values.Integer]map[values.Integer][]values.TaskId{}
			}
			if dataMap[models.TaskType(cond[0])][cond[1]] == nil {
				dataMap[models.TaskType(cond[0])][cond[1]] = map[values.Integer][]values.TaskId{}
			}
			if dataMap[models.TaskType(cond[0])][cond[1]][cond[2]] == nil {
				dataMap[models.TaskType(cond[0])][cond[1]][cond[2]] = make([]values.TaskId, 0)
			}
			dataMap[models.TaskType(cond[0])][cond[1]][cond[2]] = append(dataMap[models.TaskType(cond[0])][cond[1]][cond[2]], v.Id)
		}
	}
	return dataMap
}

func customParseMainTaskChapter() map[values.Integer][]MainTaskChapterReward {
	dataMap := map[values.Integer][]MainTaskChapterReward{}
	for _, v := range h.mainTaskChapterReward {
		dataMap[v.MainTaskChapterId] = append(dataMap[v.MainTaskChapterId], v)
	}
	return dataMap
}

func customParseAtlasType() map[values.Integer]map[values.Integer]AtlasItem {
	dataMap := map[values.Integer]map[values.Integer]AtlasItem{}
	for _, v := range h.atlasItem {
		if dataMap[v.AtlasId] == nil {
			dataMap[v.AtlasId] = map[values.Integer]AtlasItem{}
		}
		dataMap[v.AtlasId][v.Id] = v
	}
	return dataMap
}

func customParseShopMap() map[values.Integer][]GoodslistLv {
	dataMap := map[values.Integer][]GoodslistLv{}
	for _, v := range h.goodslistLv {
		dataMap[v.ShopgoodslistId] = append(dataMap[v.ShopgoodslistId], v)
	}
	return dataMap
}

func customParseShopTyp() map[values.Integer][]values.Integer {
	dataMap := map[values.Integer][]values.Integer{}
	for _, v := range h.shopgoodslist {
		if v.Id > 0 && v.Id <= 2000 {
			dataMap[0] = append(dataMap[0], v.Id)
		} else if v.Id > 2000 && v.Id <= 3000 {
			dataMap[2] = append(dataMap[2], v.Id)
		} else if v.Id > 3000 && v.Id <= 4000 {
			dataMap[1] = append(dataMap[1], v.Id)
		} else if v.Id > 4000 && v.Id <= 5000 {
			dataMap[3] = append(dataMap[3], v.Id)
		}
	}
	return dataMap
}

func customParseDropList() map[values.Integer][]DropListsMini {
	dataMap := map[values.Integer][]DropListsMini{}
	for _, v := range h.dropListsMini {
		dataMap[v.DropListsId] = append(dataMap[v.DropListsId], v)
	}
	return dataMap
}

func customParseDrop() map[values.Integer][]DropMini {
	dataMap := map[values.Integer][]DropMini{}
	for _, v := range h.dropMini {
		dataMap[v.DropId] = append(dataMap[v.DropId], v)
	}
	return dataMap
}

// func customParseHeroLvUp() map[values.HeroId][]HeroLvup {
//	data := make(map[values.HeroId][]HeroLvup)
//	for _, el := range h.heroLvup {
//		_, ok := data[el.RowHeroId]
//		if !ok {
//			data[el.RowHeroId] = make([]HeroLvup, 0)
//		}
//		data[el.RowHeroId] = append(data[el.RowHeroId], el)
//	}
//	return data
// }

func customParseMaxRoleLevel() values.Level {
	var lv values.Level
	for _, el := range h.roleLv {
		if el.Id > lv {
			lv = el.Id
		}
	}
	return lv
}

func customParseDivination() CustomDivination {
	ret := CustomDivination{}
	for _, el := range h.divination {
		ret.TotalWeight += el.ItemWeight
		ret.List = append(ret.List, el)
	}
	return ret
}

func customParseGachaWeight() map[values.GachaId]*CustomGachaWeight {
	ret := map[values.GachaId]*CustomGachaWeight{}
	for _, ga := range h.gacha {
		gachaIdx := make([]values.Integer, len(ga.GachaWeight))
		gachaWeight := make([]values.Integer, len(ga.GachaWeight))
		var weight values.Integer
		n := 0
		for k, v := range ga.GachaWeight {
			weight += v
			gachaIdx[n] = k
			gachaWeight[n] = weight
			n++
		}
		ret[ga.Id] = &CustomGachaWeight{
			TotalWeight:  weight,
			GachaIdx:     gachaIdx,
			GachaWeights: gachaWeight,
		}
	}
	return ret
}

func customParseAnecdotes() map[values.StoryId][]AnecdotesClassText {
	dataMap := make(map[values.StoryId][]AnecdotesClassText)
	for _, v := range h.anecdotesClassText {
		dataMap[v.AnecdotesId] = append(dataMap[v.AnecdotesId], v)
	}
	return dataMap
}

func customParseAnecdoteGame1() map[int64]map[int64]AnecdotesGame1Option {
	dataMap := map[int64]map[int64]AnecdotesGame1Option{}
	for _, v := range h.anecdotesGame1Option {
		if _, ok := dataMap[v.AnecdotesGame1Id]; !ok {
			dataMap[v.AnecdotesGame1Id] = map[int64]AnecdotesGame1Option{}
		}
		dataMap[v.AnecdotesGame1Id][v.Id] = v
	}
	return dataMap
}

func customParseAnecdoteGame3() map[int64]map[int64]AnecdotesGame3Option {
	dataMap := map[int64]map[int64]AnecdotesGame3Option{}
	for _, v := range h.anecdotesGame3Option {
		if _, ok := dataMap[v.AnecdotesGame3Id]; !ok {
			dataMap[v.AnecdotesGame3Id] = map[int64]AnecdotesGame3Option{}
		}
		dataMap[v.AnecdotesGame3Id][v.Id] = v
	}
	return dataMap
}

func customParseAnecdoteGame4() map[int64]map[int64]AnecdotesGame4Option {
	dataMap := map[int64]map[int64]AnecdotesGame4Option{}
	for _, v := range h.anecdotesGame4Option {
		if _, ok := dataMap[v.AnecdotesGame4Id]; !ok {
			dataMap[v.AnecdotesGame4Id] = map[int64]AnecdotesGame4Option{}
		}
		dataMap[v.AnecdotesGame4Id][v.Id] = v
	}
	return dataMap
}

func customParseAnecdoteGame5() map[int64]map[int64]AnecdotesGame5Option {
	dataMap := map[int64]map[int64]AnecdotesGame5Option{}
	for _, v := range h.anecdotesGame5Option {
		if _, ok := dataMap[v.AnecdotesGame5Id]; !ok {
			dataMap[v.AnecdotesGame5Id] = map[int64]AnecdotesGame5Option{}
		}
		dataMap[v.AnecdotesGame5Id][v.Id] = v
	}
	return dataMap
}

func customParseAnecdoteGame7() map[int64]map[int64]AnecdotesGame7Option {
	dataMap := map[int64]map[int64]AnecdotesGame7Option{}
	for _, v := range h.anecdotesGame7Option {
		if _, ok := dataMap[v.AnecdotesGame7Id]; !ok {
			dataMap[v.AnecdotesGame7Id] = map[int64]AnecdotesGame7Option{}
		}
		dataMap[v.AnecdotesGame7Id][v.Id] = v
	}
	return dataMap
}

func customParseAnecdoteGame8() map[int64]map[int64]AnecdotesGame8Option {
	dataMap := map[int64]map[int64]AnecdotesGame8Option{}
	for _, v := range h.anecdotesGame8Option {
		if _, ok := dataMap[v.AnecdotesGame8Id]; !ok {
			dataMap[v.AnecdotesGame8Id] = map[int64]AnecdotesGame8Option{}
		}
		dataMap[v.AnecdotesGame8Id][v.Id] = v
	}
	return dataMap
}

func parseCounter() {
	parser := condition.GetParser()
	for _, v := range h.achievementList {
		if len(v.TaskTypeParam) == 2 {
			parser.ParseCounter(v.TaskTypeParam[0], v.TaskTypeParam[1])
		}
	}
	return
}

func customParseMaxEquipStar() values.Level {
	var lv values.Level
	for _, el := range h.equipStar {
		if el.Id > lv {
			lv = el.Id
		}
	}
	return lv
}

func customParseAttrTrans() map[values.AttrId][]AttrTrans {
	dataMap := make(map[values.AttrId][]AttrTrans)
	for _, el := range h.attrTrans {
		if _, ok := dataMap[el.AttrId]; !ok {
			dataMap[el.AttrId] = make([]AttrTrans, 0)
		}
		dataMap[el.AttrId] = append(dataMap[el.AttrId], el)
	}
	return dataMap
}

func equipConfigCheck() {
	for _, el := range h.equip {
		attrIdCount := len(el.AttrId)
		if len(el.AttrValue) != attrIdCount {
			panic(fmt.Errorf("equip 表配置错误: AttrValue数据和AttrID数据个数不匹配,id=%d", el.Id))
		}
		if len(el.AttrStarRange) != attrIdCount {
			panic(fmt.Errorf("equip 表配置错误: AttrStarRange数据和AttrID数据个数不匹配,id=%d", el.Id))
		}
		if len(el.Attr) != attrIdCount {
			panic(fmt.Errorf("equip 表配置错误: Attr数据和AttrID数据个数不匹配,id=%d", el.Id))
		}
		for i, asr := range el.AttrStarRange {
			count := len(asr)
			if len(el.Attr[i]) != count {
				panic(fmt.Errorf("equip 表配置错误: AttrStarRange数据和Attr数据个数不匹配,id=%d", el.Id))
			}
		}
		if entryGroup, ok := r.equipEntryGroup[el.Attribute]; !ok {
			panic(fmt.Errorf("equip 表配置错误: Attribute在equip_entry表里未找到对应配置,id=%d", el.Id))
		} else {
			// 检查必出品质的词条在equip_entry表对应的qualitysection列是否存在
			var find bool
			for _, entry := range entryGroup {
				if _, ok := entry.Qualityweight[el.AttributeQuality]; ok {
					find = true
					break
				}
			}
			if !find {
				// 当未找到时，只要词条组里存在>=el.AttributeQuality的品质也算ok
				for _, entry := range entryGroup {
					for q := range entry.Qualityweight {
						if q >= el.AttributeQuality {
							find = true
						}
					}
				}
				if !find {
					panic(fmt.Errorf("equip 表配置错误: AttributeQuality在equip_entry表里对应的qualityweight列未找到对应配置,id=%d", el.Id))
				}
			}
		}
		for q := range el.QualityNum {
			if _, ok := el.QualityEffects[q]; !ok {
				panic(fmt.Errorf("equip 表配置错误: QualityNum在QualityEffects列未找到对应配置,id=%d, q=%d", el.Id, q))
			}
		}
	}
}

func customParseDialog() map[values.Integer][]DialogueOpt {
	dataMap := map[values.Integer][]DialogueOpt{}
	for _, v := range h.dialogueWord {
		for _, vv := range r.DialogueOpt.List() {
			if v.Id == vv.DialogueWordId && vv.Id > 0 {
				dataMap[v.NpcDialogueId] = append(dataMap[v.NpcDialogueId], vv)
			}
		}
	}
	return dataMap
}

func customParseEquipEntry() map[values.Integer][]EquipEntry {
	dataMap := make(map[values.Integer][]EquipEntry)
	for _, entry := range h.equipEntry {
		if entry.Attr > 0 && len(entry.Qualitysection) <= 0 && len(entry.QualityPer) <= 0 {
			panic(fmt.Errorf("equip_entry 表配置错误: Qualitysection和QualityPer均为空,id=%d", entry.Id))
		}
		_, ok := dataMap[entry.Group]
		if !ok {
			dataMap[entry.Group] = make([]EquipEntry, 0)
		}
		dataMap[entry.Group] = append(dataMap[entry.Group], entry)
		if entry.Lowvalue > 0 && entry.Maxvalue > 0 && entry.Maxvalue-entry.Lowvalue < 0 {
			panic(fmt.Errorf("equip_entry 表配置错误: maxvalue和lowvalue不能<=0,id=%d", entry.Id))
		}
		// 检查qualityweight配置项在qualitysection列是否都存在
		for id := range entry.Qualityweight {
			target := entry.Qualitysection
			if entry.Lowvalue == 0 && entry.Maxvalue == 0 {
				target = entry.Skillquality
			}
			_, ok1 := target[id]
			_, ok2 := entry.QualityPer[id]
			_, ok3 := entry.SkillLv[id]
			if !ok1 && !ok2 && !ok3 {
				panic(fmt.Errorf("equip_entry 表配置错误: qualityweight配置项在qualitysection、qualityPer、skillquality、SkillLv均未找到,id=%d", entry.Id))
			}
		}
	}
	return dataMap
}

func customParseMaxEquipRefine() (map[values.EquipSlot]values.Level, map[values.EquipSlot]map[values.Level]*EquipRefine) {
	max := make(map[values.EquipSlot]values.Level)
	dataMap := make(map[values.EquipSlot]map[values.Level]*EquipRefine)
	for i := 0; i < len(h.equipRefine); i++ {
		el := &h.equipRefine[i]
		lv := max[el.EquipSlot]
		if el.RefineLv > lv {
			max[el.EquipSlot] = el.RefineLv
		}
		if _, ok := dataMap[el.EquipSlot]; !ok {
			dataMap[el.EquipSlot] = make(map[values.Level]*EquipRefine)
		}
		dataMap[el.EquipSlot][el.RefineLv] = el
	}
	return max, dataMap
}

func guildCfgCheck() {
	for _, el := range h.guild {
		if len(el.SurprisedMag) != 2 {
			panic(fmt.Errorf("guild表SurprisedMag列配置有误，需包含2个元素，id=%d", el.Id))
		}
		if len(el.BuildFree) != 2 {
			panic(fmt.Errorf("guild表BuildFree列配置有误，需包含2个元素，id=%d", el.Id))
		}
	}
}

func customParseMaxEquipForgeLevel() values.Level {
	var lv values.Level
	for _, el := range h.forgeLevel {
		if el.Id > lv {
			lv = el.Id
		}
	}
	return lv
}

func foreFixedRecipeCheck() {
	for _, recipe := range h.forgeFixedRecipe {
		quality := make(map[values.Quality]struct{})
		for _, el := range recipe.Products {
			if len(el) != 3 {
				panic(fmt.Errorf("forge_fixed_recipe 表 Products 列配置有误，每个子元素需包含3个元素，id=%d", recipe.Id))
			}
			if _, ok := quality[el[1]]; ok {
				panic(fmt.Errorf("forge_fixed_recipe 表 Products 列配置有误，重复的品质配置，id=%d", recipe.Id))
			}
			quality[el[1]] = struct{}{}
		}
	}
}

func customParseEquipForgeLevelGroup() (map[values.Level][]ForgeLevel, values.Integer) {
	dataMap := make(map[values.Level][]ForgeLevel)
	var maxEquipLevel values.Integer
	for _, level := range h.forgeLevel {
		if _, ok := dataMap[level.EquipLevel]; !ok {
			dataMap[level.EquipLevel] = make([]ForgeLevel, 0)
		}
		dataMap[level.EquipLevel] = append(dataMap[level.EquipLevel], level)
		if level.EquipLevel > maxEquipLevel {
			maxEquipLevel = level.EquipLevel
		}
	}
	for level, el := range dataMap {
		sort.Slice(el, func(i, j int) bool {
			return el[i].Id < el[j].Id
		})
		dataMap[level] = el
	}
	return dataMap, maxEquipLevel
}

func relicsCfgCheck() {
	for _, rel := range h.relics {
		if !(len(rel.AttrId) == len(rel.AttrValue) && len(rel.AttrId) == len(rel.AttrPer) &&
			len(rel.AttrId) == len(rel.AttrStarRange) && len(rel.AttrId) == len(rel.Attr)) {
			panic(fmt.Sprintf("relics表配置错误，AttrID、AttrValue、AttrPer、AttrStarRange、Attr长度不匹配"))
		}

		for idx, attrRange := range rel.AttrStarRange {
			if len(attrRange) != len(rel.Attr[idx]) {
				panic(fmt.Sprintf("relics表配置错误，AttrStarRange 长度与 Attr不匹配"))
			}
			for i := 0; i < len(attrRange)-1; i++ {
				if attrRange[i] > attrRange[i+1] {
					panic(fmt.Sprintf("relics表配置错误，AttrStarRange 不为递增数组"))
				}
			}
		}

		for _, cost := range rel.StarsCost {
			if len(cost)%2 != 0 {
				panic(fmt.Sprintf("relics表配置错误，StarsCost 长度不是偶数"))
			}
		}
	}

	for _, qual := range r.CustomParse.relicsLevelUpCost {
		if qual[0].Id != 1 {
			panic(fmt.Sprintf("relics_lv_quality表配置错误，Id 不是从1开始"))
		}
		for i := 0; i < len(qual)-1; i++ {
			if qual[i].Id+1 != qual[i+1].Id {
				panic(fmt.Sprintf("relics_lv_quality表配置错误，Id 不是每级递增+1"))
			}
		}
	}
}

func exchangeCfgCheck() {
	for _, ex := range h.exchange {
		_, ok := r.Item.GetItemById(ex.Id)
		if !ok {
			panic(fmt.Sprintf("exchange表配置错误，id %d 在item表中不存在", ex.Id))
		}
		if len(ex.Typ) != 1 && len(ex.Typ) != 2 {
			panic(fmt.Sprintf("exchange表配置错误，typ字段长度必须为1或2,id=%d", ex.Id))
		}
		switch ex.Typ[0] {
		case 0:
		case 1:
			if len(ex.Typ) != 2 {
				panic(fmt.Sprintf("exchange表配置错误，typ字段为1时，必须配置交换个数,id=%d", ex.Id))
			}
		case 2:
		case 3:
			if len(ex.Typ) != 2 {
				panic(fmt.Sprintf("exchange表配置错误，typ字段为3时，必须配置交换个数"))
			}
		default:
			panic(fmt.Sprintf("exchange表配置错误，未支持的typ,id=%d", ex.Id))
		}
	}
}

func customParseMapSelect() (*treemap.TreeMap[int64, *MapSelect], map[int64]int64) {
	tr := treemap.New[int64, *MapSelect]()
	sceneKeyMap := make(map[int64]int64, len(h.mapSelect))
	for k, v := range h.mapSelect {
		tr.Set(v.Id, &h.mapSelect[k])
		sceneKeyMap[v.Scene] = v.Id
	}
	return tr, sceneKeyMap
}

func equipCompleteCheck() {
	for _, complete := range h.equipComplete {
		// TODO 先不判断，可能会没有套装效果的情况
		// if len(complete.SkillTwo) != 3 {
		// 	panic(fmt.Sprintf("equip_complete表 SkillTwo 配置错误，必须包含3个元素，id=%d", complete.Id))
		// }
		// if len(complete.SkillFour) != 3 {
		// 	panic(fmt.Sprintf("equip_complete表 SkillFour 配置错误，必须包含3个元素，id=%d", complete.Id))
		// }
		// if len(complete.SkillSix) != 3 {
		// 	panic(fmt.Sprintf("equip_complete表 SkillSix 配置错误，必须包含3个元素，id=%d", complete.Id))
		// }

		if len(complete.SkillTwo) == 3 {
			l0 := len(complete.SkillTwo[0])
			l1 := len(complete.SkillTwo[1])
			l2 := len(complete.SkillTwo[2])
			if l0 != l1 || l0 != l2 || l1 != l2 {
				panic(fmt.Sprintf("equip_complete表 SkillTwo 配置错误，子元素数量不等，id=%d", complete.Id))
			}
		}
		if len(complete.SkillFour) == 3 {
			l0 := len(complete.SkillFour[0])
			l1 := len(complete.SkillFour[1])
			l2 := len(complete.SkillFour[2])
			if l0 != l1 || l0 != l2 || l1 != l2 {
				panic(fmt.Sprintf("equip_complete表 SkillFour 配置错误，子元素数量不等，id=%d", complete.Id))
			}
		}
		if len(complete.SkillSix) == 3 {
			l0 := len(complete.SkillSix[0])
			l1 := len(complete.SkillSix[1])
			l2 := len(complete.SkillSix[2])
			if l0 != l1 || l0 != l2 || l1 != l2 {
				panic(fmt.Sprintf("equip_complete表 SkillSix 配置错误，子元素数量不等，id=%d", complete.Id))
			}
		}
	}
}

func customParseNpcDialogRelation() CustomNpcDialogRelationMap {
	review := make(CustomNpcDialogRelationMap)
	for _, v := range h.npcDialogue {
		opts := v.GetNpcDialogOpts(v.Id)
		if len(opts) == 0 {
			continue
		}
		for _, opt := range opts {
			if opt.OptNextDialogue == 0 {
				continue
			}
			if _, ok := review[v.Id]; ok {
				old := review[v.Id]
				old.IsEnd = false
				review[v.Id] = old
				review[opt.OptNextDialogue] = CustomDialogRelation{HeadDialogId: old.HeadDialogId, IsEnd: true}
			} else {
				review[v.Id] = CustomDialogRelation{HeadDialogId: v.Id, IsEnd: false}
				review[opt.OptNextDialogue] = CustomDialogRelation{HeadDialogId: v.Id, IsEnd: true}
			}
		}
	}
	return review
}

func customParseMedicineMap() map[values.Integer][]Medicament {
	dataMap := map[values.Integer][]Medicament{}
	for _, v := range h.medicament {
		if dataMap[v.Typ] == nil {
			dataMap[v.Typ] = make([]Medicament, 0)
		}
		dataMap[v.Typ] = append(dataMap[v.Typ], v)
	}
	for _, s := range dataMap {
		sort.Slice(s, func(i, j int) bool {
			return s[i].Level > s[j].Level
		})
	}
	return dataMap
}

func rowHeroCheck() {
	for _, el := range h.rowHero {
		if el.OriginId <= 0 {
			panic(fmt.Sprintf("row_heor表 OriginID 配置错误，OriginID<=0，id=%d", el.Id))
		}
	}
}

func combatBattleCheck() {
	for _, el := range h.combatBattle {
		if len(el.SubsectionRank) != 2 {
			panic(fmt.Sprintf("combat_battle 表 SubsectionRank 列配置错误，必须包含2个元素，id=%d", el.Id))
		}
		if len(el.RankReward) <= 0 {
			panic(fmt.Sprintf("combat_battle 表 RankReward 列配置错误，奖励为空，id=%d", el.Id))
		}
	}
}

func expeditionCostRecoveryCheck() {
	for _, el := range h.keyValue {
		if el.Key == "ExpeditionCostRecovery" {
			value, ok := el.Value.([]values.Integer)
			if !ok {
				panic(fmt.Sprintf("keyvalue表 %s 配置错误", el.Key))
			}
			if len(value) != 2 {
				panic(fmt.Sprintf("keyvalue表 %s 配置错误，必须为2个元素", el.Key))
			}
			break
		}
	}
}

func customParseInitRowHeroTalent() (map[values.HeroId]values.Integer, map[values.HeroId]values.HeroId, map[values.HeroId][]values.Integer) {
	temp := make(map[values.HeroId][]values.Integer)
	originMap := make(map[values.HeroId]values.Integer)
	for _, el := range h.rowHero {
		originMap[el.Id] = el.OriginId
		if _, ok := temp[el.OriginId]; !ok {
			temp[el.OriginId] = make([]values.Integer, 0)
		}
		temp[el.OriginId] = append(temp[el.OriginId], el.Id)
	}
	ret := make(map[values.HeroId]values.HeroId)
	for _, el := range temp {
		sort.Slice(el, func(i, j int) bool {
			return el[i] < el[j]
		})
		var initId values.HeroId
		for i, id := range el {
			if i == 0 {
				initId = id
			}
			ret[id] = initId
		}
	}
	return originMap, ret, temp
}

func customParseSysType() map[string]models.SystemType {
	res := map[string]models.SystemType{}
	for _, s := range h.system {
		for _, module := range s.ModuleName {
			res[module] = models.SystemType(s.Id)
		}
	}
	return res
}

func customParseMaxTitle() values.Integer {
	var title values.Integer
	for _, el := range h.roleLvTitle {
		if el.Id > title {
			title = el.Id
		}
	}
	return title
}

func customParseBiography() (map[values.HeroId][]Biography, map[models.TaskType][]Biography) {
	groupByHero := make(map[values.HeroId][]Biography)
	groupByTaskType := make(map[models.TaskType][]Biography)
	for _, el := range h.biography {
		if _, ok := groupByHero[el.HeroId]; !ok {
			groupByHero[el.HeroId] = make([]Biography, 0)
		}
		groupByHero[el.HeroId] = append(groupByHero[el.HeroId], el)

		var taskType models.TaskType
		unlockCondition := el.UnlockCondition
		if len(unlockCondition) == 3 {
			taskType = models.TaskType(unlockCondition[0])
		}
		if len(el.Reward) > 0 {
			if _, ok := groupByTaskType[taskType]; !ok {
				groupByTaskType[taskType] = make([]Biography, 0)
			}
			groupByTaskType[taskType] = append(groupByTaskType[taskType], el)
		}
	}
	return groupByHero, groupByTaskType
}

func customParseExpeditionQuantity() map[models.TaskType][]ExpeditionQuantity {
	groupByTaskType := make(map[models.TaskType][]ExpeditionQuantity)
	for _, el := range h.expeditionQuantity {
		var taskType models.TaskType
		unlockCondition := el.UnlockCondtion
		if len(unlockCondition) == 3 {
			taskType = models.TaskType(unlockCondition[0])
		}
		// taskType=0表示不需要解锁条件
		if _, ok := groupByTaskType[taskType]; !ok {
			groupByTaskType[taskType] = make([]ExpeditionQuantity, 0)
		}
		groupByTaskType[taskType] = append(groupByTaskType[taskType], el)
	}
	return groupByTaskType
}

func customParseExpedition() map[models.TaskType]map[values.Quality][]Expedition {
	group := make(map[models.TaskType]map[values.Quality][]Expedition)
	for _, el := range h.expedition {
		var taskType models.TaskType
		acceptCondition := el.AcceptCondtion
		if len(acceptCondition) == 3 {
			taskType = models.TaskType(acceptCondition[0])
		}
		// taskType=0表示不需要可接条件
		if _, ok := group[taskType]; !ok {
			group[taskType] = map[values.Quality][]Expedition{}
		}
		if _, ok := group[taskType][el.Quality]; !ok {
			group[taskType][el.Quality] = make([]Expedition, 0)
		}

		group[taskType][el.Quality] = append(group[taskType][el.Quality], el)
	}
	return group
}

func customParseSoulContract() (map[values.HeroId][]SoulContract, map[values.HeroId]MaxSoulContract) {
	group := make(map[values.HeroId][]SoulContract)
	max := make(map[values.HeroId]MaxSoulContract)
	for _, contract := range h.soulContract {
		if _, ok := group[contract.RoleHero]; !ok {
			group[contract.RoleHero] = make([]SoulContract, 0)
		}
		group[contract.RoleHero] = append(group[contract.RoleHero], contract)

		temp, ok := max[contract.RoleHero]
		if !ok {
			max[contract.RoleHero] = MaxSoulContract{
				Rank:  contract.Rank,
				Level: contract.Level,
			}
		} else {
			var update bool
			if contract.Rank > temp.Rank {
				temp.Rank = contract.Rank
				update = true
			}
			if contract.Level > temp.Level {
				temp.Level = contract.Level
				update = true
			}
			if update {
				max[contract.RoleHero] = temp
			}
		}
	}
	return group, max
}

func customParseExpSkip() values.Integer {
	data, ok := r.KeyValue.GetMapInt64Int64("ExpSkipEffcient")
	if !ok {
		panic(fmt.Errorf("key_value 表 配置错误： key ExpSkipEffcient 不存在id"))
	}
	var max values.Integer
	for k := range data {
		if k > max {
			max = k
		}
	}
	if max <= 0 {
		panic(fmt.Errorf("key_value 表 配置错误： key ExpSkipEffcient 最大值<=0"))
	}
	return max
}

func customParseAllTargetTaskTypes() map[models.TaskType]struct{} {
	record := make(map[models.TaskType]struct{})
	for _, cfg := range r.TargetTaskStage.List() {
		param := tasktarget.MustParseParam(cfg.TaskStageTargetParam)
		if _, ok := record[param.TaskType]; !ok {
			record[param.TaskType] = struct{}{}
		}
	}
	return record
}

func customParseAllMainTaskTypes() map[models.TaskType]struct{} {
	record := make(map[models.TaskType]struct{})
	for _, cfg := range r.MainTask.List() {
		for _, targetParam := range cfg.SubTask {
			param := tasktarget.MustParseParam(targetParam)
			if _, ok := record[param.TaskType]; !ok {
				record[param.TaskType] = struct{}{}
			}
		}
	}
	return record
}

func customParseAllNpcTaskTypes() map[models.TaskType]struct{} {
	record := make(map[models.TaskType]struct{})
	for _, cfg := range r.NpcTask.List() {
		for _, targetParam := range cfg.SubTask {
			param := tasktarget.MustParseParam(targetParam)
			if _, ok := record[param.TaskType]; !ok {
				record[param.TaskType] = struct{}{}
			}
		}
		for _, acceptParam := range cfg.AcceptTaskParam {
			if len(acceptParam) == 0 {
				continue
			}
			param := tasktarget.MustParseParam(acceptParam)
			if _, ok := record[param.TaskType]; !ok {
				record[param.TaskType] = struct{}{}
			}
		}
	}
	return record
}

func customParseAllLoopTaskTypes() map[models.TaskType]struct{} {
	record := make(map[models.TaskType]struct{})
	for _, cfg := range r.Looptask.List() {
		param := tasktarget.MustParseParam(cfg.TaskTargetParam)
		if _, ok := record[param.TaskType]; !ok {
			record[param.TaskType] = struct{}{}
		}
		if len(cfg.UnlockCondition) < 3 {
			continue
		}
		param = tasktarget.MustParseParam(cfg.UnlockCondition)
		if _, ok := record[param.TaskType]; !ok {
			record[param.TaskType] = struct{}{}
		}
	}
	return record
}

func customParseSkill() map[values.HeroSkillId]values.HeroSkillId {
	maxSkill := make(map[values.HeroSkillId]values.HeroSkillId)
	for _, el := range h.skill {
		if maxSkill[el.SkillBaseId] < el.Id {
			maxSkill[el.SkillBaseId] = el.Id
		}
	}
	return maxSkill
}

func customParseAllLimitedTimePackageConditions() map[models.TaskType]map[values.Integer]struct{} {
	record := make(map[models.TaskType]map[values.Integer]struct{})
	for _, cfg := range r.ActivityLimitedtimePackage.List() {
		param := tasktarget.MustParseParam(cfg.PackageConditions)
		if targets, ok := record[param.TaskType]; !ok {
			record[param.TaskType] = map[values.Integer]struct{}{
				param.Target: {},
			}
		} else {
			targets[param.Target] = struct{}{}
		}
	}
	return record
}

func customParseLimitedTimePackagePayList() map[values.Integer][]ActivityLimitedtimePackagePay {
	ret := make(map[values.Integer][]ActivityLimitedtimePackagePay)
	for _, cfg := range r.ActivityLimitedtimePackagePay.List() {
		if _, ok := ret[cfg.ActivityLimitedtimePackageId]; ok {
			ret[cfg.ActivityLimitedtimePackageId] = append(ret[cfg.ActivityLimitedtimePackageId], cfg)
		} else {
			ret[cfg.ActivityLimitedtimePackageId] = []ActivityLimitedtimePackagePay{cfg}
		}
	}
	return ret
}

func customParseEquipResonance() map[values.HeroId]map[values.Level]EquipResonance {
	dataMap := make(map[values.HeroId]map[values.Level]EquipResonance)
	for _, resonance := range h.equipResonance {
		if _, ok := dataMap[resonance.HeroId]; !ok {
			dataMap[resonance.HeroId] = make(map[values.Level]EquipResonance)
		}
		dataMap[resonance.HeroId][resonance.StrLv] = resonance
	}
	return dataMap
}

func customParsePersonalBossAttr() map[int64]MonsterAttr {
	dataMap := make(map[int64]MonsterAttr)
	for _, cfg := range h.monsterAttr {
		if cfg.MonsterType != 3 {
			continue
		}
		dataMap[cfg.MonsterLv] = cfg
	}
	return dataMap
}

// map<floor, []PersonalBossLv>
func customParsePersonalBossLv() map[int64][]PersonalBossLv {
	dataMap := make(map[int64][]PersonalBossLv)
	for _, cfg := range h.personalBossLv {
		if _, ok := dataMap[cfg.PersonalBossNumberFloorId]; !ok {
			dataMap[cfg.PersonalBossNumberFloorId] = []PersonalBossLv{cfg}
		}
		dataMap[cfg.PersonalBossNumberFloorId] = append(dataMap[cfg.PersonalBossNumberFloorId], cfg)
	}
	return dataMap
}

func customParseExpeditionReward() map[values.Integer][]ExpeditionReward {
	dataMap := make(map[values.Integer][]ExpeditionReward)
	for parentId, item := range h.expeditionRewardMap {
		if _, ok := dataMap[parentId]; !ok {
			dataMap[parentId] = make([]ExpeditionReward, 0)
		}
		for _, index := range item {
			dataMap[parentId] = append(dataMap[parentId], h.expeditionReward[index])
		}
	}
	for id := range dataMap {
		if len(dataMap[id]) <= 0 {
			panic(fmt.Errorf("expedition_reward 表 配置错误： %d 任务未配置奖励数据", id))
		}
		sort.Slice(dataMap[id], func(i, j int) bool {
			return dataMap[id][i].Id > dataMap[id][j].Id
		})
	}
	return dataMap
}
func customParseFashion() map[values.HeroId]values.FashionId {
	dataMap := make(map[values.HeroId]values.FashionId)
	for _, fashion := range h.fashion {
		if fashion.IsDefault == 1 {
			if _, ok := dataMap[fashion.Hero]; ok {
				panic(fmt.Errorf("英雄 %d 重复的初始时装配置", fashion.Hero))
			}
			dataMap[fashion.Hero] = fashion.Id
		}
	}
	for _, hero := range h.rowHero {
		if _, ok := dataMap[hero.OriginId]; !ok && hero.OriginId != 99 {
			panic(fmt.Errorf("英雄 %d 未配置初始时装", hero.OriginId))
		}
	}
	return dataMap
}

func customParseRelicsAttr() map[models.TaskType][]values.Integer {
	m := map[values.Integer]models.TaskType{}
	for _, fa := range h.relicsFunctionAttr {
		m[fa.Id] = models.TaskType(fa.RelicsType[0])
	}
	res := map[models.TaskType][]values.Integer{}
	for _, rlc := range h.relics {
		if rfa, exist := m[rlc.RelicsFunctionAttrId]; exist {
			if _, has := res[rfa]; !has {
				res[rfa] = make([]values.Integer, 0, 8)
			}
			res[rfa] = append(res[rfa], rlc.Id)
		}
	}
	return res
}

func customParseSoulSkill() map[values.HeroId]values.HeroSkillId {
	res := map[values.HeroId]values.HeroSkillId{}
	for _, hero := range h.rowHero {
		for _, skillId := range hero.SkillId {
			if skillIdx, exist := h.skillMap[skillId]; exist {
				cfg := h.skill[skillIdx]
				if cfg.IsSoulSkill {
					res[hero.Id] = cfg.Id
					break
				}
			}
		}
	}
	return res
}

func (i *CustomParse) GetSoulSkill() map[values.HeroId]values.HeroSkillId {
	return i.soulSkill
}

func (i *CustomParse) GetRelicsAttrFunc() map[models.TaskType][]values.Integer {
	return r.relicsFuncAttr
}

func customParsePersonalBossMonsterAttr() map[values.Level]MonsterAttr {
	dataMap := make(map[values.Level]MonsterAttr)
	for _, cfg := range h.monsterAttr {
		if cfg.PlayMonsterType != 2 {
			continue
		}
		dataMap[cfg.MonsterLv] = cfg
	}
	return dataMap
}

// func customParseExpeditionReward() map[values.Integer][]ExpeditionReward {
// 	dataMap := make(map[values.Integer][]ExpeditionReward)
// 	for parentId, item := range h.expeditionRewardMap {
// 		if _, ok := dataMap[parentId]; !ok {
// 			dataMap[parentId] = make([]ExpeditionReward, 0)
// 		}
// 		for _, index := range item {
// 			dataMap[parentId] = append(dataMap[parentId], h.expeditionReward[index])
// 		}
// 	}
// 	for id := range dataMap {
// 		if len(dataMap[id]) <= 0 {
// 			panic(fmt.Errorf("expedition_reward 表 配置错误： %d 任务未配置奖励数据", id))
// 		}
// 		sort.Slice(dataMap[id], func(i, j int) bool {
// 			return dataMap[id][i].Id > dataMap[id][j].Id
// 		})
// 	}
// 	return dataMap
// }

func customParseTargetTaskStageList() map[values.Integer][]TargetTaskStage {
	ret := make(map[values.Integer][]TargetTaskStage)
	for _, cfg := range r.TargetTaskStage.List() {
		if _, ok := ret[cfg.TargetTaskId]; ok {
			ret[cfg.TargetTaskId] = append(ret[cfg.TargetTaskId], cfg)
		} else {
			ret[cfg.TargetTaskId] = []TargetTaskStage{cfg}
		}
	}
	return ret
}

func customParseEquipStar() []values.ItemId {
	itemIdList := make([]values.ItemId, 0)
	temp := make(map[values.ItemId]struct{})
	for _, star := range h.equipStar {
		for id := range star.Cost {
			if _, ok := temp[id]; !ok {
				temp[id] = struct{}{}
				itemIdList = append(itemIdList, id)
			}
		}
	}
	return itemIdList
}

func customParseRacingRankReward() values.Integer {
	var max values.Integer
	for _, el := range h.combatBattle {
		if len(el.SubsectionRank) != 2 {
			panic(fmt.Errorf("combat_battle 表 SubsectionRank 列配置错误： %d", el.Id))
		}
		if el.Id > max {
			max = el.Id
		}
	}
	return max
}

func customParseGuildBlessing() values.Integer {
	var max values.Integer
	for _, el := range h.guildBlessing {
		if el.PageId > max {
			max = el.PageId
		}
	}
	return max
}
