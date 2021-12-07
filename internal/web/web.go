package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/web/api"
	"github.com/ShoshinNikita/budget-manager/internal/web/middlewares"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages"
	"github.com/ShoshinNikita/budget-manager/static"
)

type Server struct {
	config Config
	log    logger.Logger
	db     Database

	server *http.Server

	version string
	gitHash string
}

type Database interface {
	api.DB
	pages.DB
}

func NewServer(cfg Config, db Database, log logger.Logger, version, gitHash string) *Server {
	s := &Server{
		config: cfg,
		db:     db,
		log:    log,
		//
		version: version,
		gitHash: gitHash,
	}
	s.server = &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: s.buildServerHandler(),
	}

	return s
}

func (s *Server) buildServerHandler() http.Handler {
	router := http.NewServeMux()

	if !s.config.UseEmbed {
		s.log.Warn("don't use embedded templates and static files")
	}

	// Add API routes
	s.addRoutes(router)
	if s.config.EnableProfiling {
		// Enable pprof handlers
		s.log.Warn("pprof handlers are enabled")
		s.addPprofRoutes(router)
	}

	// Add File Handler
	fs := http.FS(static.New(s.config.UseEmbed))
	fileHandler := http.StripPrefix("/static/", http.FileServer(fs))
	fileHandler = middlewares.CachingMiddleware(fileHandler, time.Hour*24*30, s.gitHash) // cache for 1 month
	router.Handle("/static/", fileHandler)

	// Wrap the handler in middlewares. The last middleware will be called first and so on
	var handler http.Handler = router
	if !s.config.Auth.Disable {
		handler = middlewares.BasicAuthMiddleware(handler, s.config.Auth.BasicAuthCreds, s.log)
		if len(s.config.Auth.BasicAuthCreds) == 0 {
			s.log.Warn("auth is enabled, but list of creds is empty")
		}
	} else {
		s.log.Warn("auth is disabled")
	}
	handler = middlewares.LoggingMiddleware(handler, s.log)
	handler = middlewares.RequestIDMeddleware(handler)

	return handler
}

func (s Server) ListenAndServer() error {
	s.log.Debug("start server")

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "server error")
	}

	return nil
}

func (s Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return s.server.Shutdown(ctx)
}
