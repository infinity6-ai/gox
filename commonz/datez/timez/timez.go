package timez

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/infinity6-ai/gox/commonz/datez"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

var (
	SimpleMonth               = NewConverter("SimpleMonth", `^[0-9]{4}-[0-9]{2}$`, NewConverterFuncsFromPattern("2006-01"))
	SimpleDay                 = NewConverter("SimpleDay", `^[0-9]{4}-[0-9]{2}-[0-9]{2}$`, NewConverterFuncsFromPattern("2006-01-02"))
	ISO8601Millis             = NewConverter("ISO8601Millis", `^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$`, NewConverterFuncsFromPattern("2006-01-02T15:04:05.000Z"))
	ISO8601Seconds            = NewConverter("ISO8601Seconds", `^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$`, NewConverterFuncsFromPattern("2006-01-02T15:04:05Z"))
	ISO8601BasicMillis        = NewConverter("ISO8601BasicMillis", `^[0-9]{8}T[0-9]{9}Z$`, NewConverterFuncs(parserISO8601BasicMillis, formatterISO8601BasicMillis))
	ISO8601BasicMillisWithDot = NewConverter("ISO8601BasicMillisWithDot", `^[0-9]{8}T[0-9]{6}\.[0-9]{3}Z$`, NewConverterFuncsFromPattern("20060102T150405.000Z"))
	ISO8601BasicSeconds       = NewConverter("ISO8601BasicSeconds", `^[0-9]{8}T[0-9]{6}Z$`, NewConverterFuncsFromPattern("20060102T150405Z"))
	ISO8601MillisWithOffset   = NewConverter("ISO8601WithOffset", `^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}[-+][0-9]{2}:[0-9]{2}$`, NewConverterFuncsFromPattern("2006-01-02T15:04:05.000-07:00"))
	ISO8601WithOffset         = NewConverter("ISO8601WithOffset", `^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[-+][0-9]{2}:[0-9]{2}$`, NewConverterFuncsFromPattern("2006-01-02T15:04:05-07:00"))
	EpochMillis               = NewConverter("EpochMillis", `^[0-9]{1,20}$`, NewConverterFuncs(parserEpoch, formatterEpoch))
)

var converters = []*Converter{
	SimpleMonth,
	SimpleDay,
	ISO8601Millis,
	ISO8601Seconds,
	ISO8601BasicMillis,
	ISO8601BasicMillisWithDot,
	ISO8601BasicSeconds,
	ISO8601MillisWithOffset,
	ISO8601WithOffset,
	EpochMillis,
}

var ErrUnsupported = (func() error {
	msg := "ErrorUnsupported, required one of these: "
	for i, converter := range converters {
		msg += fmt.Sprintf("%s[%s]", converter.name, converter.pattern)
		if i < len(converters)-1 {
			msg += ", "
		}
	}
	return errors.New(msg)
})()

func WrapErrUnsupported(original string) error {
	return fmt.Errorf("ErrorUnsupported: %s, %w", original, ErrUnsupported)
}

type Parser func(value string) (time.Time, error)
type Formatter func(value time.Time) string

type Converter struct {
	name      string
	pattern   string
	matcher   *regexp.Regexp
	parser    Parser
	formatter Formatter
}

type ConverterFuncs struct {
	Parser    Parser
	Formatter Formatter
}

func NewConverterFuncs(parser Parser, formatter Formatter) *ConverterFuncs {
	return &ConverterFuncs{
		Parser:    parser,
		Formatter: formatter,
	}
}

func NewConverterFuncsFromPattern(pattern string) *ConverterFuncs {
	return NewConverterFuncs(func(value string) (time.Time, error) {
		t, err := time.Parse(pattern, value)
		if err != nil {
			return t, err
		}
		t = t.UTC()
		return t, nil
	}, func(value time.Time) string {
		return value.UTC().Format(pattern)
	})
}

func NewConverter(name string, matcher string, fns *ConverterFuncs) *Converter {
	return &Converter{
		name:      name,
		pattern:   matcher,
		matcher:   regexp.MustCompile(matcher),
		parser:    fns.Parser,
		formatter: fns.Formatter,
	}
}

func (me *Converter) Format(t time.Time) string {
	return me.formatter(t)
}

func (me *Converter) Match(value string) bool {
	return me.matcher.MatchString(value)
}

func (me *Converter) Parse(value string) (time.Time, error) {
	if me.Match(value) {
		return me.parser(value)
	}
	var zero time.Time
	return zero, WrapErrUnsupported(value)
}

func parserISO8601BasicMillis(value string) (time.Time, error) {
	valueWithDot := value[:15] + "." + value[15:18] + "Z"
	t, err := time.Parse("20060102T150405.000Z", valueWithDot)
	if err != nil {
		return t, err
	}
	t = t.UTC()
	return t, err
}

func formatterISO8601BasicMillis(value time.Time) string {
	ret := value.UTC().Format("20060102T150405.000Z")
	return strings.ReplaceAll(ret, ".", "")
}

func parserEpoch(value string) (time.Time, error) {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		var zero time.Time
		return zero, err
	}
	ret := datez.TimeParseEpoch(n)
	return ret, nil
}

func formatterEpoch(value time.Time) string {
	ret := datez.EpochFromTime(value)
	return strconv.FormatInt(ret, 10)
}

func Lookup(value string) *Converter {
	for _, converter := range converters {
		if converter.Match(value) {
			return converter
		}
	}
	return nil
}

func Parse(value string) (time.Time, error) {
	c := Lookup(value)
	if c == nil {
		var zero time.Time
		return zero, WrapErrUnsupported(value)
	}
	return c.Parse(value)
}

func MustParse(value string) time.Time {
	ret, err := Parse(value)
	errorz.Check(err)
	return ret
}
