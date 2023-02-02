package idgenerate

type IDGenKey string

const (
	RoleIDKey      IDGenKey = "GEN_ROLE_ID_KEY"
	GuideIDKey     IDGenKey = "GEN_GUIDE_ID_KEY"
	LoadTestUid    IDGenKey = "GEN_LOAD_TEST_UID" // 压力测试uid
	MatchRoomUid   IDGenKey = "GEN_MATCH_ROOM_UID"
	CenterBattleId IDGenKey = "GEN_CENTER_BATTLE_ID"
)

var initValue = map[IDGenKey]int64{
	RoleIDKey:      1000000,
	GuideIDKey:     100,
	LoadTestUid:    10000,
	MatchRoomUid:   100000,
	CenterBattleId: 100000,
}
