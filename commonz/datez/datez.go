package datez

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

var http_pattern *regexp.Regexp

func init() {
	var err error
	http_pattern, err = regexp.Compile(`^[0-9]{4}\-[0-9]{2}\-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$`)
	errorz.Check(err)
}

func TimeFormatTZ(t time.Time) string {
	tstr := t.UTC().Format("20060102T150405.000Z")
	return strings.Replace(tstr, ".", "", 1)
}

func TimeParseTZ(tz string) time.Time {
	checker.Equal(19, len(tz), "wrong tz")
	tzStr := tz[:15] + "." + tz[15:]
	t, err := time.Parse("20060102T150405.000Z", tzStr)
	errorz.Check(err)
	return t.UTC()
}

func TimeParseHttp(ts string) time.Time {
	checker.RegexMatch(http_pattern, ts, "ts")
	t, err := time.Parse("2006-01-02T15:04:05.000Z", ts)
	errorz.Check(err)
	return t.UTC()
}

func TimeFormatHttp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func NowMilliInt() int64 {
	return time.Now().UTC().UnixMilli()
}

func NowMilli() string {
	return strconv.FormatInt(NowMilliInt(), 10)
}

func NowTZ() string {
	return TimeFormatTZ(time.Now().UTC())
}

func NowHttp() string {
	return TimeFormatHttp(time.Now().UTC())
}

func NowTime() time.Time {
	return time.Now().UTC()
}

func NowTimePointer() *time.Time {
	ret := NowTime()
	return &ret
}

func CurrentMonth() time.Time {
	date := NowTime()
	date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	return date
}

func TimeParseEpoch(epochMillis int64) time.Time {
	return time.UnixMilli(epochMillis).UTC()
}

func EpochFormatTZ(epochMillis int64) string {
	return TimeFormatTZ(TimeParseEpoch(epochMillis))
}

func EpochParseTZ(tz string) int64 {
	return TimeParseTZ(tz).UnixMilli()
}

func EpochStringParseTZ(tz string) string {
	return strconv.FormatInt(EpochParseTZ(tz), 10)
}

func EpochFromTime(ts time.Time) int64 {
	return ts.UTC().UnixMilli()
}

func DiffMonths(a, b time.Time) int {
	a = time.Date(a.Year(), a.Month(), 1, 0, 0, 0, 0, time.UTC)
	b = time.Date(b.Year(), b.Month(), 1, 0, 0, 0, 0, time.UTC)

	reversed := 1
	if a.After(b) {
		a, b = b, a
		reversed = -1
	}
	ret := 0
	for a.Before(b) {
		a = a.AddDate(0, 1, 0)
		ret += 1
	}
	return ret * reversed
}

func DiffMonthsEpoch(a, b int64) int {
	parsedA := TimeParseEpoch(a)
	parsedB := TimeParseEpoch(b)

	return DiffMonths(parsedA, parsedB)
}
