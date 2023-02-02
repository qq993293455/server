package util

import (
	"time"

	"coin-server/common/timer"
	"coin-server/rule"
)

// DefaultTodayRefreshTime 系统今日默认每日刷新时间（UTC 0点 + 偏移多少秒）
func DefaultTodayRefreshTime() time.Time {
	offset, _ := rule.MustGetReader(nil).KeyValue.GetInt64("DefaultRefreshTime")
	return timer.BeginOfDay(timer.Now()).Add(time.Duration(offset) * time.Second)
}

// DefaultNextRefreshTime 系统下一次默认每日刷新时间（UTC 0点 + 偏移多少秒）
func DefaultNextRefreshTime() time.Time {
	rt := DefaultTodayRefreshTime()
	if rt.After(timer.Now()) {
		return rt
	}
	return rt.AddDate(0, 0, 1)
}

// DefaultCurWeekRefreshTime 系统本周默认刷新时间（每周第一天 UTC 0点 + 偏移多少秒）
func DefaultCurWeekRefreshTime() time.Time {
	offset, _ := rule.MustGetReader(nil).KeyValue.GetInt64("DefaultRefreshTime")
	return timer.BeginOfWeek(timer.Now().Add(-time.Duration(offset) * time.Second)).Add(time.Duration(offset) * time.Second)
}

// DefaultWeeklyRefreshTime 系统每周默认刷新时间（每周第一天 UTC 0点 + 偏移多少秒）
func DefaultWeeklyRefreshTime() time.Time {
	rt := DefaultCurWeekRefreshTime()
	if rt.After(timer.Now()) {
		return rt
	}
	return rt.AddDate(0, 0, 7)
}
