package dao

/*import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/game-server/service/achievement/values"
)

func GetCounter(ctx *ctx.Context) (values.CounterI, *errmsg.ErrMsg) {
	res := &dao.AchievementCounter{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetCtxRedis(ctx), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.AchievementCounter{
			RoleId: ctx.RoleId,
			Cnt:    map[int64]int64{},
		}
		ctx.NewOrm().SetPB(redisclient.GetCtxRedis(ctx), res)
	}
	if res.Cnt == nil {
		res.Cnt = map[int64]int64{}
	}
	return values.NewCounter(res), nil
}

func SaveCounter(ctx *ctx.Context, i values.CounterI) {
	ctx.NewOrm().SetPB(redisclient.GetCtxRedis(ctx), i.ToDao())
}
*/
