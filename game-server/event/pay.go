package event

import (
	"coin-server/common/values"
)

type PaySuccess struct {
	PcId       values.Integer // 购买项id（charge表里的id）
	PaidTime   values.Integer // 玩家付款时间（秒）
	ExpireTime values.Integer // 过期时间（仅订阅有效，秒）
}
