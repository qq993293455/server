package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	event2 "coin-server/game-server/event"
	rankval "coin-server/game-server/service/rank/values"
)

type RankService interface {
	CreateRank(ctx *ctx.Context, rankType enum.RankType) values.RankId
	UpdateValue(ctx *ctx.Context, value *models.RankValue) *errmsg.ErrMsg
	DeleteValue(ctx *ctx.Context, rankId values.RankId, ownerId values.GuildId) *errmsg.ErrMsg
	DeleteRank(ctx *ctx.Context, rankId values.RankId) *errmsg.ErrMsg
	ClearRanKInMem(ctx *ctx.Context, rankId values.RankId)
	InitRankToMem(ctx *ctx.Context, rankId values.RankId, rankType enum.RankType) *errmsg.ErrMsg
	MemRankValueChange(ctx *ctx.Context, data *event2.MemRankValueChangeData) *errmsg.ErrMsg

	GetRank(ctx *ctx.Context, rankId values.RankId) (rank rankval.RankAgg)
	GetTotalNum(ctx *ctx.Context, rankId values.RankId) (values.Integer, *errmsg.ErrMsg)
	GetValueById(ctx *ctx.Context, rankId values.RankId, ownerId values.GuildId) (*models.RankValue, *errmsg.ErrMsg)
	GetScoreValue1ByIds(ctx *ctx.Context, rankId values.RankId, ownerIds []values.GuildId) (map[values.GuildId]values.Integer, *errmsg.ErrMsg)
	GetScoreValue2ByIds(ctx *ctx.Context, rankId values.RankId, ownerIds []values.GuildId) (map[values.GuildId]models.RankAndScore, *errmsg.ErrMsg)
	GetValueByRank(ctx *ctx.Context, rankId values.RankId, rank values.Integer) (*models.RankValue, *errmsg.ErrMsg)
	GetValueByRange(ctx *ctx.Context, rankId values.RankId, start, end values.Integer) ([]*models.RankValue, *errmsg.ErrMsg)
}
