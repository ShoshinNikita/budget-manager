package tests

import (
	"fmt"
	"testing"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/db/sqlite"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

type Test interface {
	Test(t *testing.T, host string)
}

// TestFn is a single test function that implements 'Test' interface
type TestFn func(t *testing.T, host string)

func (fn TestFn) Test(t *testing.T, host string) {
	t.Helper()

	fn(t, host)
}

// TestCases is a set of test cases that implements 'Test' interface.
// All test cases are run consistently.
type TestCases []struct {
	Name string
	Fn   TestFn
}

func (testCases TestCases) Test(t *testing.T, host string) {
	t.Helper()

	for _, tt := range testCases {
		tt := tt
		ok := t.Run(tt.Name, func(t *testing.T) {
			tt.Fn(t, host)
		})
		if !ok {
			t.FailNow()
		}
	}
}

type TestEnv struct {
	Name       string
	Cfg        app.Config
	Components []StartComponentFn
}

type TestEnvOption func(env *TestEnv)

// RunTest runs the passed test with all possible environments. Environment options are
// applied to all environments
func RunTest(t *testing.T, test Test, opts ...TestEnvOption) {
	t.Helper()

	for _, env := range []TestEnv{
		{
			Name:       "postgres",
			Cfg:        getDefaultConfig(db.Postgres),
			Components: []StartComponentFn{StartPostgreSQL},
		},
		{
			Name:       "sqlite",
			Cfg:        getDefaultConfig(db.Sqlite3),
			Components: []StartComponentFn{StartSQLite},
		},
	} {
		env := env
		for _, opt := range opts {
			opt(&env)
		}
		t.Run(env.Name, func(t *testing.T) {
			t.Parallel()

			prepareApp(t, &env.Cfg, env.Components...)

			host := fmt.Sprintf("localhost:%d", env.Cfg.Server.Port)

			test.Test(t, host)
		})
	}
}

func getDefaultConfig(dbType db.Type) app.Config {
	return app.Config{
		Logger: logger.Config{
			Mode:  "dev",
			Level: "error",
		},
		DB: app.DBConfig{
			Type: dbType,
			Postgres: pg.Config{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Database: "postgres",
			},
			SQLite: sqlite.Config{
				Path: "",
			},
		},
		Server: web.Config{
			UseEmbed:        true,
			EnableProfiling: false,
			Auth: web.AuthConfig{
				Disable: true,
			},
		},
	}
}
