package event

import "coin-server/common/values"

type GotHero struct {
	OriginId values.HeroId
}
