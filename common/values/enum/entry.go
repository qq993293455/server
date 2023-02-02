package enum

import "coin-server/common/values"

// 词条枚举

type EntryId = values.Integer

const (
	EntryDivinationCountIncr EntryId = 3 // 每日占卜次数增加
	EntryUpgradeExpReduce    EntryId = 4 // 升级所需经验减少
	EntryIdleExpIncr         EntryId = 7 // 挂机经验提高
	EntryIdleExpIncrPercent  EntryId = 8 // 挂机经验提高 万分比
	//EntryIdleGoldIncrPercent EntryId = 9 // 挂机金币收益提高 万分比
)

// EntryListIdle 词条列表 挂机相关
func EntryListIdle() []EntryId {
	return []EntryId{
		EntryIdleExpIncr,
		EntryIdleExpIncrPercent,
		//EntryIdleGoldIncrPercent,
	}
}
