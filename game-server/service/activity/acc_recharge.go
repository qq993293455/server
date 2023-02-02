package activity

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/models"
	protosvc "coin-server/common/proto/service"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/activity/dao"
	"coin-server/rule"
)

func (svc *Service) AccRechargeData(c *ctx.Context, _ *protosvc.Activity_AccRechargeDataRequest) (*protosvc.Activity_AccRechargeDataResponse, *errmsg.ErrMsg) {
	role, err := svc.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	data, err := dao.GetAccRecharge(c)
	if err != nil {
		return nil, err
	}
	list := rule.MustGetReader(c).ActivityAccumulatedrecharge.List()
	cfg := make([]*models.AccRechargeCfg, 0, len(list))
	for _, v := range list {
		cfg = append(cfg, &models.AccRechargeCfg{
			Id:             v.Id,
			DemandAmount:   v.DemandAmount,
			RechargeReward: v.RechargeReward,
		})
	}
	return &protosvc.Activity_AccRechargeDataResponse{
		TotalRecharge: role.Recharge,
		DrawList:      data.DrawList,
		Cfg:           cfg,
	}, nil
}

func (svc *Service) AccRechargeDraw(c *ctx.Context, req *protosvc.Activity_AccRechargeDrawRequest) (*protosvc.Activity_AccRechargeDrawResponse, *errmsg.ErrMsg) {
	reader := rule.MustGetReader(c)
	rechargeAcc, ok := reader.ActivityAccumulatedrecharge.GetActivityAccumulatedrechargeById(req.Idx + 1)
	if !ok {
		return nil, errmsg.NewErrActivityGiftNotExist()
	}
	role, err := svc.GetRoleByRoleId(c, c.RoleId)
	if err != nil {
		return nil, err
	}
	data, err := dao.GetAccRecharge(c)
	if err != nil {
		return nil, err
	}
	if role.Recharge < rechargeAcc.DemandAmount {
		return nil, errmsg.NewErrActivityNeedPaid()
	}
	if _, exist := data.DrawList[req.Idx]; exist {
		return nil, errmsg.NewErrEventDoNotCollectAgain()
	}
	data.DrawList[req.Idx] = true
	rewards := make(map[values.ItemId]values.Integer, 0)
	for i := 0; i < len(rechargeAcc.RechargeReward); i += 2 {
		rewards[rechargeAcc.RechargeReward[i]] = rechargeAcc.RechargeReward[i+1]
	}
	if _, err := svc.AddManyItem(c, c.RoleId, rewards); err != nil {
		return nil, err
	}
	if len(data.DrawList) == reader.ActivityAccumulatedrecharge.Len() {
		c.PushMessage(&protosvc.Activity_ActivityEndedPush{
			Activity: &models.Activity{
				Id: enum.AccRecharge,
			},
		})
	}
	dao.SaveAccRecharge(c, data)
	return &protosvc.Activity_AccRechargeDrawResponse{}, nil
}

func (svc *Service) showAccRecharge(c *ctx.Context) bool {
	data, err := dao.GetAccRecharge(c)
	if err != nil {
		return false
	}
	if len(data.DrawList) == rule.MustGetReader(c).ActivityAccumulatedrecharge.Len() {
		return false
	}
	return true
}

func (svc *Service) CheatRechargeRequest(c *ctx.Context, req *protosvc.Activity_CheatRechargeRequest) (*protosvc.Activity_CheatRechargeResponse, *errmsg.ErrMsg) {
	c.PublishEventLocal(&event.RechargeAmountEvt{
		Amount: req.Amount,
	})
	return &protosvc.Activity_CheatRechargeResponse{}, nil
}
