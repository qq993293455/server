package dao

import "encoding/json"

type RankValue struct {
	RankId    string `db:"rank_id" json:"rank_id"`
	RankType  int64  `db:"rank_type" json:"rank_type"`
	OwnerId   string `db:"owner_id" json:"owner_id"`
	Rank      int64  `db:"rank" json:"rank"`
	Value1    int64  `db:"value1" json:"value1"`
	Value2    int64  `db:"value2" json:"value2"`
	Value3    int64  `db:"value3" json:"value3"`
	Value4    int64  `db:"value4" json:"value4"`
	Extra     string `db:"extra" json:"extra"`
	Shard     int64  `db:"shard" json:"shard"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
	DeletedAt int64  `db:"deleted_at" json:"deleted_at"`
}

func MarshalExtra(extra map[string]string) string {
	if len(extra) == 0 {
		return ""
	}
	b, _ := json.Marshal(extra)
	return string(b)
}

func UnmarshalExtra(extra string) map[string]string {
	ret := make(map[string]string)
	json.Unmarshal([]byte(extra), &ret)
	return ret
}
