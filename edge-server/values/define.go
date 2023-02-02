package values

import (
	"coin-server/edge-server/children"
)

type KillSelfEvent struct {
	Child *children.Child
}
