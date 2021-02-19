package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	envs := []struct{ key, value string }{
		{"DEBUG", "true"},
		{"LOGGER_MODE", "develop"},
		{"LOGGER_LEVEL", "fatal"},
		{"DB_TYPE", "mongodb"},
		{"DB_PG_HOST", "example.com"},
		{"DB_PG_PORT", "8888"},
		{"DB_PG_USER", "user"},
		{"DB_PG_PASSWORD", "qwerty"},
		{"DB_PG_DATABASE", "db"},
		{"SERVER_PORT", "6666"},
		{"SERVER_USE_EMBED", "false"},
		{"SERVER_CREDENTIALS", "user:qwerty,admin:admin"},
	}
	for _, env := range envs {
		os.Setenv(env.key, env.value)
	}

	want := Config{
		Logger: logger.Config{
			Debug: true,
			Level: "fatal",
			Mode:  "develop",
		},
		DBType: "mongodb",
		PostgresDB: pg.Config{
			Host:     "example.com",
			Port:     8888,
			User:     "user",
			Password: "qwerty",
			Database: "db",
		},
		Server: web.Config{
			Port:     6666,
			UseEmbed: false,
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
