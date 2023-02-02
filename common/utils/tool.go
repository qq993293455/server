package utils

import (
	"math/rand"
	"strconv"
	"time"

	"coin-server/common/utils/generic"
)

func MaxNumber[T generic.Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func MinNumber[T generic.Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}

func MaxFloat64(a, b float64) float64 {
	if a >= b {
		return a
	}
	return b
}

func MinFloat64(a, b float64) float64 {
	if a <= b {
		return a
	}
	return b
}

func Rand100Per() float64 {
	return float64(rand.Intn(101)) / 100
}

func GetCurrWeek() int64 {
	now := time.Now().UTC()
	return GetWeekWithTime(now)
}
func GetWeekWithTime(now time.Time) int64 {
	year, week := now.ISOWeek()
	str := strconv.Itoa(year) + strconv.Itoa(week)
	num, err := strconv.Atoi(str)
	Must(err)
	return int64(num)
}
