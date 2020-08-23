package logger

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/caller"
)

// {year}-{month}-{day} {hour}:{minute}:{second}
const timeLayout = "2006-01-02 15:04:05"

type Config struct {
	Debug bool `env:"DEBUG" envDefault:"false"`

	// Mode is a mode of Logger. Valid options: prod, production, dev, develop.
	// Default value is prod
	Mode string `env:"LOGGER_MODE" envDefault:"prod"`

	// Level is a level of logger. Valid options: debug, info, warn, error, fatal.
	// It is always debug, when debug mode is on
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}

func New(cnf Config) *logrus.Logger {
	log := logrus.New()
	log.SetReportCaller(true)

	// Set passed log level
	log.SetLevel(logLevelFromString(cnf.Level))

	switch cnf.Mode {
	case "prod", "production":
		log.SetFormatter(&logrus.JSONFormatter{
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				// Skip file field
				return caller.FormatCaller(frame.Func), ""
			},
		})
	case "dev", "develop":
		fallthrough
	default:
		log.SetFormatter(devFormatter{})
	}

	// Always use debug level in debug mode
	if cnf.Debug {
		log.SetLevel(logrus.DebugLevel)
	}

	return log
}

// logLevelFromString converts passed string to logrus log level
func logLevelFromString(lvl string) logrus.Level {
	switch lvl {
	case "dbg", "debug":
		return logrus.DebugLevel
	case "inf", "info":
		return logrus.InfoLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "err", "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	default:
		return logrus.InfoLevel
	}
}

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
