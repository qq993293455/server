package rule

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/logger"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
	"math/rand"
	"strconv"

	"go.uber.org/zap"
)

const RobotHead = "R:R:@:"

type ArenaPublicConfig struct {
	ArenaPlayerLimit        int              // 基础竞技场人数上限
	ArenaTicket             int              // 基础竞技场挑战道具
	ArenaRobotLimit         int              // 基础竞技场机器人数量
	ArenaFreeNum            int              // 基础竞技场每日免费挑战次数
	ArenaTicketCost         []values.Integer // 基础竞技场挑战道具购买次数/消耗道具/消耗数量
	ArenaMatchingParam      []float64        // 基础竞技场挑战匹配系数
	ArenaChallengeRefreshCD int              // 基础竞技场挑战刷新冷却时间（秒）
	ArenaVictoryReward      []*models.Item   // 基础竞技场胜利奖励
	ArenaLoseReward         []*models.Item   // 基础竞技场失败奖励
	ArenaDailyPrizeTime     int              // 基础竞技场每日奖励结算时间
	ArenaSeasonPrizeTime    []values.Integer // 基础竞技场赛季奖励结算时间
	ArenaBattleDuration     int              // 基础竞技场战斗最大时长（秒）
	DefaultRefreshTime      int              // 系统刷新时间 0点+DefaultRefreshTime
	ArenaDailyMail          int64            // 基础竞技场每日结算邮件模板
	ArenaSeasonMail         int64            // 基础竞技场赛季结算邮件模板
	ArenaOpeningDuration    int64            // 基础竞技场单个开放时长（天）
}

func GetAllRobot(ctx *ctx.Context) *[]rulemodel.PvPRobot {
	config := rule.MustGetReader(ctx).PvPRobot.List()[0:]
	return &config
}

func GetRobot(ctx *ctx.Context, id string) (*rulemodel.PvPRobot, bool) {
	if len(id) <= len(RobotHead) {
		return nil, false
	}

	robotId := id[len(RobotHead):]
	roleId, error := strconv.ParseInt(robotId, 10, 64)
	if error != nil {
		return nil, false
	}

	config, ok := rule.MustGetReader(ctx).PvPRobot.GetPvPRobotById(values.Integer(roleId))
	if !ok {
		return nil, ok
	}
	return config, ok
}

func GetRobotName(ctx *ctx.Context) (string, bool) {
	name_lise := rule.MustGetReader(ctx).RobotName.List()
	size := len(name_lise)
	if size == 0 {
		return "", false
	}
	first_index := rand.Intn(size)
	second_index := rand.Intn(size)
	return name_lise[first_index].FirstName + " " + name_lise[second_index].SecondName, true
}

func GetRobotHeroInfo(ctx *ctx.Context, hero_id int64) (int64, bool) {
	config, ok := rule.MustGetReader(ctx).RowHero.GetRowHeroById(hero_id)
	if !ok {
		return 0, false
	}
	return config.OriginId, true
}

func GetVirtualCombatDuelDuration(ctx *ctx.Context) (int64, bool) {
	countDown, ok := rule.MustGetReader(ctx).KeyValue.GetInt64("DuelDuration")
	return countDown, ok
}

func GetArenaDefalutPublicInfo(ctx *ctx.Context, log *logger.Logger) *ArenaPublicConfig {
	var ok bool = false
	var config *ArenaPublicConfig = new(ArenaPublicConfig)

	config.ArenaPlayerLimit, ok = rule.MustGetReader(ctx).KeyValue.GetInt("ArenaPlayerLimit")
	if !ok {
		log.Error("ArenaPlayerLimit error")
		return nil
	}
	config.ArenaTicket, ok = rule.MustGetReader(ctx).KeyValue.GetInt("ArenaTicket")
	if !ok {
		log.Error("ArenaTicket error")
		return nil
	}
	config.ArenaRobotLimit, ok = rule.MustGetReader(ctx).KeyValue.GetInt("ArenaRobotLimit")
	if !ok {
		log.Error("ArenaRobotLimit error")
		return nil
	}
	config.ArenaFreeNum, ok = rule.MustGetReader(ctx).KeyValue.GetInt("ArenaFreeNum")
	if !ok {
		log.Error("ArenaFreeNum error")
		return nil
	}
	config.ArenaTicketCost, ok = rule.MustGetReader(ctx).KeyValue.GetIntegerArray("ArenaTicketCost")
	if !ok {
		log.Error("ArenaTicketCost error")
		return nil
	}
	if len(config.ArenaTicketCost) != 3 {
		log.Error("ArenaTicketCost error", zap.Any("len", len(config.ArenaTicketCost)))
		return nil
	}
	config.ArenaMatchingParam, ok = rule.MustGetReader(ctx).KeyValue.GetFloatArray("ArenaMatchingParam")
	if !ok {
		log.Error("ArenaMatchingParam error")
		return nil
	}
	if len(config.ArenaMatchingParam) != 2 {
		log.Error("ArenaMatchingParam error", zap.Any("len", len(config.ArenaMatchingParam)))
		return nil
	}
	config.ArenaChallengeRefreshCD, ok = rule.MustGetReader(ctx).KeyValue.GetInt("ArenaChallengeRefreshCD")
	if !ok {
		log.Error("ArenaChallengeRefreshCD error")
		return nil
	}
	config.ArenaVictoryReward, ok = rule.MustGetReader(ctx).KeyValue.GetItemArray("ArenaVictoryReward")
	if !ok {
		log.Error("ArenaVictoryReward error")
		return nil
	}
	config.ArenaLoseReward, ok = rule.MustGetReader(ctx).KeyValue.GetItemArray("ArenaLoseReward")
	if !ok {
		log.Error("ArenaLoseReward error")
		return nil
	}
	config.ArenaDailyPrizeTime, ok = rule.MustGetReader(ctx).KeyValue.GetInt("ArenaDailyPrizeTime")
	if !ok {
		log.Error("ArenaDailyPrizeTime error")
		return nil
	}
	config.ArenaBattleDuration, ok = rule.MustGetReader(ctx).KeyValue.GetInt("ArenaBattleDuration")
	if !ok {
		log.Error("ArenaBattleDuration error")
		return nil
	}
	config.ArenaSeasonPrizeTime, ok = rule.MustGetReader(ctx).KeyValue.GetIntegerArray("ArenaSeasonPrizeTime")
	if !ok {
		log.Error("ArenaSeasonPrizeTime error")
		return nil
	}
	if len(config.ArenaSeasonPrizeTime) != 2 {
		log.Error("ArenaSeasonPrizeTime error", zap.Any("len", len(config.ArenaSeasonPrizeTime)))
		return nil
	}
	config.DefaultRefreshTime, ok = rule.MustGetReader(ctx).KeyValue.GetInt("DefaultRefreshTime")
	if !ok {
		log.Error("DefaultRefreshTime error")
		return nil
	}
	config.ArenaDailyMail, ok = rule.MustGetReader(ctx).KeyValue.GetInt64("ArenaDailyMail")
	if !ok {
		log.Error("ArenaDailyMail error")
		return nil
	}
	config.ArenaSeasonMail, ok = rule.MustGetReader(ctx).KeyValue.GetInt64("ArenaSeasonMail")
	if !ok {
		log.Error("ArenaSeasonMail error")
		return nil
	}

	config.ArenaOpeningDuration, ok = rule.MustGetReader(ctx).KeyValue.GetInt64("ArenaOpeningDuration")
	if !ok {
		log.Error("ArenaOpeningDuration error")
		return nil
	}
	config.ArenaOpeningDuration = config.ArenaOpeningDuration * 86400

	return config
}

func GetRobotLimit(atype models.ArenaType, log *logger.Logger) (int64, *errmsg.ErrMsg) {
	if atype == models.ArenaType_ArenaType_Default {
		config := GetArenaDefalutPublicInfo(ctx.GetContext(), log)
		if config == nil {
			return 0, errmsg.NewErrArenaConfig()
		}
		return int64(config.ArenaRobotLimit), nil
	}
	return 0, errmsg.NewErrArenaType()
}

func GetPlayerLimit(atype models.ArenaType, log *logger.Logger) (int64, *errmsg.ErrMsg) {
	if atype == models.ArenaType_ArenaType_Default {
		config := GetArenaDefalutPublicInfo(ctx.GetContext(), log)
		if config == nil {
			return 0, errmsg.NewErrArenaConfig()
		}
		return int64(config.ArenaPlayerLimit), nil
	}
	return 0, errmsg.NewErrArenaType()
}

func GetAllPvpRankReward(ctx *ctx.Context) []rulemodel.PvPRankReward {
	configs := rule.MustGetReader(ctx).PvPRankReward.List()[0:]
	return configs
}
