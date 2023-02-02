package rule

import (
	"time"

	"coin-server/common/ctx"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/rule"
	rulemodel "coin-server/rule/rule-model"
)

func GetMaxGuildLevel(ctx *ctx.Context) values.Level {
	return rule.MustGetReader(ctx).Guild.MaxLevel()
}

func GetLeaderPosition(ctx *ctx.Context) values.GuildPosition {
	return rule.MustGetReader(ctx).GuildPosition.GetLeader()
}

func GetMemberPosition(ctx *ctx.Context) values.GuildPosition {
	return rule.MustGetReader(ctx).GuildPosition.GetMember()
}

func GuildPositionList(ctx *ctx.Context) []values.GuildPosition {
	return rule.MustGetReader(ctx).GuildPosition.GuildPositionList()
}

func GetPermissions(ctx *ctx.Context, p values.GuildPosition) (*rulemodel.GuildPosition, bool) {
	return rule.MustGetReader(ctx).GuildPosition.GetGuildPositionById(p)
}

func GetGuildConfigByLevel(ctx *ctx.Context, lv values.Level) (*rulemodel.Guild, bool) {
	return rule.MustGetReader(ctx).Guild.GetGuildById(lv)
}

func GetGuildCreateCost(ctx *ctx.Context) (*models.Item, bool) {
	item, ok := rule.MustGetReader(ctx).KeyValue.GetItem("GuildCreateCost")
	return item, ok
}

func GetGuildUserLevelLimit(ctx *ctx.Context) values.Level {
	// v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildUserLevelLimit")
	// return v
	// 已改为走系统解锁的配置
	return 0
}

func GetGuildApplyListLimit(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildApplyListLimit")
	return v
}

func GetGuildUserApplyListLimit(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildApplyListForUserLimit")
	return v
}

func GetApplyExpired(ctx *ctx.Context) values.TimeStamp {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildApplyListExpired")
	return time.Now().Add(time.Second * time.Duration(v)).Unix()
}

func GetApplyExpiredVal(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildApplyListExpired")
	return values.Integer(time.Duration(v) * time.Millisecond)
}

func GetRemoveCd(ctx *ctx.Context) values.TimeStamp {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildRemoveCD")
	return time.Now().Add(time.Second * time.Duration(v)).Unix()
}

func GetExitCd(ctx *ctx.Context) values.TimeStamp {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildExitCD")
	return time.Now().Add(time.Second * time.Duration(v)).Unix()
}

func GetInviteLimit(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildInviteListLimit")
	return v
}

func GetInviteExpired(ctx *ctx.Context) values.TimeStamp {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildInviteExpired")
	return time.Now().Add(time.Second * time.Duration(v)).Unix()
}

func GetGuildNameLen(ctx *ctx.Context) int {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildNameLen")
	return int(v)
}

func GetGuildIntroLen(ctx *ctx.Context) int {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildIntroLen")
	return int(v)
}

func GetGuildNoticeLen(ctx *ctx.Context) int {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildNoticeLen")
	return int(v)
}

func GetGuildPositionNameLen(ctx *ctx.Context) int {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildPosLen")
	return int(v)
}

func GetGuildGreetingLen(ctx *ctx.Context) int {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildWelcomeLen")
	return int(v)
}

func GetGuildLeaderAutoHandOverDay(ctx *ctx.Context) values.Integer {
	day, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildLeaderAutoHandOverDay")
	return day * 24 * 60 * 60
}

func GetGuildFirstJoinRewards(ctx *ctx.Context) (*models.Item, bool) {
	rewards, ok := rule.MustGetReader(ctx).KeyValue.GetItem("GuildFirIn")
	return rewards, ok
}

func GetGuildWorldFreeInvite(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildInviteFre")
	return v
}

func GetGuildWorldPayInvite(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildInvitePay")
	return v
}

func GetGuildWorldInviteCost(ctx *ctx.Context) (*models.Item, bool) {
	v, ok := rule.MustGetReader(ctx).KeyValue.GetItem("GuildInviteCost")
	return v, ok
}
func GetGuildModifyNameItemCost(ctx *ctx.Context) (*models.Item, bool) {
	v, ok := rule.MustGetReader(ctx).KeyValue.GetItem("GuildRename")
	return v, ok
}

func GetGuildModifyNameCost(ctx *ctx.Context) (*models.Item, bool) {
	v, ok := rule.MustGetReader(ctx).KeyValue.GetItem("GuildModifyNameCost")
	return v, ok
}

func GuildInviteMsgLen(ctx *ctx.Context) int {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("GuildInviteMsgLen")
	return int(v)
}

func GetMaxMemberCount(ctx *ctx.Context, level values.Level) values.Integer {
	item, ok := rule.MustGetReader(ctx).Guild.GetGuildById(level)
	if !ok {
		return 0
	}
	var count values.Integer
	for _, v := range item.MembersCount {
		count += v
	}
	return count
}

func GetMemberCountByPosition(ctx *ctx.Context, level values.Level, p values.GuildPosition) values.Integer {
	item, ok := rule.MustGetReader(ctx).Guild.GetGuildById(level)
	if !ok {
		return 0
	}
	count, _ := item.MembersCount[p]
	return count
}

func FlagCheck(ctx *ctx.Context, flag values.Integer) bool {
	list := rule.MustGetReader(ctx).GuildSign.List()
	for _, sign := range list {
		if sign.Id == flag {
			return true
		}
	}
	return false
}

func LanguageCheck(ctx *ctx.Context, lang values.Integer) bool {
	list := rule.MustGetReader(ctx).VerifyLanguage.List()
	for _, l := range list {
		if l.Id == lang {
			return true
		}
	}
	return false
}

func PositionCheck(ctx *ctx.Context, position values.GuildPosition) bool {
	list := rule.MustGetReader(ctx).GuildPosition.GuildPositionList()
	for _, p := range list {
		if p == position {
			return true
		}
	}
	return false
}

func GetBlessById(ctx *ctx.Context, id values.Integer) (*rulemodel.GuildBlessing, bool) {
	return rule.MustGetReader(ctx).GuildBlessing.GetGuildBlessingById(id)
}

func GetQueueDuration(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("BlessingTime")
	return v
}

func GetOneMemberQueueDuration(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("BlessingAddMan")
	return v
}

func GetMaxBlessStage(ctx *ctx.Context) values.Integer {
	v, _ := rule.MustGetReader(ctx).KeyValue.GetInt64("BlessingStorey")
	return v
}

func GetMaxBlessPage(ctx *ctx.Context) values.Integer {
	return rule.MustGetReader(ctx).GuildBlessing.GetMaxGuildBlessingPage()
}
