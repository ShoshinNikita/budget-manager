package logger

import (
	"runtime"

	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/caller"
	"github.com/ShoshinNikita/budget-manager/internal/web/api/models"
)

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger
	WithRequest(models.Request) Logger

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

type Fields logrus.Fields

type Config struct {
	Debug bool `env:"DEBUG" envDefault:"false"`

	// Mode is a mode of Logger. Valid options: prod, production, dev, develop.
	// Default value is prod
	Mode string `env:"LOGGER_MODE" envDefault:"prod"`

	// Level is a level of logger. Valid options: debug, info, warn, error, fatal.
	// It is always debug, when debug mode is on
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}

func New(cnf Config) Logger {
	log := logrus.New()
	log.SetReportCaller(true)

	logLevel := parseLogLevel(cnf.Level)
	if cnf.Debug {
		// Always use the debug level in the debug mode
		logLevel = logrus.DebugLevel
	}
	log.SetLevel(logLevel)

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

	return logger{log}
}

// parseLogLevel converts passed string to a logrus log level
func parseLogLevel(lvl string) logrus.Level {
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

// logger is a wrapper for logrus.FieldLogger that implements Logger interface
type logger struct {
	logrus.FieldLogger
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{l.FieldLogger.WithField(key, value)}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{l.FieldLogger.WithFields(logrus.Fields(fields))}
}

func (l logger) WithError(err error) Logger {
	return logger{l.FieldLogger.WithError(err)}
}

func (l logger) WithRequest(req models.Request) Logger {
	if fields := structToFields(req); len(fields) != 0 {
		return l.WithFields(fields)
	}
	return l
}
