package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

func TestConfig(t *testing.T) {
	require := require.New(t)

	envs := []struct{ key, value string }{
		{"DEBUG", "true"},
		{"LOGGER_MODE", "develop"},
		{"LOGGER_LEVEL", "fatal"},
		{"DB_HOST", "example.com"},
		{"DB_PORT", "8888"},
		{"DB_USER", "user"},
		{"DB_PASSWORD", "qwerty"},
		{"DB_DATABASE", "db"},
		{"SERVER_PORT", "6666"},
		{"SERVER_CREDENTIALS", "user:qwerty,admin:admin"},
	}
	for _, env := range envs {
		os.Setenv(env.key, env.value)
	}

	want := Config{
		Debug: true,
		Logger: logger.Config{
			Level: "fatal",
			Mode:  "develop",
		},
		DB: db.Config{
			Host:     "example.com",
			Port:     8888,
			User:     "user",
			Password: "qwerty",
			Database: "db",
		},
		Server: web.Config{
			Port:           6666,
			CacheTemplates: true,
			Credentials: web.Credentials{
				"user":  "qwerty",
				"admin": "admin",
			},
		},
	}

	app := NewApp()
	err := app.ParseConfig()
	require.Nil(err)
	require.Equal(want, app.config)
}