package rule

import (
	"errors"
	"fmt"
	"math"

	"coin-server/common/utils"
	wr "coin-server/common/utils/weightedrand"
	"coin-server/common/values"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"
)

//MustGetPersonalBossEventResetTime 个人BOSS重置间隔时间 天
func MustGetPersonalBossEventResetTime() int {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("PersonalBossEventResetTime") // 天
	if !ok {
		panic(errors.New("KeyValue PersonalBossEventResetTime not found"))
	}
	return int(val)
}

//MustGetPersonalBossSharingInterval 个人BOSS分享间隔
func MustGetPersonalBossSharingInterval() int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("PersonalBossSharingInterval") // 秒
	if !ok {
		panic(errors.New("KeyValue PersonalBossSharingInterval not found"))
	}
	return val
}

//MustGetPersonalBossGetHelpPointsUpperLimit 获得助战点上限
func MustGetPersonalBossGetHelpPointsUpperLimit() int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("PersonalBossGetHelpPointsUpperLimit")
	if !ok {
		panic(errors.New("KeyValue PersonalBossGetHelpPointsUpperLimit not found"))
	}
	return val
}

//MustGetPersonalBossHelpByPointsNumber 可被助战的次数
func MustGetPersonalBossHelpByPointsNumber() int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("PersonalBossHelpByPointsNumber")
	if !ok {
		panic(errors.New("KeyValue PersonalBossHelpByPointsNumber not found"))
	}
	return val
}

//MustGetPersonalBossHelpPointsNumber 可助战的次数
func MustGetPersonalBossHelpPointsNumber() int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("PersonalBossHelpPointsNumber")
	if !ok {
		panic(errors.New("KeyValue PersonalBossHelpPointsNumber not found"))
	}
	return val
}

//MustGetPersonalBossHelpPoints 助战一次获得的点数
func MustGetPersonalBossHelpPoints() int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("PersonalBossHelpPoints")
	if !ok {
		panic(errors.New("KeyValue PersonalBossHelpPoints not found"))
	}
	return val
}

func RandHelpReward(lv int64) int64 {
	cfgList := rule.MustGetReader(nil).PersonalBossHelpPointsReward.List()
	temp := cfgList[0]
	for _, cfg := range cfgList {
		if len(cfg.PlayerLevelRange) < 2 {
			panic(errors.New(fmt.Sprintf("PersonalBossHelpPointsReward PlayerLevelRange config failed. id: %d", cfg.Id)))
		}
		if cfg.PlayerLevelRange[1] == -1 {
			cfg.PlayerLevelRange[1] = math.MaxInt64
		}
		if lv >= cfg.PlayerLevelRange[0] && lv <= cfg.PlayerLevelRange[1] {
			temp = cfg
			break
		}
	}
	choices := make([]*wr.Choice[values.ItemId, int64], 0, len(temp.HelpPointsTreasureChest))
	for itemId, weight := range temp.HelpPointsTreasureChest {
		choices = append(choices, wr.NewChoice(itemId, weight))
	}
	ch, err := wr.NewChooser(choices...)
	utils.Must(err)
	return ch.Pick()
}

func MustGetPersonalBossHelpPointsExchangePrice() map[int64]int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetMapInt64Int64("PersonalBosstHelpPointsExchangePrice")
	if !ok {
		panic(errors.New("KeyValue PersonalBosstHelpPointsExchangePrice not found"))
	}
	return val
}

func MustGetPersonalBossEventCountdownTime() int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("PersonalBossEventCountdownTime") // 秒
	if !ok {
		panic(errors.New("KeyValue PersonalBossEventCountdownTime not found"))
	}
	return val
}

//CalCardsPoint 计算所需助战点
func CalCardsPoint(cp map[int64]int64) (map[int64]int64, int64) {
	reader := rule.MustGetReader(nil)
	var total int64
	for k := range cp {
		cfg, ok := reader.PersonalBossHelperCards.GetPersonalBossHelperCardsById(k)
		if !ok {
			panic(fmt.Sprintf("PersonalBossHelperCards config not found. id: %d", k))
		}
		cp[k] = cfg.CardConsumption
		total += cfg.CardConsumption
	}
	return cp, total
}

func GetPersonalBossLvByRoleLv(floor int64, roleLv values.Level) rule_model.PersonalBossLv {
	reader := rule.MustGetReader(nil)
	pbl := reader.PersonalBossLv.GetPersonalBossLvByFloor(floor)
	for _, cfg := range pbl {
		if len(cfg.PlayerLevelRange) < 2 {
			panic(errors.New(fmt.Sprintf("PersonalBossLv PlayerLevelRange config failed. id: %d", cfg.Id)))
		}
		if cfg.PlayerLevelRange[1] == -1 {
			cfg.PlayerLevelRange[1] = math.MaxInt64
		}
		if roleLv >= cfg.PlayerLevelRange[0] && roleLv <= cfg.PlayerLevelRange[1] {
			return cfg
		}
	}
	panic(errors.New(fmt.Sprintf("PersonalBossLv not found. floor: %d, role_lv: %d", floor, roleLv)))
}

func GetPersonalBossLibrary() []rule_model.PersonalBossLibrary {
	bosses := make([]rule_model.PersonalBossLibrary, 0)
	for _, b := range rule.MustGetReader(nil).PersonalBossLibrary.List() {
		if b.WhetherToUseBoss != 1 { // "是否使用该boss 1、使用 2、不使用"
			continue
		}
		bosses = append(bosses, b)
	}
	return bosses
}

// MustGetInitialValueOfHelperPoints 个人boss助战点数默认值
func MustGetInitialValueOfHelperPoints() int64 {
	val, ok := rule.MustGetReader(nil).KeyValue.GetInt64("InitialValueOfHelperPoints")
	if !ok {
		panic(errors.New("KeyValue InitialValueOfHelperPoints not found"))
	}
	return val
}
