package logger

type Config struct {
	// Level is a level of logger. Valid options: debug, info, warn, error, fatal.
	// It is always debug, when debug mode is on
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}
