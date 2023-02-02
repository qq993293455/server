package guild

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	pbdao "coin-server/common/proto/dao"
	"coin-server/common/values"
	"coin-server/common/values/enum"
	"coin-server/game-server/event"
	"coin-server/game-server/service/guild/dao"

	"github.com/jinzhu/now"
)

func (svc *Service) HandlerUserDailyActiveUpdate(ctx *ctx.Context, d *event.ItemUpdate) *errmsg.ErrMsg {
	var count, incr values.Integer

	for i, item := range d.Items {
		if item.ItemId == enum.DailyTaskActive {
			count = item.Count
			incr = d.Incr[i]
			break
		}
	}
	if count <= 0 {
		return nil
	}
	user, err := dao.NewGuildUser(ctx.RoleId).Get(ctx)
	if err != nil {
		return err
	}
	time := now.BeginningOfDay().Unix()
	// 有公会，更新member，否则更新user
	if user.GuildId != "" {
		if err := svc.getLock(ctx, guildMemberLock+user.GuildId); err != nil {
			return err
		}
		guild, err := dao.NewGuild(user.GuildId).Get(ctx)
		if err != nil {
			return err
		}
		members, err := dao.NewGuildMember(user.GuildId).Get(ctx)
		if err != nil {
			return err
		}
		var member *pbdao.GuildMember
		for _, guildMember := range members {
			if guildMember.RoleId == ctx.RoleId {
				member = guildMember
				break
			}
		}
		if member == nil {
			return nil
		}
		if len(member.ActiveValue) == 0 {
			member.ActiveValue = make(map[int64]int64)
		}
		member.ActiveValue[time] = count
		member.TotalActiveValue += incr
		lastKey := svc.get7DaysLastKey(ctx)
		for key := range member.ActiveValue {
			if key < lastKey {
				delete(member.ActiveValue, key)
			}
		}
		if err := dao.NewGuildMember(user.GuildId).SaveOne(ctx, member); err != nil {
			return err
		}
		return svc.updateToGuildFilterServer(ctx, svc.dao2model(ctx, guild, members))
	} else {
		user.ActiveValue[time] = count
		user.TotalActiveValue += incr
		return dao.NewGuildUser(user.RoleId).Save(ctx, user)
	}
}

func (svc *Service) UserCombatValueChange(ctx *ctx.Context, d *event.UserCombatValueChange) *errmsg.ErrMsg {
	user, err := dao.NewGuildUser(d.RoleId).Get(ctx)
	if err != nil {
		return err
	}
	if user.GuildId == "" {
		return nil
	}
	memberDao := dao.NewGuildMember(user.GuildId)
	members, err := memberDao.Get(ctx)
	if err != nil {
		return err
	}
	var member *pbdao.GuildMember
	var combatValue values.Integer
	for _, m := range members {
		if m.RoleId == d.RoleId {
			member = m
		} else {
			combatValue += m.CombatValue
		}
	}
	if member != nil {
		member.CombatValue = d.Value
		if err := memberDao.SaveOne(ctx, member); err != nil {
			return err
		}
	}
	combatValue += d.Value
	return svc.updateGuildRankValue(ctx, user.GuildId, combatValue)
}
