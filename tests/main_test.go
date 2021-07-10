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

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test")
	}

	if !checkDocker() {
		t.Fatal("docker is required for integration tests")
	}

	tests := []struct {
		name           string
		config         app.Config
		startComponent func(*testing.T, *require.Assertions, *app.Config) *Component
	}{
		{
			name: "pg",
			config: app.Config{
				DBType: "postgres",
				PostgresDB: pg.Config{
					Host:     "localhost",
					Port:     5432,
					User:     "postgres",
					Password: "postgres",
					Database: "postgres",
				},
				Server: web.Config{
					Port:            0, // will be set later
					UseEmbed:        true,
					SkipAuth:        true,
					Credentials:     nil,
					EnableProfiling: false,
				},
			},
			startComponent: func(t *testing.T, require *require.Assertions, cfg *app.Config) *Component {
				port := getFreePort(require)
				cfg.PostgresDB.Port = port

				t.Logf("use port %d for PostgreSQL container", port)

				c := &Component{
					ImageName: "postgres:12-alpine",
					Ports: [][2]int{
						{port, 5432},
					},
					Env: []string{
						"POSTGRES_PASSWORD=postgres",
					},
				}

				err := c.Run()
				require.NoError(err)

				return c
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			cfg := tt.config
			if tt.startComponent != nil {
				component := tt.startComponent(t, require, &cfg)

				t.Cleanup(func() {
					t.Log("stop component")

					err := component.Stop()
					require.NoError(err)
				})
			}

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

			runTests(require)
		})
	}
}

func checkDocker() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	return true
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
