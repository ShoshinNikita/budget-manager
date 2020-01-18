package logger

import (
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
		log.SetFormatter(&logrus.TextFormatter{})
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
