package config

import (
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/env"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web"
)

type Config struct {
	Store  StoreConfig
	Server web.Config
}

type StoreConfig struct {
	Bolt BoltConfig
}

type BoltConfig struct {
	Path string
}

func ParseConfig() (Config, error) {
	cfg := Config{
		Store: StoreConfig{
			Bolt: BoltConfig{
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
		{"STORE_BOLT_PATH", &cfg.Store.Bolt.Path},
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
