package app

import (
	"github.com/ShoshinNikita/budget-manager/v2/internal/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/env"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web"
)

type Config struct {
	Logger logger.Config
	DB     DBConfig
	Server web.Config
}

type DBConfig struct{}

func ParseConfig() (Config, error) {
	cfg := Config{
		Logger: logger.Config{
			Mode:  "prod",
			Level: "info",
		},
		DB: DBConfig{},
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
