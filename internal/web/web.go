package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ShoshinNikita/budget-manager/v2/internal/app"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/logger"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/api"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/middlewares"
	"github.com/ShoshinNikita/budget-manager/v2/internal/web/pprof"
)

type Server struct {
	config  Config
	version string
	gitHash string

	log     logger.Logger
	service app.Service

	server *http.Server
}

func NewServer(cfg Config, log logger.Logger, service app.Service, version, gitHash string) *Server {
	s := &Server{
		config:  cfg,
		version: version,
		gitHash: gitHash,
		//
		log:     log,
		service: service,
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

	api.New(s.service, s.log, s.version, s.gitHash).RegisterHandlers(router)

	if s.config.EnableProfiling {
		s.log.Warn("pprof handlers are enabled")

		pprof.New().RegisterHandlers(router)
	}

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
