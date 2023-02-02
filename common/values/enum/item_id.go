package enum

import "coin-server/common/values"

// 特殊道具枚举、如资源类型
const (
	// Gold 金币
	Gold values.ItemId = 101
	// BoundDiamond 绑定钻石 通过系统产生或者钻石兑换获得，用于在钻石商店购买道具
	BoundDiamond values.ItemId = 102
	// Diamond 充值钻石 充值获得的货币，可以1:1兑换成绑钻
	Diamond values.ItemId = 103
	// RedPack 红包开启道具
	RedPack values.ItemId = 1000
	// DailyTaskActive 每日任务活跃度
	DailyTaskActive values.ItemId = 109
	// WeeklyTaskActive 每周任务活跃度
	WeeklyTaskActive values.ItemId = 110
	// FriendPoint 好友点数
	FriendPoint values.ItemId = 107
	// RoleExp 主角经验
	RoleExp values.ItemId = 111
	// MeltExp 装备熔炼经验
	MeltExp values.ItemId = 100013
	// RuneDust 符文粉尘
	RuneDust values.ItemId = 124
	// CureRitualSuccessRateUp 救治仪式成功率（增加）
	CureRitualSuccessRateUp values.ItemId = 120
	// CureRitualSuccessRateDown 救治仪式成功率（减少）
	CureRitualSuccessRateDown values.ItemId = 121
	CrusadersCoin                           = 104 // 学会货币
	JudegersCoin                            = 105 // 审判军货币
	RebelsCoin                              = 106 // 教团货币
	GuildCoin                               = 113 // 公会货币
	ArenaCoin                               = 116 // 基础竞技场货币
)

var CurrencyList = []values.ItemId{Gold, BoundDiamond, GuildCoin, ArenaCoin, CrusadersCoin, JudegersCoin, RebelsCoin}
