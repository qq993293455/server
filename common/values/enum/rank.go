package enum

type RankType int64

func (e RankType) ToInt64() int64 {
	return int64(e)
}

const (
	RankNormal RankType = 1
	RankGuild  RankType = 2 // 公会
	RankArena  RankType = 3 // 竞技场(默认)
)
