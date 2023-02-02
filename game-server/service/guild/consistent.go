package guild

import (
	"coin-server/common/values"
	"coin-server/common/values/enum"

	"stathat.com/c/consistent"
)

var c = consistent.New()

func init() {
	for _, k := range enum.GuildAllIdKey {
		c.Add(k)
	}
}

func getGuildIdKey(id values.GuildId) string {
	key, err := c.Get(id)
	if err != nil {
		return enum.GuildAllIdKey[0]
	}
	return key
}
