package values

type RankCompare struct {}

func (rc *RankCompare) Equal(a, b int64) bool {
	return a == b
}

func (rc *RankCompare) Less(a, b int64) bool {
	return a < b
}
