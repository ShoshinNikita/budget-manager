package main

import (
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ShoshinNikita/budget-manager/v2/cmd"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
)

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
// @description Easy-to-use, lightweight and self-hosted solution to track your finances - [GitHub](https://github.com/ShoshinNikita/budget-manager/v2)
//
// @BasePath /api
//
// @securityDefinitions.basic BasicAuth
//
// @license.name MIT
// @license.url https://github.com/ShoshinNikita/budget-manager/v2/blob/master/LICENSE
//

func main() {
	defaultConfig := cmd.DefaultConfig{
		Version: version,
		GitHash: gitHash,
	}
	log := logger.New()

	command, err := getCommand(defaultConfig, log)
	if err != nil {
		stdlog.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- command.Run()
	}()

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	// Wait for an interrupt signal or a command error
	select {
	case <-term:
		log.Warn("got an interrupt signal")
	case err := <-errCh:
		if err != nil {
			log.WithError(err).Error("command finished with error")
		} else {
			log.Info("command finished successfully")
		}
	}

	command.Shutdown()
}

func getCommand(defaultCfg cmd.DefaultConfig, log logger.Logger) (res cmd.Command, err error) {
	var command string
	if args := os.Args[1:]; len(args) > 0 {
		command = args[0]
	}

	switch command {
	case "", "budget-manager":
		cfg, err := cmd.ParseBudgetManagerConfig(defaultCfg)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't parse config for command %q", command)
		}
		res, err = cmd.NewBudgetManagerCommand(cfg, log)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't prepare command %q", command)
		}
	default:
		return nil, errors.Errorf("invalid command %q", command)
	}

	return res, nil
}
