package values

import (
	"coin-server/common/proto/dao"
	"coin-server/common/values"
)

type CounterI interface {
	GetRoleId() values.RoleId
	GetCnt(typ values.AchievementId) (res values.Integer)
	Add(typ values.AchievementId, cnt values.Integer)
	Update(typ values.AchievementId, cnt values.Integer)
	ToDao() *dao.AchievementCounter
}

type Counter struct {
	values *dao.AchievementCounter
}

func NewCounter(val *dao.AchievementCounter) CounterI {
	return &Counter{values: val}
}

func (c *Counter) GetRoleId() values.RoleId {
	return c.values.RoleId
}

func (c *Counter) GetCnt(typ values.AchievementId) (res values.Integer) {
	var exist bool
	if res, exist = c.values.Cnt[typ]; !exist {
		return 0
	}
	return res
}

func (c *Counter) Add(typ values.AchievementId, cnt values.Integer) {
	c.values.Cnt[typ] += cnt
}

func (c *Counter) Update(typ values.AchievementId, cnt values.Integer) {
	c.values.Cnt[typ] = cnt
}

func (c *Counter) ToDao() *dao.AchievementCounter {
	return c.values
}
