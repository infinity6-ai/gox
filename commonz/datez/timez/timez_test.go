package timez_test

import (
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/datez/timez"
	"github.com/stretchr/testify/require"
)

func checkConverter(t *testing.T, str string, ts time.Time) {
	c := timez.Lookup(str)
	require.NotNil(t, c)

	require.Equal(t, str, c.Format(ts))

	parsedTime, err := c.Parse(str)
	require.NoError(t, err)
	require.Equal(t, ts, parsedTime)

	parsedTime, err = timez.Parse(str)
	require.NoError(t, err)
	require.Equal(t, ts, parsedTime)
}

func TestUnitTimezParseSimpleMonth(t *testing.T) {
	checkConverter(t, "2024-12", time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC))
	checkConverter(t, "2004-11", time.Date(2004, 11, 1, 0, 0, 0, 0, time.UTC))
}

func TestUnitTimezParseSimpleDay(t *testing.T) {
	checkConverter(t, "2024-12-31", time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC))
	checkConverter(t, "2004-11-04", time.Date(2004, 11, 4, 0, 0, 0, 0, time.UTC))
}

func TestUnitTimezParseISO8601Millis(t *testing.T) {
	checkConverter(t, "2024-01-01T12:30:05.123Z", time.Date(2024, 1, 1, 12, 30, 5, 123000000, time.UTC))
}

func TestUnitTimezParseISO8601Seconds(t *testing.T) {
	checkConverter(t, "2024-01-01T12:30:05Z", time.Date(2024, 1, 1, 12, 30, 5, 0, time.UTC))
}

func TestUnitTimezParseISO8601BasicMillis(t *testing.T) {
	checkConverter(t, "20240101T123005123Z", time.Date(2024, 1, 1, 12, 30, 5, 123000000, time.UTC))
}

func TestUnitTimezParseISO8601BasicMillisWithDot(t *testing.T) {
	checkConverter(t, "20240101T123005.123Z", time.Date(2024, 1, 1, 12, 30, 5, 123000000, time.UTC))
}

func TestUnitTimezParseISO8601BasicSeconds(t *testing.T) {
	checkConverter(t, "20240101T123005Z", time.Date(2024, 1, 1, 12, 30, 5, 0, time.UTC))
}

func TestUnitTimezParseISO8601WithOffset(t *testing.T) {
	c := timez.Lookup("2026-05-26T09:55:36+00:00")
	require.NotNil(t, c)

	require.Equal(t, "2026-05-26T13:55:36+00:00", c.Format(time.Date(2026, 5, 26, 13, 55, 36, 0, time.UTC)))
	require.Equal(t, "2026-05-26T13:55:36+00:00", c.Format(time.Date(2026, 5, 26, 9, 55, 36, 0, time.FixedZone("-0400", -4*60*60))))

	require.Equal(t, time.Date(2026, 5, 26, 13, 55, 36, 0, time.UTC), timez.MustParse("2026-05-26T13:55:36-00:00"))
	require.Equal(t, time.Date(2026, 5, 26, 13, 55, 36, 0, time.UTC), timez.MustParse("2026-05-26T09:55:36-04:00"))
	require.Equal(t, time.Date(2026, 5, 26, 13, 55, 36, 0, time.UTC), timez.MustParse("2026-05-26T13:55:36+00:00"))
	// require.Equal(t, time.Date(2026, 5, 26, 13, 55, 36, 0, time.UTC), timez.MustParse("2023-07-01T00:00:00.000-03:00"))

}

func TestUnitTimezParseISO8601MillisWithOffset(t *testing.T) {
	c := timez.Lookup("2026-05-26T09:55:36.345+00:00")
	require.NotNil(t, c)

	require.Equal(t, "2026-05-26T13:55:36.345+00:00", c.Format(time.Date(2026, 5, 26, 13, 55, 36, 345000000, time.UTC)))
	require.Equal(t, "2026-05-26T13:55:36.345+00:00", c.Format(time.Date(2026, 5, 26, 9, 55, 36, 345000000, time.FixedZone("-0400", -4*60*60))))

	require.Equal(t, time.Date(2026, 5, 26, 13, 55, 36, 345000000, time.UTC), timez.MustParse("2026-05-26T13:55:36.345-00:00"))
	require.Equal(t, time.Date(2026, 5, 26, 13, 55, 36, 345000000, time.UTC), timez.MustParse("2026-05-26T09:55:36.345-04:00"))
	require.Equal(t, time.Date(2026, 5, 26, 13, 55, 36, 345000000, time.UTC), timez.MustParse("2026-05-26T13:55:36.345+00:00"))
	require.Equal(t, time.Date(2023, 7, 1, 3, 0, 0, 0, time.UTC), timez.MustParse("2023-07-01T00:00:00.000-03:00"))

}

func TestUnitTimezParseUnsupportedFormat(t *testing.T) {
	input := "2024-01-01 12:30:05"
	parsedTime, err := timez.Parse(input)
	require.ErrorIs(t, err, timez.ErrUnsupported)
	require.Zero(t, parsedTime)
}

func TestUnitTimezParseMalformedISO8601Millis(t *testing.T) {
	input := "2024-01-01T12:30:05.12Z" // two digits for millis
	parsedTime, err := timez.Parse(input)
	require.ErrorIs(t, err, timez.ErrUnsupported)
	require.Zero(t, parsedTime)
}

func TestUnitTimezParseEpoch(t *testing.T) {
	checkConverter(t, "1704112205123", time.Date(2024, 1, 1, 12, 30, 5, 123000000, time.UTC))
}

func TestUnitTimezUnsupported(t *testing.T) {
	input := "123456789012345678901"
	parsedTime, err := timez.Parse(input)
	require.ErrorIs(t, err, timez.ErrUnsupported)
	require.Zero(t, parsedTime)
}
