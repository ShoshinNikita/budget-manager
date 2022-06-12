package logger

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/caller"
)

// {year}-{month}-{day} {hour}:{minute}:{second}
const timeLayout = "2006-01-02 15:04:05"

// devFormatter is used as a log formatter in the developer mode
type devFormatter struct{}

var _ logrus.Formatter = devFormatter{}

// Format formats '*logrus.Entry'
func (devFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buff := &bytes.Buffer{}
	if entry.Buffer != nil {
		buff = entry.Buffer
	}

	// Time
	buff.WriteString(color.HiGreenString(entry.Time.Format(timeLayout)))
	buff.WriteByte(' ')

	// Level
	buff.Write(logLevelToString(entry.Level))
	buff.WriteByte(' ')

	// Message
	buff.WriteString(entry.Message)
	buff.WriteByte(' ')

	// Fields

	// Add the caller to fields
	if entry.HasCaller() {
		entry.Data["func"] = caller.FormatCaller(entry.Caller.Func)
	}

	// Sort field keys
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write
	coloredPrintf := logLevelToPrintfFunction(entry.Level)
	for i, k := range keys {
		v := entry.Data[k]
		if isEmptyString(v) {
			v = `""`
		}

		buff.WriteString(coloredPrintf(k))
		buff.WriteByte('=')
		fmt.Fprint(buff, v)
		if i+1 != len(keys) {
			buff.WriteByte(' ')
		}
	}

	buff.WriteByte('\n')
	return buff.Bytes(), nil
}

// Ready to print colored log levels
//
//nolint:gochecknoglobals
var (
	coloredTraceLvl = []byte(color.HiMagentaString("[TRC]"))
	coloredDebugLvl = []byte(color.HiMagentaString("[DBG]"))
	coloredInfoLvl  = []byte(color.CyanString("[INF]"))
	coloredWarnLvl  = []byte(color.YellowString("[WRN]"))
	coloredErrLvl   = []byte(color.RedString("[ERR]"))
	coloredPanicLvl = []byte(color.New(color.BgRed).Sprint("[PNC]"))
	coloredFatalLvl = []byte(color.New(color.BgRed).Sprint("[FAT]"))
)

// logLevelToString converts logrus log level to colored string
// It returns bytes for greater convenience
func logLevelToString(lvl logrus.Level) []byte {
	switch lvl {
	case logrus.TraceLevel:
		return coloredTraceLvl
	case logrus.DebugLevel:
		return coloredDebugLvl
	case logrus.InfoLevel:
		return coloredInfoLvl
	case logrus.WarnLevel:
		return coloredWarnLvl
	case logrus.ErrorLevel:
		return coloredErrLvl
	case logrus.PanicLevel:
		return coloredPanicLvl
	case logrus.FatalLevel:
		return coloredFatalLvl
	default:
		return []byte("[...]")
	}
}

// logLevelToPrintfFunction returns a function to print colored output according to
// passed logrus log level
func logLevelToPrintfFunction(lvl logrus.Level) func(format string, a ...interface{}) string {
	switch lvl {
	case logrus.TraceLevel:
		return color.YellowString
	case logrus.DebugLevel:
		return color.HiMagentaString
	case logrus.InfoLevel:
		return color.CyanString
	case logrus.WarnLevel:
		return color.YellowString
	case logrus.ErrorLevel:
		return color.RedString
	case logrus.FatalLevel:
		return color.New(color.BgRed).Sprintf
	case logrus.PanicLevel:
		return color.New(color.BgRed).Sprintf
	default:
		return color.New().Sprintf
	}
}

func isEmptyString(i interface{}) bool {
	v := reflect.ValueOf(i)
	return v.Kind() == reflect.String && v.String() == ""
}
