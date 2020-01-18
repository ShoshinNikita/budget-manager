package logger

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

// {year}/{month}/{day} {hour}:{minute}:{second}
const timeLayout = "2006/01/02 15:04:05"

type Config struct {
	// Mode is a mode of Logger. Valid options: prod, production, dev, develop.
	// Default value is prod
	Mode string `env:"LOGGER_MODE" envDefault:"prod"`

	// Level is a level of logger. Valid options: debug, info, warn, error, fatal.
	// It is always debug, when debug mode is on
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}

func New(cnf Config, debug bool) *logrus.Logger {
	log := logrus.New()

	// Set passed log level
	log.SetLevel(logLevelFromString(cnf.Level))

	switch cnf.Mode {
	case "prod", "production":
		log.SetFormatter(&logrus.JSONFormatter{})
	case "dev", "develop":
		fallthrough
	default:
		log.SetFormatter(DevFormatter{})
	}

	// Always use debug level in debug mode
	if debug {
		log.SetLevel(logrus.DebugLevel)
	}

	return log
}

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

type DevFormatter struct{}

var _ logrus.Formatter = DevFormatter{}

// Format formats
func (DevFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	buff := &bytes.Buffer{}
	if entry.Buffer != nil {
		buff = entry.Buffer
	}

	// Write time
	buff.WriteString(color.HiGreenString(entry.Time.Format(timeLayout)))
	buff.WriteByte(' ')

	// Write level
	buff.Write(logLevelToString(entry.Level))
	buff.WriteByte(' ')

	// Write message
	buff.WriteString(entry.Message)
	buff.WriteByte(' ')

	// Sort data keys
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write data
	coloredPrintf := logLevelToPrintfFunction(entry.Level)
	for _, k := range keys {
		v := entry.Data[k]

		buff.WriteString(coloredPrintf(k))
		buff.WriteByte('=')
		fmt.Fprint(buff, v)
		buff.WriteByte(' ')
	}

	buff.WriteByte('\n')
	return buff.Bytes(), nil
}

// nolint:gochecknoglobals
var (
	coloredTraceLvl = []byte(color.HiMagentaString("[TRC]"))
	coloredDebugLvl = []byte(color.HiMagentaString("[DBG]"))
	coloredInfoLvl  = []byte(color.CyanString("[INF]"))
	coloredWarnLvl  = []byte(color.YellowString("[WRN]"))
	coloredErrLvl   = []byte(color.RedString("[ERR]"))
	coloredPanicLvl = []byte(color.New(color.BgRed).Sprint("[PNC]"))
	coloredFatalLvl = []byte(color.New(color.BgRed).Sprint("[FAT]"))
)

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
