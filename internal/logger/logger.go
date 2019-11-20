package logger

import "github.com/ShoshinNikita/go-clog/v3"

type Config struct {
	// Mode is a mode of Logger. Valid options: prod, production, dev, develop.
	// Default value is prod
	Mode string `env:"LOGGER_MODE" envDefault:"prod"`

	// Level is a level of logger. Valid options: debug, info, warn, error, fatal.
	// It is always debug, when debug mode is on
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}

func New(cnf Config, debug bool) *clog.Logger {
	var loggerConfig *clog.Config
	switch cnf.Mode {
	case "prod", "production":
		loggerConfig = clog.NewProdConfig()
	case "dev", "develop":
		loggerConfig = clog.NewDevConfig()
	default:
		loggerConfig = clog.NewDevConfig()
	}

	// Use production mode by default
	log := loggerConfig.SetLevel(logLevelFromString(cnf.Level)).Build()
	if debug {
		log = clog.NewDevConfig().SetLevel(clog.LevelDebug).Build()
	}

	return log
}

func logLevelFromString(lvl string) clog.LogLevel {
	switch lvl {
	case "dbg", "debug":
		return clog.LevelDebug
	case "inf", "info":
		return clog.LevelInfo
	case "warn", "warning":
		return clog.LevelWarn
	case "err", "error":
		return clog.LevelError
	case "fatal":
		return clog.LevelFatal
	default:
		return clog.LevelInfo
	}
}
