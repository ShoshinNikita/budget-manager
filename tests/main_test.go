package tests

import (
	"net"
	"os/exec"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg"
	"github.com/ShoshinNikita/budget-manager/internal/web"
)

type (
	StartComponentFn func(*testing.T, *app.Config) *Component
	RunTestSetFn     func(*testing.T, app.Config)
)

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test")
	}

	if !checkDocker() {
		t.Fatal("docker is required for integration tests")
	}

	tests := []struct {
		name string
		//
		config            app.Config
		startComponentFns []StartComponentFn
		//
		runTestSetFn RunTestSetFn
	}{
		{
			name: "basic usage",
			//
			config: app.Config{
				DBType:     "postgres",
				PostgresDB: pg.Config{Host: "localhost", Port: 5432, User: "postgres", Database: "postgres"},
				Server:     web.Config{UseEmbed: true, SkipAuth: true, Credentials: nil, EnableProfiling: false},
			},
			startComponentFns: []StartComponentFn{startPostgreSQL},
			//
			runTestSetFn: testBasicUsage,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Start components
			cfg := tt.config
			for i, fn := range tt.startComponentFns {
				component := fn(t, &cfg)

				i := i
				t.Cleanup(func() {
					t.Logf("stop component #%d", i+1)

					err := component.Stop()
					require.NoError(err)
				})
			}

			// Start app on a free port
			serverPort := getFreePort(require)
			t.Logf("use port %d for web server", serverPort)

			cfg.Server.Port = serverPort

			app := app.NewApp(cfg, newLogger(), "", "")
			err := app.PrepareComponents()
			require.NoError(err)

			appErrCh := make(chan error, 1)
			go func() {
				appErrCh <- app.Run()
			}()
			t.Cleanup(func() {
				t.Log("stop app")

				app.Shutdown()

				err := <-appErrCh
				require.NoError(err)
			})

			// Run tests
			tt.runTestSetFn(t, cfg)
		})
	}
}

func checkDocker() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	return true
}

// startPostgreSQL starts a fresh PostgreSQL instance in a docker container.
// It updates PostgreSQL config with a chosen port
func startPostgreSQL(t *testing.T, cfg *app.Config) *Component {
	require := require.New(t)

	port := getFreePort(require)
	cfg.PostgresDB.Port = port

	t.Logf("use port %d for PostgreSQL container", port)

	c := &Component{
		ImageName: "postgres:12-alpine",
		Ports: [][2]int{
			{port, 5432},
		},
		Env: []string{
			"POSTGRES_HOST_AUTH_METHOD=trust",
		},
	}

	err := c.Run()
	require.NoError(err)

	return c
}

func getFreePort(require *require.Assertions) (port int) {
	listener, err := net.Listen("tcp", "")
	require.NoError(err)
	defer listener.Close()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	require.True(ok)

	return tcpAddr.Port
}

func newLogger() *logrus.Logger {
	return logrus.New()
}
