package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/v2/internal/web"
)

func TestParseBudgetManagerConfig(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	envs := []struct{ key, value string }{
		{"STORE_BOLT_PATH", "./db.db"},
		//
		{"SERVER_PORT", "6666"},
		{"SERVER_USE_EMBED", "false"},
		{"SERVER_ENABLE_PROFILING", "true"},
		{"SERVER_AUTH_DISABLE", "true"},
		{"SERVER_AUTH_BASIC_CREDS", "user:qwerty,admin:admin"},
	}
	for _, env := range envs {
		os.Setenv(env.key, env.value)
	}

	want := BudgetManagerConfig{
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
	want.Store.Bolt.Path = "./db.db"

	cfg, err := ParseBudgetManagerConfig(DefaultConfig{})
	require.Nil(err)
	require.Equal(want, cfg)
}