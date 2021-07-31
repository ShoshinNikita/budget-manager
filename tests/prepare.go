package tests

import (
	"os/exec"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
)

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
	serverPort := getFreePort(t)
	t.Logf("use port %d for web server", serverPort)

	cfg.Server.Port = serverPort

	app := app.NewApp(*cfg, logrus.New(), "", "")
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

	// Make sure app is started
	time.Sleep(100 * time.Millisecond)
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
