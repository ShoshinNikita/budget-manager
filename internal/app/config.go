package app

import (
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/db/sqlite"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/env"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

type Config struct {
	Logger logger.Config

	DBType     string
	PostgresDB pg.Config
	SQLiteDB   sqlite.Config

	Server web.Config
}

func ParseConfig() (Config, error) {
	cfg := Config{
		Logger: logger.Config{
			Mode:  "prod",
			Level: "info",
		},
		//
		DBType: "postgres",
		PostgresDB: pg.Config{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "",
			Database: "postgres",
		},
		SQLiteDB: sqlite.Config{
			Path: "./var/budget-manager.db",
		},
		//
		Server: web.Config{
			Port:            8080,
			UseEmbed:        true,
			EnableProfiling: false,
			Auth: web.AuthConfig{
				Disable:         false,
				Type:            "basic",
				BasicAuthCreds:  nil,
				TOTPAuthSecrets: nil,
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
		{"DB_TYPE", &cfg.DBType},
		{"DB_PG_HOST", &cfg.PostgresDB.Host},
		{"DB_PG_PORT", &cfg.PostgresDB.Port},
		{"DB_PG_USER", &cfg.PostgresDB.User},
		{"DB_PG_PASSWORD", &cfg.PostgresDB.Password},
		{"DB_PG_DATABASE", &cfg.PostgresDB.Database},
		{"DB_SQLITE_PATH", &cfg.SQLiteDB.Path},
		//
		{"SERVER_PORT", &cfg.Server.Port},
		{"SERVER_USE_EMBED", &cfg.Server.UseEmbed},
		{"SERVER_ENABLE_PROFILING", &cfg.Server.EnableProfiling},
		{"SERVER_AUTH_DISABLE", &cfg.Server.Auth.Disable},
		{"SERVER_AUTH_TYPE", &cfg.Server.Auth.Type},
		{"SERVER_AUTH_BASIC_CREDS", &cfg.Server.Auth.BasicAuthCreds},
		{"SERVER_AUTH_TOTP_SECRETS", &cfg.Server.Auth.TOTPAuthSecrets},
	} {
		if err := env.Load(v.key, v.target); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}
