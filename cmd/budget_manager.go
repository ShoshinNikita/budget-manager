package cmd

import (
	"io"
	"time"

	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/env"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/service"
	"github.com/ShoshinNikita/budget-manager/v2/internal/store/bolt"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web"
)

type BudgetManagerConfig struct {
	DefaultConfig

	Store struct {
		Bolt struct {
			Path string
		}
	}
	Server web.Config
}

func ParseBudgetManagerConfig(defaultCfg DefaultConfig) (BudgetManagerConfig, error) {
	cfg := BudgetManagerConfig{
		DefaultConfig: defaultCfg,
		Server: web.Config{
			Port:            8080,
			UseEmbed:        true,
			EnableProfiling: false,
			Auth: web.AuthConfig{
				Disable:        false,
				BasicAuthCreds: nil,
			},
		},
	}
	cfg.Store.Bolt.Path = "./var/budget-manager.db"

	for _, v := range []struct {
		key    string
		target interface{}
	}{
		{"STORE_BOLT_PATH", &cfg.Store.Bolt.Path},
		//
		{"SERVER_PORT", &cfg.Server.Port},
		{"SERVER_USE_EMBED", &cfg.Server.UseEmbed},
		{"SERVER_ENABLE_PROFILING", &cfg.Server.EnableProfiling},
		{"SERVER_AUTH_DISABLE", &cfg.Server.Auth.Disable},
		{"SERVER_AUTH_BASIC_CREDS", &cfg.Server.Auth.BasicAuthCreds},
	} {
		if err := env.Load(v.key, v.target); err != nil {
			return BudgetManagerConfig{}, err
		}
	}
	return cfg, nil
}

// BudgetManagerCommand runs Budget Manager. It is a default command
type BudgetManagerCommand struct {
	config BudgetManagerConfig

	log logger.Logger

	storeConn        io.Closer
	accountStore     app.AccountStore
	transactionStore app.TransactionStore
	categoryStore    app.CategoryStore

	service app.Service

	server *web.Server

	shutdownSignal chan struct{}
}

// NewBudgetManagerCommand returns a command to run Budget Manager
func NewBudgetManagerCommand(cfg BudgetManagerConfig, log logger.Logger) (*BudgetManagerCommand, error) {
	cmd := &BudgetManagerCommand{
		config:         cfg,
		log:            log,
		shutdownSignal: make(chan struct{}),
	}

	if err := cmd.prepareStores(); err != nil {
		return nil, errors.Wrap(err, "couldn't prepare stores")
	}
	cmd.prepareService()
	cmd.prepareWebServer()

	return cmd, nil
}

func (cmd *BudgetManagerCommand) prepareStores() error {
	// TODO: move to a special package?
	boltConn, err := bbolt.Open(cmd.config.Store.Bolt.Path, 0o600, &bbolt.Options{
		Timeout: time.Second,
	})
	if err != nil {
		return errors.Wrap(err, "couldn't open bolt store")
	}

	cmd.storeConn = boltConn

	if cmd.accountStore, err = bolt.NewAccountsStore(boltConn); err != nil {
		return errors.Wrap(err, "couldn't create accounts store")
	}
	if cmd.transactionStore, err = bolt.NewTransactionsStore(boltConn); err != nil {
		return errors.Wrap(err, "couldn't create transactions store")
	}
	if cmd.categoryStore, err = bolt.NewCategoriesStore(boltConn); err != nil {
		return errors.Wrap(err, "couldn't create categories store")
	}
	return nil
}

func (cmd *BudgetManagerCommand) prepareService() {
	cmd.service = service.NewService(cmd.accountStore, cmd.transactionStore, cmd.categoryStore)
}

func (cmd *BudgetManagerCommand) prepareWebServer() {
	cmd.server = web.NewServer(cmd.config.Server, cmd.log, cmd.config.Version, cmd.config.GitHash)
}

// Run runs web server. This method should be called in a goroutine
func (cmd *BudgetManagerCommand) Run() error {
	cmd.log.WithFields(logger.Fields{
		"version":  cmd.config.Version,
		"git_hash": cmd.config.GitHash,
	}).Info("start")

	errCh := make(chan error, 1)
	startBackgroundJob := func(errorMsg string, f func() error) {
		go func() {
			err := f()
			if err != nil {
				cmd.log.WithError(err).Error(errorMsg)
			}
			errCh <- err
		}()
	}
	startBackgroundJob("web server failed", cmd.server.ListenAndServer)

	return <-errCh
}

// Shutdown shutdowns the components
func (cmd *BudgetManagerCommand) Shutdown() {
	cmd.log.Info("shutdown")
	close(cmd.shutdownSignal)

	cmd.log.Debug("close web server")
	if err := cmd.server.Shutdown(); err != nil {
		cmd.log.WithError(err).Error("couldn't shutdown the server gracefully")
	}

	cmd.log.Debug("close the store connection")
	if err := cmd.storeConn.Close(); err != nil {
		cmd.log.WithError(err).Error("couldn't close the store connection gracefully")
	}
}