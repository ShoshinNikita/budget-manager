package logger

import "github.com/ShoshinNikita/go-clog/v3"

type Config struct {
	// Level is a level of logger. Valid options: debug, info, warn, error, fatal.
	// It is always debug, when debug mode is on
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}

func New(cnf Config, debug bool) *clog.Logger {

	// Use production mode by default
	log := clog.NewProdConfig().SetLevel(logLevelFromString(cnf.Level)).Build()
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
