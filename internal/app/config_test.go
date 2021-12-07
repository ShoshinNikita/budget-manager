package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/db/sqlite"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

func TestParseConfig(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	envs := []struct{ key, value string }{
		{"LOGGER_MODE", "develop"},
		{"LOGGER_LEVEL", "fatal"},
		{"DB_TYPE", "mongodb"},
		{"DB_PG_HOST", "example.com"},
		{"DB_PG_PORT", "8888"},
		{"DB_PG_USER", "user"},
		{"DB_PG_PASSWORD", "qwerty"},
		{"DB_PG_DATABASE", "db"},
		{"DB_SQLITE_PATH", "./var/db.db"},
		{"SERVER_PORT", "6666"},
		{"SERVER_USE_EMBED", "false"},
		{"SERVER_ENABLE_PROFILING", "true"},
		{"SERVER_AUTH_DISABLE", "true"},
		{"SERVER_AUTH_BASIC_CREDS", "user:qwerty,admin:admin"},
	}
	for _, env := range envs {
		os.Setenv(env.key, env.value)
	}

	want := Config{
		Logger: logger.Config{
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
		SQLiteDB: sqlite.Config{
			Path: "./var/db.db",
		},
		Server: web.Config{
			Port:            6666,
			UseEmbed:        false,
			EnableProfiling: true,
			Auth: web.AuthConfig{
				Disable: true,
				BasicAuthCreds: web.Credentials{
					"user":  "qwerty",
					"admin": "admin",
				},
			},
		},
	}

	cfg, err := ParseConfig()
	require.Nil(err)
	require.Equal(want, cfg)
}
