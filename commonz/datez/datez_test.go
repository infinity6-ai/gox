package datez_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/datez"
	"github.com/stretchr/testify/require"
)

func TestUnitTime(t *testing.T) {
	require.Equal(t, "20240817T215307963Z", datez.EpochFormatTZ(int64(1723931587963)))
	require.Equal(t, int64(1723931587963), datez.EpochParseTZ("20240817T215307963Z"))
}

func TestUnitHttp(t *testing.T) {
	require.Equal(t, "2024-08-17T21:53:07.963Z", datez.TimeFormatHttp(datez.TimeParseEpoch(int64(1723931587963))))
	require.Equal(t, int64(1723931587963), datez.EpochFromTime(datez.TimeParseHttp("2024-08-17T21:53:07.963Z")))
}

func TestUnitDiffMonths(t *testing.T) {
	require.Equal(t, 13, datez.DiffMonths(datez.TimeParseTZ("20240817T215307963Z"), datez.TimeParseTZ("20250917T215307963Z")))
	require.Equal(t, -11, datez.DiffMonths(datez.TimeParseTZ("20250717T215307963Z"), datez.TimeParseTZ("20240817T215307963Z")))
	require.Equal(t, 0, datez.DiffMonths(datez.TimeParseTZ("20240817T215307963Z"), datez.TimeParseTZ("20240820T215307963Z")))
	require.Equal(t, 1, datez.DiffMonths(datez.TimeParseTZ("20240830T215307963Z"), datez.TimeParseTZ("20240901T215307963Z")))
}

func TestUnitTimeFormatTZ(t *testing.T) {
	tm := time.Date(2024, 8, 17, 21, 53, 7, 963000000, time.UTC)
	require.Equal(t, "20240817T215307963Z", datez.TimeFormatTZ(tm))
}

func TestUnitTimeParseTZ(t *testing.T) {
	tm := time.Date(2024, 8, 17, 21, 53, 7, 963000000, time.UTC)
	parsedTime := datez.TimeParseTZ("20240817T215307963Z")
	require.Equal(t, tm, parsedTime)
}

func TestUnitTimeParseHttp(t *testing.T) {
	tm := time.Date(2024, 8, 17, 21, 53, 7, 963000000, time.UTC)
	parsedTime := datez.TimeParseHttp("2024-08-17T21:53:07.963Z")
	require.Equal(t, tm, parsedTime)
}

func TestUnitNowMilliInt(t *testing.T) {
	now := time.Now().UTC().UnixMilli()
	require.InDelta(t, now, datez.NowMilliInt(), 100)
}

func TestUnitNowMilli(t *testing.T) {
	now := time.Now().UTC().UnixMilli()
	require.InDelta(t, now, func() int64 {
		val, _ := strconv.ParseInt(datez.NowMilli(), 10, 64)
		return val
	}(), 100)
}

func TestUnitNowTZ(t *testing.T) {
	now := time.Now().UTC()
	nowTZ := datez.NowTZ()
	parsed, _ := time.Parse("20060102T150405.000Z", nowTZ[:15]+"."+nowTZ[15:])
	require.InDelta(t, now.Unix(), parsed.Unix(), 1)
}

func TestUnitNowHttp(t *testing.T) {
	now := time.Now().UTC()
	nowHttp := datez.NowHttp()
	parsed, _ := time.Parse("2006-01-02T15:04:05.000Z", nowHttp)
	require.InDelta(t, now.Unix(), parsed.Unix(), 1)
}

func TestUnitNowTime(t *testing.T) {
	now := time.Now().UTC()
	require.InDelta(t, now.Unix(), datez.NowTime().Unix(), 1)
}

func TestUnitTimeParseEpoch(t *testing.T) {
	tm := time.UnixMilli(1723931587963).UTC()
	require.Equal(t, tm, datez.TimeParseEpoch(1723931587963))
}

func TestUnitEpochStringParseTZ(t *testing.T) {
	epochStr := datez.EpochStringParseTZ("20240817T215307963Z")
	require.Equal(t, "1723931587963", epochStr)
}

func TestUnitDiffMonthsEpoch(t *testing.T) {
	epochA := int64(1723931587963) // Aug 2024
	epochB := int64(1758155587963) // Sep 2025
	require.Equal(t, 13, datez.DiffMonthsEpoch(epochA, epochB))
}
