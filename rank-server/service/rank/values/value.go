package values

import (
	"coin-server/common/proto/dao"
	"coin-server/common/values"
)

type TopRankItem struct {
	RoleId values.RoleId
	*dao.TopRankItem
}
