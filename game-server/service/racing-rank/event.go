package racing_rank

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/game-server/event"
	"coin-server/game-server/service/racing-rank/dao"
)

func (svc *Service) HandleLogin(ctx *ctx.Context, d *event.Login) *errmsg.ErrMsg {
	status, err := dao.GetStatus(ctx)
	if err != nil {
		return err
	}
	// 检查匹配结果
	if err := svc.matchCheck(ctx, status); err != nil {
		return err
	}
	// // 处理赛季结算
	// saveStatus, err := svc.handleSettlement(ctx, status)
	// if err != nil {
	// 	return err
	// }
	// if saveStatus {
	// 	if err := dao.SaveStatus(ctx, status); err != nil {
	// 		return err
	// 	}
	// }
	// // 检查MySQL是否存在结算时间节点
	// endTime, err := dao.GetEndTime(ctx.RoleId)
	// if err != nil {
	// 	return err
	// }
	// max := values.Integer(5 * 60 * 1000)
	// // endTime>0表示存在结算时间节点，如果结算时间点已超过5分钟，则删除结算时间节点
	// // 正常情况下racingrank-server那边会在endTime的时候做结算，并删除掉结算时间节点
	// if endTime > 0 && timer.StartTime(ctx.StartTime).Unix()-endTime > max {
	// 	if err := dao.DeleteEndTime(ctx.RoleId); err != nil {
	// 		return err
	// 	}
	// } else if endTime > timer.StartTime(ctx.StartTime).Unix() {
	// 	return svc.taskCheck(ctx, endTime)
	// }
	return nil
}
