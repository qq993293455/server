package dao

import (
	"fmt"
	"strconv"
	"strings"

	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/orm"
	daopb "coin-server/common/proto/dao"
	"coin-server/common/redisclient"
)

func GetRewardsUserInfo(c *ctx.Context, roleId string) (*daopb.GuildBossUserInfo, *errmsg.ErrMsg) {
	res := &daopb.GuildBossUserInfo{RoleId: roleId}
	_, err := c.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, err
	}
	if res.AllRewards == nil {
		res.AllRewards = map[string]int64{}
	}
	return res, nil
}

func SaveRewardsUserInfo(c *ctx.Context, gb *daopb.GuildBossUserInfo) {
	c.NewOrm().SetPB(redisclient.GetDefaultRedis(), gb)
}

func GetGuildBossUserFightInfo(c *ctx.Context, roleId string) (*daopb.GuildBossUserFightInfo, *errmsg.ErrMsg) {
	res := &daopb.GuildBossUserFightInfo{RoleId: roleId}
	_, err := c.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetManyGuildBossUserFightInfo(c *ctx.Context, roleIds []string) ([]*daopb.GuildBossUserFightInfo, *errmsg.ErrMsg) {
	s := make([]*daopb.GuildBossUserFightInfo, 0, len(roleIds))
	out := make([]orm.RedisInterface, 0, len(roleIds))
	for _, v := range roleIds {
		gbfi := &daopb.GuildBossUserFightInfo{RoleId: v}
		s = append(s, gbfi)
		out = append(out, gbfi)
	}
	_, err := c.NewOrm().MGetPB(redisclient.GetDefaultRedis(), out...)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func SaveManyGuildBossUserFightInfo(c *ctx.Context, dataS []*daopb.GuildBossUserFightInfo) {
	in := make([]orm.RedisInterface, 0, len(dataS))
	for _, v := range dataS {
		in = append(in, v)
	}
	c.NewOrm().MSetPB(redisclient.GetDefaultRedis(), in)
}

func SaveGuildBossUserFightInfo(c *ctx.Context, gb *daopb.GuildBossUserFightInfo) {
	c.NewOrm().SetPB(redisclient.GetDefaultRedis(), gb)
}

func GetGuildBoss(c *ctx.Context, guildDayId string) (*daopb.GuildBoss, *errmsg.ErrMsg) {
	res := &daopb.GuildBoss{GuildDayId: guildDayId}
	_, err := c.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, err
	}
	if res.Damages == nil {
		res.Damages = map[string]*daopb.GuildBossDamageAndFightCount{}
	}
	if res.Day == 0 {
		strS := strings.Split(guildDayId, ":")
		if len(strS) != 2 {
			panic(fmt.Sprintf("guildDayId:%s invalid", guildDayId))
		}
		day, err := strconv.Atoi(strS[1])
		if err != nil {
			panic(err)
		}
		res.Day = int64(day)
	}
	return res, nil
}

func SaveGuildBoss(c *ctx.Context, gb *daopb.GuildBoss) {
	c.NewOrm().SetPB(redisclient.GetDefaultRedis(), gb)
}

func GetManyGuildBoss(c *ctx.Context, guildDayIds []string) (map[string]*daopb.GuildBoss, *errmsg.ErrMsg) {
	m := make(map[string]*daopb.GuildBoss, len(guildDayIds))
	if len(guildDayIds) == 0 {
		return m, nil
	}
	for _, v := range guildDayIds {
		m[v] = &daopb.GuildBoss{GuildDayId: v}
	}
	out := make([]orm.RedisInterface, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	nfi, err := c.NewOrm().MGetPB(redisclient.GetDefaultRedis(), out...)
	if err != nil {
		return nil, err
	}
	if len(nfi) > 0 {
		for _, v := range nfi {
			delete(m, out[v].PK())
		}
	}
	return m, nil
}

func GetGuildBossList(c *ctx.Context, guildId string) (*daopb.GuildBossList, *errmsg.ErrMsg) {
	res := &daopb.GuildBossList{GuildId: guildId}
	_, err := c.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func SaveGuildBossList(c *ctx.Context, gbl *daopb.GuildBossList) {
	c.NewOrm().SetPB(redisclient.GetDefaultRedis(), gbl)
}

func DeleteGBS(c *ctx.Context, guildDayIds []string) {
	if len(guildDayIds) == 0 {
		return
	}
	deleteIds := make([]orm.RedisInterface, 0, len(guildDayIds))
	for _, v := range guildDayIds {
		deleteIds = append(deleteIds, &daopb.GuildBoss{GuildDayId: v})
	}
	c.NewOrm().DelManyPB(redisclient.GetDefaultRedis(), deleteIds...)
}
