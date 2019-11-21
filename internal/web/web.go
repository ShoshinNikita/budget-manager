package web

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ShoshinNikita/go-clog/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget_manager/internal/db"
)

type Config struct {
	Port int `env:"SERVER_PORT" envDefault:"8080"`
}

type Server struct {
	server *http.Server
	db     *db.DB
	log    *clog.Logger

	config Config
}

func NewServer(cnf Config, db *db.DB, log *clog.Logger) *Server {
	//nolint:gosimple
	return &Server{
		db:     db,
		log:    log,
		config: cnf,
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
		Addr:    ":" + strconv.Itoa(s.config.Port),
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
