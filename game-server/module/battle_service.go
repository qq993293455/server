package module

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	servicepb "coin-server/common/proto/service"
	"coin-server/common/values"
)

type BattleService interface {
	NicknameChange(c *ctx.Context, nickName string) *errmsg.ErrMsg
	TempBagExpProfit(c *ctx.Context, roleId string) (float64, *errmsg.ErrMsg) // 获取临时背包每秒经验收益
	GetCurrBattleInfo(c *ctx.Context, req *servicepb.GameBattle_GetCurrBattleInfoRequest) (*servicepb.GameBattle_GetCurrBattleInfoResponse, *errmsg.ErrMsg)
	GetCurBattleSrvId(c *ctx.Context) (values.Integer, *errmsg.ErrMsg)
}
