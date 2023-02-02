package values

import (
	"coin-server/common/errmsg"
	protomodels "coin-server/common/proto/models"
	"coin-server/common/values"
)

type AchievementI interface {
	RoleId() values.RoleId
	GetAll() map[values.AchievementId]*AchievementDetail
	GetDetail(typ values.AchievementId) (*AchievementDetail, *errmsg.ErrMsg)
	GetPoint() values.Integer
	AddPoint(i values.Integer)

	CheatClear(id values.AchievementId)
}

type achievement struct {
	roleId  values.RoleId
	point   values.Integer
	details map[values.AchievementId]*AchievementDetail
}

func NewAchievement(id values.RoleId, point values.Integer, details map[values.AchievementId]*AchievementDetail) AchievementI {
	return &achievement{
		roleId:  id,
		point:   point,
		details: details,
	}
}

func (a *achievement) RoleId() values.RoleId {
	return a.roleId
}

func (a *achievement) GetPoint() values.Integer {
	return a.point
}

func (a *achievement) AddPoint(i values.Integer) {
	a.point = a.point + i
	if a.point < 0 {
		a.point = 0
	}
}

func (a *achievement) GetAll() map[values.AchievementId]*AchievementDetail {
	return a.details
}

func (a *achievement) GetDetail(typ values.AchievementId) (*AchievementDetail, *errmsg.ErrMsg) {
	if _, exist := a.details[typ]; !exist {
		return nil, errmsg.NewErrAchievementNotExist()
	}
	return a.details[typ], nil
}

func (a *achievement) SetDetail(typ values.AchievementId, detail *AchievementDetail) *errmsg.ErrMsg {
	if _, exist := a.details[typ]; !exist {
		return errmsg.NewErrAchievementNotExist()
	}
	a.details[typ] = detail
	return nil
}

func (a *achievement) CheatClear(id values.AchievementId) {
	a.details[id] = &AchievementDetail{
		AchievementId: id,
		CurrGear:      1,
		Gears:         GearSlice{},
	}
}

type AchievementDetail struct {
	// 成就类型
	AchievementId values.AchievementId
	// 当前档位
	CurrGear values.Integer
	// 当前领取的档位
	CollectedGear values.Integer
	// 总档位
	TotalGear values.Integer
	// 当前完成数
	CurrCnt values.Integer
	// 上次完成时间
	DoneTime values.Integer
	// 列表
	Gears GearSlice
}

func (ad AchievementDetail) IsFinished() bool {
	return ad.CurrGear == ad.TotalGear && ad.Gears.IsDone(ad.CurrGear)
}

func (ad AchievementDetail) HasUnread() bool {
	if ad.IsFinished() {
		return ad.CurrGear != ad.CollectedGear
	}
	return ad.CurrGear != ad.CollectedGear+1
}

func (ad AchievementDetail) ToProtoView() *protomodels.AchievementView {
	return &protomodels.AchievementView{
		Id:            ad.AchievementId,
		TotalGear:     ad.TotalGear,
		CurrGear:      ad.CurrGear,
		CollectedGear: ad.CollectedGear,
		CurrCnt:       ad.CurrCnt,
		HasUnread:     ad.HasUnread(),
		DoneTime:      ad.DoneTime,
	}
}

type GearSlice []byte

func (gs GearSlice) GetGear(gear values.Integer) byte {
	return gs[gear-1]
}

func (gs GearSlice) IsDone(gear values.Integer) bool {
	if gs[gear-1]&2 == 0 {
		return false
	}
	return true
}

func (gs GearSlice) IsCollected(gear values.Integer) bool {
	if gs[gear-1]&1 == 0 {
		return false
	}
	return true
}

func (gs GearSlice) Done(gear values.Integer) {
	gs[gear-1] = gs[gear-1] | 2
}

func (gs GearSlice) DoneIdx(idx int) {
	gs[idx] = gs[idx] | 2
}

func (gs GearSlice) Collect(gear values.Integer) {
	gs[gear-1] = gs[gear-1] | 1
}

func (gs GearSlice) ResetCollect(idx int) {
	if gs[idx]&1 == 0 {
		return
	}
	gs[idx] -= 1
}
