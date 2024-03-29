package main

import (
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
)

//nolint:gochecknoglobals
var (
	// version is a version of the app. It must be set during the build process with -ldflags flag
	version = "unknown"
	// gitHash is the last commit hash. It must be set during the build process with -ldflags flag
	gitHash = "unknown"
)

// Swagger General Info
//
//nolint:lll
//
// @title Budget Manager API
// @version v0.2
// @description Easy-to-use, lightweight and self-hosted solution to track your finances - [GitHub](https://github.com/ShoshinNikita/budget-manager)
//
// @BasePath /api
//
// @securityDefinitions.basic BasicAuth
//
// @license.name MIT
// @license.url https://github.com/ShoshinNikita/budget-manager/blob/master/LICENSE
//

func main() {
	cfg, err := app.ParseConfig()
	if err != nil {
		stdlog.Fatalf("couldn't parse config: %s\n", err)
	}
	log := logger.New(cfg.Logger)

	app := app.NewApp(cfg, log, version, gitHash)

	if err := app.PrepareComponents(); err != nil {
		stdlog.Fatalf("couldn't prepare components: %s\n", err)
	}

	appErrCh := make(chan error, 1)
	go func() {
		appErrCh <- app.Run()
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	// Wait for an interrupt signal or an app error
	select {
	case <-term:
		log.Warn("got an interrupt signal")
	case err := <-appErrCh:
		log.WithError(err).Error("app finished with error")
	}

	app.Shutdown()
}
