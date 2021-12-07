package app

import (
	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/db/sqlite"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/env"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

type Config struct {
	Logger logger.Config
	DB     DBConfig
	Server web.Config
}

type DBConfig struct {
	Type     db.Type
	Postgres pg.Config
	SQLite   sqlite.Config
}

func ParseConfig() (Config, error) {
	cfg := Config{
		Logger: logger.Config{
			Mode:  "prod",
			Level: "info",
		},
		DB: DBConfig{
			Type: db.Postgres,
			Postgres: pg.Config{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				Database: "postgres",
			},
			SQLite: sqlite.Config{
				Path: "./var/budget-manager.db",
			},
		},
		//
		Server: web.Config{
			Port:            8080,
			UseEmbed:        true,
			EnableProfiling: false,
			Auth: web.AuthConfig{
				Disable:        false,
				BasicAuthCreds: nil,
			},
		},
	}

	for _, v := range []struct {
		key    string
		target interface{}
	}{
		{"LOGGER_MODE", &cfg.Logger.Mode},
		{"LOGGER_LEVEL", &cfg.Logger.Level},
		//
		{"DB_TYPE", &cfg.DB.Type},
		{"DB_PG_HOST", &cfg.DB.Postgres.Host},
		{"DB_PG_PORT", &cfg.DB.Postgres.Port},
		{"DB_PG_USER", &cfg.DB.Postgres.User},
		{"DB_PG_PASSWORD", &cfg.DB.Postgres.Password},
		{"DB_PG_DATABASE", &cfg.DB.Postgres.Database},
		{"DB_SQLITE_PATH", &cfg.DB.SQLite.Path},
		//
		{"SERVER_PORT", &cfg.Server.Port},
		{"SERVER_USE_EMBED", &cfg.Server.UseEmbed},
		{"SERVER_ENABLE_PROFILING", &cfg.Server.EnableProfiling},
		{"SERVER_AUTH_DISABLE", &cfg.Server.Auth.Disable},
		{"SERVER_AUTH_BASIC_CREDS", &cfg.Server.Auth.BasicAuthCreds},
	} {
		if err := env.Load(v.key, v.target); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}
