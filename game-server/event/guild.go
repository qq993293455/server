package event

import "coin-server/common/values"

type BlessActivated struct {
	Stage     int64
	Page      int64
	Activated []int64
}

type GuildAction = int64

const (
	GuildActionJoin  GuildAction = 1
	GuildActionLeave GuildAction = 1
)

// GuildEvent 公会事件
type GuildEvent struct {
	RoleId  values.RoleId
	GuildId values.GuildId
	Action  GuildAction
}
