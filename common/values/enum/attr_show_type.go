package enum

import "coin-server/common/values"

type AttrShowType = values.Integer

const (
	Direct  AttrShowType = 1
	Percent AttrShowType = 2
)
