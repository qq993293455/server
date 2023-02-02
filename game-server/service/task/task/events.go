package task

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	lesssvcpb "coin-server/common/proto/less_service"
	"coin-server/common/values"
	event2 "coin-server/common/values/event"
	"coin-server/game-server/event"
	"coin-server/game-server/service/task/task/dao"
	"coin-server/rule"
)

// HandleUpdateTarget 接收各系统发出的打点目标更新事件
func (s *Service) HandleUpdateTarget(ctx *ctx.Context, e *event.UpdateTarget) *errmsg.ErrMsg {
	if e.RoleId == ctx.RoleId {
		// 是自己的计数
		r := rule.MustGetReader(ctx)
		cfg, ok := r.TaskType.GetTaskTypeById(values.Integer(e.Typ))
		if !ok {
			return nil
		}
		// 如果是累计任务，更新计数器，发出累计数目
		if cfg.IsAccumulate {
			// 更新计数器
			counter, err := dao.GetCondByType(ctx, e.RoleId, e.Typ)
			if err != nil {
				return err
			}
			incr := e.Count
			if e.Replace {
				incr = e.Count - counter.Count[e.Id]
				counter.Count[e.Id] = e.Count
			} else {
				counter.Count[e.Id] += e.Count
			}
			dao.SaveCond(ctx, e.RoleId, counter)
			// 发出目标更新事件
			ctx.PublishEventLocal(&event.TargetUpdate{
				Typ:          e.Typ,
				Id:           e.Id,
				Count:        counter.Count[e.Id],
				Incr:         incr,
				IsAccumulate: cfg.IsAccumulate,
			})
		} else {
			// 不是累计任务，直接转发
			ctx.PublishEventLocal(&event.TargetUpdate{
				Typ:          e.Typ,
				Id:           e.Id,
				Count:        e.Count,
				Incr:         e.Count,
				IsAccumulate: cfg.IsAccumulate,
				IsReplace:    e.Replace,
			})
		}
	} else {
		// 其他玩家的计数
		role, err := s.GetRoleByRoleId(ctx, e.RoleId)
		if err != nil {
			return err
		}
		user, err := s.GetUserById(ctx, role.UserId)
		if err != nil {
			return err
		}
		ctx.PublishEventRemote(role.RoleId, user.ServerId, role.UserId, &lesssvcpb.User_UpdateTargetPush{
			Typ:     values.Integer(e.Typ),
			Id:      e.Id,
			Count:   e.Count,
			Replace: e.Replace,
		})
	}
	return nil
}

// HandleTargetUpdate 回调所有注册的handler
func (s *Service) HandleTargetUpdate(ctx *ctx.Context, d *event.TargetUpdate) *errmsg.ErrMsg {
	ht, ok := s.register.handlers[d.Typ]
	if !ok {
		return nil
	}
	hk, ok := ht[d.Id]
	if !ok {
		return nil
	}

	// 对所有数量变更感兴趣的事件
	all, hasAll := hk[event2.AllCount]
	// 对特定数量变更感兴趣的事件
	handlers, has := hk[d.Count]
	if hasAll {
		for _, h := range all {
			err := h.H(ctx, d, h.Args)
			if err != nil {
				return err
			}
		}
	}
	if has {
		for _, h := range handlers {
			err := h.H(ctx, d, h.Args)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
