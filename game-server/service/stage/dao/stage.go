package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/utils"
	"coin-server/common/values"
	"coin-server/rule"
	rule_model "coin-server/rule/rule-model"
)

const DoneVal int64 = 3

func GetStage(ctx *ctx.Context) (*dao.Stage, *errmsg.ErrMsg) {
	res := &dao.Stage{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	if !ok {
		res = &dao.Stage{
			RoleId:    ctx.RoleId,
			CurrStage: 0,
			TypeCount: map[int64]int64{},
		}
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), res)
	}
	return res, nil
}

func SaveStage(ctx *ctx.Context, i *dao.Stage) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), i)
}

func GetFirstStageID(c *ctx.Context) *rule_model.MapSelect {
	return &rule.MustGetReader(c).MapSelect.List()[0]
}

func GetExploreById(ctx *ctx.Context, roleId values.RoleId, stageId values.Integer) (*dao.StageExplore, *errmsg.ErrMsg) {
	explore := &dao.StageExplore{StageId: stageId}
	_, err := ctx.NewOrm().HGetPB(redisclient.GetUserRedis(), getExploreKey(roleId), explore)
	if err != nil {
		return nil, err
	}
	if explore.Explore == nil {
		explore.Explore = map[int64]int64{}
	}
	return explore, err
}

func GetExplore(ctx *ctx.Context, roleId values.RoleId) ([]*dao.StageExplore, *errmsg.ErrMsg) {
	explore := make([]*dao.StageExplore, 0)
	err := ctx.NewOrm().HGetAll(redisclient.GetUserRedis(), getExploreKey(roleId), &explore)
	if err != nil {
		return nil, err
	}
	return explore, nil
}

func SaveExplore(ctx *ctx.Context, roleId values.RoleId, explore *dao.StageExplore) {
	ctx.NewOrm().HSetPB(redisclient.GetUserRedis(), getExploreKey(roleId), explore)
	return
}

func getExploreKey(roleId values.RoleId) string {
	return utils.GenDefaultRedisKey(values.StageExplore, values.Hash, roleId)
}
