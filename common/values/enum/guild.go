package enum

import "strconv"

// GuildAllIdKey 公会id集合的key
var GuildAllIdKey = genKeys()

func genKeys() []string {
	count := 5
	list := make([]string, 0, count)
	for i := 0; i < count; i++ {
		list = append(list, "guild_id_key0"+strconv.Itoa(i+1))
	}
	return list
}
