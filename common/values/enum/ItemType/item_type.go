package ItemType

import "coin-server/common/values"

// 道具类型枚举
const (
	Materials  values.ItemType = 1  // 材料
	Medicine   values.ItemType = 2  // 药
	Equipment  values.ItemType = 5  // 装备
	Relics     values.ItemType = 6  // 遗物
	SkillStone values.ItemType = 8  // 技能石
	TalentRune values.ItemType = 11 // 天赋符文

	Hero            values.ItemType = 34 // 英雄转换道具 获得后直接转为英雄
	FashionActivate values.ItemType = 36 // 时装激活道具
	Avatar          values.ItemType = 37 // 头像&头像框转换道具 获得后直接转为头像&头像框
)
