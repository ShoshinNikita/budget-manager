package app

import (
	"github.com/caarlos0/env/v6"

	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/db/sqlite"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

type Config struct {
	Logger logger.Config

	DBType     string `env:"DB_TYPE" envDefault:"postgres"`
	PostgresDB pg.Config
	SQLiteDB   sqlite.Config

	Server web.Config
}

func ParseConfig() (cfg Config, err error) {
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
