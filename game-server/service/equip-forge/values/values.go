package values

import "coin-server/common/values"

const InitEquipForgeLevel = 1 // 玩家初始打造等级

type Product struct {
	BoxId   values.Integer
	Quality values.Quality
	Weight  values.Integer
}
