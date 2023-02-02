package enum

import (
	"strconv"
	"strings"

	"coin-server/common/values"
)

const (
	ROLE_ID_BASE = "user_id_key"
	ROLE_ID_MOD  = 1 << 10
	INIT_LEVEL   = 1 // 角色初始等级
)

const DefaultLanguage = int64(10) // EN

// RoleAllIdKey 玩家id集合的key
var RoleAllIdKey [ROLE_ID_MOD]string

func init() {
	sb := strings.Builder{}
	for i := 0; i < ROLE_ID_MOD; i++ {
		sb.WriteString(ROLE_ID_BASE)
		sb.WriteString(strconv.Itoa(i))
		RoleAllIdKey[i] = sb.String()
		sb.Reset()
	}
}

func GetRoleIdKey(id values.Integer) string {
	slot := id & (ROLE_ID_MOD - 1)
	return RoleAllIdKey[slot]
}
