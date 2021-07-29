package tests

import (
	"net"
	"os/exec"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
)

type StartComponentFn func(*testing.T, *app.Config) *Component

func prepareApp(t *testing.T, cfg *app.Config, components ...StartComponentFn) {
	t.Helper()

	checkTestMode(t)
	checkDocker(t)

	require := require.New(t)

	// Start components
	for _, fn := range components {
		component := fn(t, cfg)

		t.Cleanup(func() {
			t.Logf("stop component %q", component.ImageName)

			err := component.Stop()
			require.NoError(err)
		})
	}

	// Start app on a free port
	serverPort := getFreePort(require)
	t.Logf("use port %d for web server", serverPort)

	cfg.Server.Port = serverPort

	app := app.NewApp(*cfg, newLogger(), "", "")
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
}

// checkTestMode checks the test mode (whether the -short flag is set) and skips the test if needed
func checkTestMode(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("skip integration test")
	}
}

func checkDocker(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Fatal("docker is required for integration tests")
	}
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
