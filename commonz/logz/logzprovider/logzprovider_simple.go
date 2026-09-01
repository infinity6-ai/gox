package logzprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/infinity6-ai/gox/commonz/logz/logzcolor"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"golang.org/x/term"
)

var I6_LOGZ_NO_COLOR = false
var mlogger *log.Logger

func initColor() {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		I6_LOGZ_NO_COLOR = true
	}
}

func init() {
	initColor()
	mlogger = log.New(os.Stderr, "", 0)
}

func SimpleProvider(ctx context.Context, entry *logzspec.Entry) {
	paramsFormatted := applyMap(logzcolor.BLUE, logzcolor.CYAN, entry.Params)
	// logContext := applyMap(logzcolor.YELLOW, logzcolor.CYAN, entry.Context)
	appender := entry.Appender

	levelColor := getLevelColor(entry.Level)
	// showStack := entry.Level == logzspec.ERROR
	showStack := entry.Error != ""

	msg := apply(logzcolor.BLUE, "%s", EpochFormatTZ(entry.CreatedAt))
	msg += " " + apply(logzcolor.BRIGHT_MAGENTA, "%s", appender)
	msg += " " + apply(levelColor, "%s", string(entry.Level))
	msg += " " + apply(logzcolor.CYAN, "%s", string(entry.Id))
	msg += " " + apply(logzcolor.BRIGHT_BLACK, "[%s]", entry.Origin)
	msg += " " + apply(logzcolor.WHITE, "%s", entry.Operation)
	msg += " " + paramsFormatted

	// msg += applyAudit(logzcolor.BOLD_BRIGHT_BLUE, logzcolor.CYAN, entry.Audit)

	// if len(entry.Context) > 0 {
	// 	msg += ", " + apply(logzcolor.BRIGHT_BLACK, "ctx:")
	// 	msg += " " + logContext
	// }

	if entry.Error != "" {
		errorFormatted := Format(entry.Error)
		msg += ", " + apply(logzcolor.BRIGHT_BLACK, "error:")
		msg += " " + apply(logzcolor.BOLD_BRIGHT_RED, "%s", errorFormatted)
	}
	if showStack {
		stackFormatted := Format(entry.Stack)
		msg += ", " + apply(logzcolor.BRIGHT_BLACK, "stack:")
		msg += " " + apply(logzcolor.BRIGHT_RED, "%s", stackFormatted)
	}
	mlogger.Printf("%s", msg)

	if showStack {
		mlogger.Printf("PRINT ERROR STACK: %s", entry.Stack)
	}
}

func apply(color logzcolor.Color, format string, a ...any) string {
	ret := fmt.Sprintf(format, a...)
	if I6_LOGZ_NO_COLOR {
		return ret
	}
	return color.Apply(ret)
}

// func applyAudit(ck logzcolor.Color, cv logzcolor.Color, audit *logzauditdata.Audit) string {
// 	if audit == nil {
// 		return ""
// 	}
// 	if audit.ReqResp == nil {
// 		return ""
// 	}
// 	ret := " [" + apply(ck, "%s", audit.ReqResp.Req.Method)
// 	ret += " " + apply(cv, "%s", audit.ReqResp.Req.Path)
// 	if len(audit.ReqResp.Req.Query) > 0 {
// 		ret += "?" + applyQuery(ck, cv, audit.ReqResp.Req.Query)
// 	}
// 	if audit.ReqResp.Resp != nil {
// 		ret += " " + apply(ck, "%d", audit.ReqResp.Resp.Status)
// 	}
// 	ret += "]"
// 	return ret
// }

// func applyQuery(ck logzcolor.Color, cv logzcolor.Color, m map[string][]string) string {
// 	ret := ""
// 	maxv := 15
// 	first := true

// 	for _, k := range mapz.SortedKeys(m) {
// 		varray := m[k]
// 		for _, v := range varray {
// 			v = util.Ascii([]byte(v), maxv+1)
// 			if len(v) > maxv {
// 				v += "..."
// 			}
// 			if !first {
// 				ret += "&"
// 			}
// 			first = false
// 			ret += apply(ck, "%s", k)
// 			ret += "=" + apply(cv, "%s", v)
// 		}
// 	}
// 	return ret
// }

func applyMap(ck logzcolor.Color, cv logzcolor.Color, m map[string]string) string {
	ret := "["

	for i, k := range slices.Sorted(maps.Keys(m)) {
		v := m[k]
		k = apply(ck, "%s", k)
		v = apply(cv, "%s", v)
		ret += fmt.Sprintf("%s: %s", k, v)
		if i < len(m)-1 {
			ret += ", "
		}
	}
	ret += "]"
	return ret
}

func Format(v any) []byte {
	ret, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return ret
}

func getLevelColor(level logzspec.Level) logzcolor.Color {
	switch level {
	case logzspec.DEBUG:
		return logzcolor.BOLD_YELLOW
	case logzspec.ERROR:
		return logzcolor.BOLD_RED
	default:
		return logzcolor.BOLD_BLUE
	}

}

func EpochFormatTZ(epochMillis int64) string {
	t := time.UnixMilli(epochMillis).UTC()
	tstr := t.Format("2006-01-02 15:04:05.000Z")
	return tstr
}
