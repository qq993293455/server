package dao

import (
	"coin-server/common/ctx"
	"coin-server/common/errmsg"
	"coin-server/common/proto/dao"
	"coin-server/common/redisclient"
	"coin-server/common/values"
	aval "coin-server/game-server/service/achievement/values"
	rule_model "coin-server/rule/rule-model"
)

func Get(ctx *ctx.Context) (data aval.AchievementI, err *errmsg.ErrMsg) {
	res := &dao.Achievement{RoleId: ctx.RoleId}
	ok, err := ctx.NewOrm().GetPB(redisclient.GetDefaultRedis(), res)
	if err != nil {
		return nil, errmsg.NewErrorDB(err)
	}
	var isChange = false
	if !ok {
		res = &dao.Achievement{
			RoleId: ctx.RoleId,
			Status: map[int64]*dao.AchievementStatus{},
		}
	}
	data, isChange = daoToValue(res)
	if isChange {
		ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), valueToDao(data))
	}
	return data, nil
}

func Save(ctx *ctx.Context, i aval.AchievementI) {
	ctx.NewOrm().SetPB(redisclient.GetDefaultRedis(), valueToDao(i))
}

func daoToValue(d *dao.Achievement) (aval.AchievementI, bool) {
	achievementRule := rule_model.GetReader().Achievement.List()
	details := map[values.AchievementId]*aval.AchievementDetail{}
	isChange := false
	for _, r := range achievementRule {
		var currGear values.Integer = 1
		if _, exist := d.Status[r.Id]; !exist {
			d.Status[r.Id] = &dao.AchievementStatus{
				CurrGear: 1,
			}
		}
		if d.Status[r.Id].CurrGear > 1 {
			currGear = d.Status[r.Id].CurrGear
		}
		totalGear := len(rule_model.GetReader().AchievementGear()[r.Id])
		gears := []byte(d.Status[r.Id].Gears)
		if totalGear > len(gears) {
			newGears := make([]byte, totalGear)
			for idx, gear := range gears {
				newGears[idx] = gear
			}
			gears = newGears
			isChange = true
		}
		details[r.Id] = &aval.AchievementDetail{
			CurrGear:      currGear,
			AchievementId: r.Id,
			CollectedGear: d.Status[r.Id].CollectGear,
			DoneTime:      d.Status[r.Id].DoneTime,
			CurrCnt:       d.Status[r.Id].CurrCnt,
			TotalGear:     values.Integer(totalGear),
			Gears:         gears,
		}
	}
	return aval.NewAchievement(d.RoleId, d.Point, details), isChange
}

func valueToDao(val aval.AchievementI) *dao.Achievement {
	d := &dao.Achievement{
		RoleId: val.RoleId(),
		Point:  val.GetPoint(),
	}
	poDetails := map[values.AchievementId]*dao.AchievementStatus{}
	for achievementId, detail := range val.GetAll() {
		var currGear values.Integer = 1
		if detail.CurrGear > 1 {
			currGear = detail.CurrGear
		}
		poDetails[achievementId] = &dao.AchievementStatus{
			CurrGear:    currGear,
			CollectGear: detail.CollectedGear,
			CurrCnt:     detail.CurrCnt,
			DoneTime:    detail.DoneTime,
			Gears:       string(detail.Gears),
		}
	}
	d.Status = poDetails
	return d
}
