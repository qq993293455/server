package timer

import (
	"sync"
	"time"

	"coin-server/common/ctx"
	"coin-server/common/logger"

	avl "github.com/emirpasic/gods/trees/avltree"
	"github.com/emirpasic/gods/utils"
)

func init() {
	Timer = &timer{}
	Timer.run()
	sl := avl.NewWith(utils.Int64Comparator)
	ticker_ = &ticker{
		sl:    sl,
		Mutex: sync.Mutex{},
	}
	ticker_.run()
}

var (
	Timer   *timer
	ticker_ *ticker
)

// 过多久后执行
func AfterFunc(duration time.Duration, f func()) {
	ticker_.add(Now().Add(duration), f)
}

func AfterFuncWithCtx(c *ctx.Context, duration time.Duration, f func(*ctx.Context)) {
	uc := ctx.GetPoolContext()
	uc.ServerHeader = c.ServerHeader
	uc.TraceLogger = logger.GetTraceLoggerWith(c.TraceId, c.RoleId)
	ticker_.add(Now().Add(duration), func() {
		defer uc.Release()
		f(uc)
	})
}

// 到时间点执行
func UntilFunc(t time.Time, f func()) {
	ticker_.add(t, f)
}

func UntilFuncWithCtx(c *ctx.Context, t time.Time, f func(*ctx.Context)) {
	uc := ctx.GetPoolContext()
	uc.ServerHeader = c.ServerHeader
	uc.TraceLogger.ResetInitFiledS(c.TraceId, c.RoleId)
	ticker_.add(t, func() {
		defer uc.Release()
		f(uc)
	})
}

// 间隔时间循环执行
func Ticker(duration time.Duration, f func() bool) {
	ticker_.addTicker(Now().Add(duration), duration, f)
}

func TickerWithCtx(c *ctx.Context, duration time.Duration, f func(*ctx.Context) bool) {
	uc := ctx.GetPoolContext()
	uc.ServerHeader = c.ServerHeader
	uc.TraceLogger.ResetInitFiledS(c.TraceId, c.RoleId)
	ticker_.addTicker(Now().Add(duration), duration, func() (loop bool) {
		defer func() {
			if !loop {
				uc.Release()
			}
		}()
		return f(uc)
	})
}

func DayPass(begin time.Time, end time.Time) int64 {
	b := BeginOfDay(begin)
	e := BeginOfDay(end)
	d := int64(time.Hour * 24)
	return (e.UnixNano() - b.UnixNano()) / d
}

// BeginOfDay 返回一天的开始时刻
func BeginOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// EndOfDay 返回一天的结束时刻
func EndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, 999999999, t.Location())
}

// NextDay 返回下一天的同一时刻
func NextDay(t time.Time) time.Time {
	d := time.Hour * 24
	return t.Add(d)
}

// LastDay  返回前一天的同一时刻
func LastDay(t time.Time) time.Time {
	d := time.Hour * -24
	return t.Add(d)
}

// BeginOfWeek 返回一周的开始时刻,默认为周日0点
func BeginOfWeek(t time.Time) time.Time {
	t = BeginOfDay(t)
	wd := int(t.Weekday())

	weekStartDay := int(time.Monday)
	if wd < weekStartDay {
		wd = wd + 7 - weekStartDay
	} else {
		wd = wd - weekStartDay
	}

	return t.AddDate(0, 0, -wd)
}

// EndOfWeek 返回一周的结束时刻
func EndOfWeek(t time.Time) time.Time {
	wd := t.Weekday()
	day := 7 - wd
	duration := time.Hour * time.Duration(day) * 24
	w := EndOfDay(t)
	w = w.Add(duration)
	return w
}

// Now 当前时间,误差 1 milli
func Now() time.Time {
	return Timer.Now()
}

func StartTime(start int64) time.Time {
	return time.Unix(0, start).Add(Timer.offset * time.Second)
}

// Unix 当前秒级时间戳,误差 1 milli
func Unix() int64 {
	return Timer.Unix()
}

// UnixMilli 当前毫秒级时间戳,误差 1 milli
func UnixMilli() int64 {
	return Timer.UnixMilli()
}

// UnixNano 当前纳秒级时间戳,误差 1 milli
func UnixNano() int64 {
	return Timer.UnixNano()
}

// NowToString 当前时间转换成 "2006-01-02 15:04:05" 格式
func NowToString() string {
	return Timer.ToString()
}

// NowString 当前时间转换成 "2006-01-02 15:04:05.999999999 -0700 MST" 格式
func NowString() string {
	return Timer.String()
}

func GetRefreshTime(lastTime time.Time, refreshHour int, refreshMin int) (time.Time, time.Time) {
	y, m, d := lastTime.UTC().Date()
	todayRefreshTime := time.Date(y, m, d, refreshHour, refreshMin, 0, 0, lastTime.UTC().Location())
	tomorrowRefreshTime := time.Date(y, m, d+1, refreshHour, refreshMin, 0, 0, lastTime.UTC().Location())
	return todayRefreshTime, tomorrowRefreshTime
}

func GetNextRefreshTime(refreshHour int, refreshMin int) int64 {
	localNow := Now().UTC()
	todayRefreshTime, tomorrowRefreshTime := GetRefreshTime(localNow, refreshHour, refreshMin)
	if localNow.UTC().Unix() < todayRefreshTime.UTC().Unix() {
		return todayRefreshTime.UTC().Unix()
	}
	return tomorrowRefreshTime.UTC().Unix()
}

// OverRefreshTime 判断是否超过刷新时间
func OverRefreshTime(lastRefreshTimeSecond int64, refreshHour int, refreshMin int) bool {
	lastTime := time.Unix(lastRefreshTimeSecond, 0).UTC()
	todayRefreshTime, tomorrowRefreshTime := GetRefreshTime(lastTime, refreshHour, refreshMin)
	//fmt.Println(lastTime.UTC().Unix(), todayRefreshTime.UTC().Unix(), tomorrowRefreshTime.UTC().Unix())
	if lastTime.UTC().Unix() < todayRefreshTime.UTC().Unix() {
		return Now().Unix() > todayRefreshTime.UTC().Unix()
	} else {
		return Now().Unix() > tomorrowRefreshTime.UTC().Unix()
	}
}
