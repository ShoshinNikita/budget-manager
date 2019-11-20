package web

import (
	"context"
	"net/http"
	"time"

	"github.com/ShoshinNikita/go-clog/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget_manager/internal/db"
)

type Server struct {
	server *http.Server
	db     *db.DB
	log    *clog.Logger

	config serverConfig
}

type serverConfig struct {
	Port string
}

type NewServerOptions struct {
	Port string
}

func NewServer(opts NewServerOptions, db *db.DB, log *clog.Logger) *Server {
	//nolint:gosimple
	return &Server{
		db:  db,
		log: log.WithPrefix("[server]"),
		config: serverConfig{
			Port: opts.Port,
		},
	}
}

func (s *Server) Prepare() {
	router := mux.NewRouter()

	// Add API routes
	s.log.Debug("add routes")
	s.addRoutes(router)

	// Add File Handler
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	fileHandler = cacheMiddleware(fileHandler, time.Hour*24*30) // cache for 1 month
	router.PathPrefix("/static/").Handler(fileHandler)

	s.server = &http.Server{
		Addr:    s.config.Port,
		Handler: router,
	}
}

func (s Server) ListenAndServer() error {
	s.log.Info("start server")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		err = errors.Wrap(err, "ListenAndServe returned error")

		s.log.Error(err)
		return err
	}

	return nil
}

func (s Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err := s.server.Shutdown(ctx)
	if err != nil {
		s.log.Errorf("can't shutdown server gracefully: %s", err)
	}

	return err
}
