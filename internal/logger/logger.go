package logger

import (
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/caller"
)

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
