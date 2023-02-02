package module

import (
	"time"

	"coin-server/common/ctx"
)

type RefreshService interface {
	GetCurrDayFreshTime(c *ctx.Context) time.Time
	GetCurrDayFreshTimeWith(c *ctx.Context, now time.Time) time.Time
	GetActivityCurrDayFreshTime(c *ctx.Context) time.Time
	GetActivityCurrWeekFreshTime(c *ctx.Context) time.Time
}
