package event

import "coin-server/common/values"

type RLFinishEvent struct {
	RoguelikeId values.RoguelikeId
	IsSucc      bool
}
