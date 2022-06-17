package app

import (
	"io"
	"time"

	"go.etcd.io/bbolt"

	"github.com/ShoshinNikita/budget-manager/v2/internal/accounts"
	pkgAccountsService "github.com/ShoshinNikita/budget-manager/v2/internal/accounts/service"
	pkgAccountsStore "github.com/ShoshinNikita/budget-manager/v2/internal/accounts/store"
	"github.com/ShoshinNikita/budget-manager/v2/internal/categories"
	pkgCategoriesService "github.com/ShoshinNikita/budget-manager/v2/internal/categories/service"
	pkgCategoriesStore "github.com/ShoshinNikita/budget-manager/v2/internal/categories/store"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/transactions"
	pkgTransactionsService "github.com/ShoshinNikita/budget-manager/v2/internal/transactions/service"
	pkgTransactionsStore "github.com/ShoshinNikita/budget-manager/v2/internal/transactions/store"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web"
)

type App struct {
	config  Config
	version string
	gitHash string

	log       logger.Logger
	server    *web.Server
	storeConn io.Closer

	accountsStore   accounts.Store
	categoriesStore categories.Store

	transactionsStore   transactions.Store
	transactionsService transactions.Service

	accountsService   accounts.Service
	categoriesService categories.Service

	shutdownSignal chan struct{}
}

// NewApp returns a new instance of App
func NewApp(cfg Config, log logger.Logger, version, gitHash string) (*App, error) {
	app := &App{
		config:  cfg,
		version: version,
		gitHash: gitHash,
		//
		log: log,
		//
		shutdownSignal: make(chan struct{}),
	}

	if err := app.prepareStores(); err != nil {
		return nil, errors.Wrap(err, "couldn't prepare stores")
	}
	app.prepareServices()
	app.prepareWebServer()

	return app, nil
}

func (app *App) prepareStores() error {
	// TODO: move to a special package?
	boltConn, err := bbolt.Open(app.config.Store.Bolt.Path, 0o600, &bbolt.Options{
		Timeout: time.Second,
	})
	if err != nil {
		return errors.Wrap(err, "couldn't open bolt store")
	}

	app.storeConn = boltConn

	if app.accountsStore, err = pkgAccountsStore.NewBolt(boltConn); err != nil {
		return errors.Wrap(err, "couldn't create accounts store")
	}
	if app.transactionsStore, err = pkgTransactionsStore.NewBolt(boltConn); err != nil {
		return errors.Wrap(err, "couldn't create transactions store")
	}
	if app.categoriesStore, err = pkgCategoriesStore.NewBolt(boltConn); err != nil {
		return errors.Wrap(err, "couldn't create categories store")
	}
	return nil
}

func (app *App) prepareServices() {
	app.transactionsService = pkgTransactionsService.NewService(app.transactionsStore)
	app.accountsService = pkgAccountsService.NewService(app.accountsStore, app.transactionsService)
	app.categoriesService = pkgCategoriesService.NewService(app.categoriesStore)
}

func (app *App) prepareWebServer() {
	app.server = web.NewServer(app.config.Server, app.log, app.version, app.gitHash)
}

// Run runs web server. This method should be called in a goroutine
func (app *App) Run() error {
	app.log.WithFields(logger.Fields{
		"version":  app.version,
		"git_hash": app.gitHash,
	}).Info("start app")

	errCh := make(chan error, 1)
	startBackgroundJob := func(errorMsg string, f func() error) {
		go func() {
			err := f()
			if err != nil {
				app.log.WithError(err).Error(errorMsg)
			}
			errCh <- err
		}()
	}
	startBackgroundJob("web server failed", app.server.ListenAndServer)

	return <-errCh
}

// Shutdown shutdowns the app components
func (app *App) Shutdown() {
	app.log.Info("shutdown app")
	close(app.shutdownSignal)

	app.log.Debug("close web server")
	if err := app.server.Shutdown(); err != nil {
		app.log.WithError(err).Error("couldn't shutdown the server gracefully")
	}

	app.log.Debug("close the store connection")
	if err := app.storeConn.Close(); err != nil {
		app.log.WithError(err).Error("couldn't close the store connection gracefully")
	}
}
