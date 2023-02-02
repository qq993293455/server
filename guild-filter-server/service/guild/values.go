package guild

import (
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/values"
)

type Guild struct {
	Id               values.GuildId
	Name             string
	Level            values.Level
	Lang             values.Integer
	CombatValueLimit values.Integer
	AutoJoin         bool
	Active           values.Integer
	Full             bool
	Count            int64
}

func NewGuild(g *pbdao.Guild) Guild {
	return Guild{
		Id:               g.Id,
		Name:             g.Name,
		Level:            g.Level,
		Lang:             g.Lang,
		CombatValueLimit: g.CombatValueLimit,
		AutoJoin:         g.AutoJoin,
		Active:           0,
		Full:             false,
		Count:            g.Count,
	}
}
