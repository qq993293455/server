package db

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/proto/models"
	"coin-server/common/redisclient"
	"coin-server/rule"
)

func GetBattleSetting(c *ctx.Context) (*dao.BattleSetting, *errmsg.ErrMsg) {
	u := &dao.BattleSetting{RoleId: c.RoleId}
	ok, err := c.NewOrm().GetPB(redisclient.GetUserRedis(), u)
	if err != nil {
		return nil, err
	}
	if !ok {
		num, has := rule.MustGetReader(c).KeyValue.GetInt64("ScreenPlayerRecommendNum")
		if !has {
			return nil, errmsg.NewInternalErr("KV not exist")
		}
		hp, has := rule.MustGetReader(c).KeyValue.GetInt64("TakeMedicineHPPercentage")
		if !has {
			return nil, errmsg.NewInternalErr("KV not exist")
		}
		mp, has := rule.MustGetReader(c).KeyValue.GetInt64("TakeMedicineMPPercentage")
		if !has {
			return nil, errmsg.NewInternalErr("KV not exist")
		}
		u.Data = &models.BattleSettingData{
			ScreenCurCount:          num,
			IsShowOtherPlayer:       true,
			IsShowOtherPlayerEffect: true,
			IsShowFightText:         true,
			Hp:                      hp,
			Mp:                      mp,
		}
		c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
		return u, nil
	}
	return u, nil
}

func SaveBattleSetting(c *ctx.Context, u *dao.BattleSetting) {
	c.NewOrm().SetPB(redisclient.GetUserRedis(), u)
}
