package utils

import (
	"fmt"
	"strings"

	"coin-server/common/values"
)

// 例如:     item:k:E4UZN
func GenRedisKey(moduleName values.ModuleName, keyType values.RedisKeyType, roleId values.RoleId, extra string) string {
	bufLen := len(moduleName) + 1 + 1 + 1 + len(roleId)
	lExtra := len(extra)
	if lExtra > 0 {
		bufLen += 1 + lExtra
	}
	bs := make([]byte, bufLen)
	bl := 0
	bl += copy(bs[bl:], moduleName)
	bl += copy(bs[bl:], ":")
	bl += copy(bs[bl:], keyType)
	bl += copy(bs[bl:], ":")
	bl += copy(bs[bl:], roleId)
	if lExtra > 0 {
		bl += copy(bs[bl:], ":")
		bl += copy(bs[bl:], extra)
	}
	return string(bs)
}

func GenDefaultRedisKey(moduleName values.ModuleName, keyType values.RedisKeyType, roleId values.RoleId) string {
	return GenRedisKey(moduleName, keyType, roleId, "")
}

// 比如从 item:k:E4UZN 中提取出 item 和 k  和 E4UZN
func ExtractFromRedisKey(str string) (name values.ModuleName, keyType values.RedisKeyType, roleId values.RoleId, err error) {
	strList := strings.Split(str, ":")
	if len(strList) < 3 {
		err = fmt.Errorf("invalid redis key length!  %s", str)
		return
	}
	name = values.ModuleName(strList[0])
	keyType = values.RedisKeyType(strList[1])
	roleId = values.RoleId(strList[2])
	return
}
