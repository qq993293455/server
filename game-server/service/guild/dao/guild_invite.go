package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
)

func GetInvite(ctx *ctx.Context, guildId values.GuildId, refresh bool, lang, combatValueLimit values.Integer) ([]values.RoleId, *errmsg.ErrMsg) {
	out := &dao.GuildInvite{GuildId: guildId}
	if refresh {
		list, err := getFromMySQL(lang, combatValueLimit)
		if err != nil {
			return nil, err
		}
		out.RoleId = list
		saveInvite(ctx, out)
		return list, nil
	}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetGuildRedis(), out)
	if err != nil {
		return nil, err
	}
	if !ok {
		list, err := getFromMySQL(lang, combatValueLimit)
		if err != nil {
			return nil, err
		}
		out.RoleId = list
		saveInvite(ctx, out)
		return list, nil
	}
	return out.RoleId, nil
}

func saveInvite(ctx *ctx.Context, invite *dao.GuildInvite) {
	ctx.NewOrm().SetPB(redisclient.GetGuildRedis(), invite)
}

func getFromMySQL(lang, combatValueLimit values.Integer) ([]string, *errmsg.ErrMsg) {
	query := "SELECT role_id FROM `roles`"
	args := make([]interface{}, 0)
	var where string
	if lang != 1 {
		where += " language=?"
		args = append(args, lang)
	}
	if combatValueLimit > 0 {
		where += " AND power>=?"
		args = append(args, combatValueLimit)
	}
	if where != "" {
		query += " WHERE " + where
	}
	query += " LIMIT 100"
	data := make([]string, 0)
	err := orm.GetMySQL().Select(&data, query, args...)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	return data, nil
}
