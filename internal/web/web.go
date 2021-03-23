package web

import (
	"context"
	"encoding"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/ShoshinNikita/budget-manager/internal/web/api"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages"
	"github.com/ShoshinNikita/budget-manager/static"
)

// -------------------------------------------------
// Config
// -------------------------------------------------

type Config struct { //nolint:maligned
	Port int `env:"SERVER_PORT" envDefault:"8080"`

	// UseEmbed defines whether server should use embedded templates and static files.
	UseEmbed bool `env:"SERVER_USE_EMBED" envDefault:"true"`

	// SkipAuth disables auth. Works only in Debug mode!
	SkipAuth bool `env:"SERVER_SKIP_AUTH"`
	// Credentials is a list of pairs 'login:password' separated by comma.
	// Example: "login:password,user:qwerty"
	Credentials Credentials `env:"SERVER_CREDENTIALS"`

	EnableProfiling bool `env:"SERVER_ENABLE_PROFILING" envDefault:"false"`
}

type Credentials map[string]string

var _ encoding.TextUnmarshaler = &Credentials{}

func (c *Credentials) UnmarshalText(text []byte) error {
	m := make(Credentials)

	pairs := strings.Split(string(text), ",")
	for _, pair := range pairs {
		split := strings.Split(pair, ":")
		if len(split) != 2 {
			return errors.New("invalid credential pair")
		}

		login := split[0]
		password := split[1]
		if login == "" || password == "" {
			return errors.New("login and password can't be empty")
		}

		m[login] = password
	}

	*c = m

	return nil
}

// -------------------------------------------------
// Server
// -------------------------------------------------

type Server struct {
	config Config
	log    logrus.FieldLogger
	db     Database

	server *http.Server

	version string
	gitHash string
}

type Database interface {
	api.DB
	pages.DB
}

func NewServer(cnf Config, db Database, log logrus.FieldLogger, version, gitHash string) *Server {
	return &Server{
		config: cnf,
		db:     db,
		log:    log,
		//
		version: version,
		gitHash: gitHash,
	}
}

func (s *Server) Prepare() {
	router := mux.NewRouter()
	router.StrictSlash(true)

	if !s.config.UseEmbed {
		s.log.Warn("don't use embedded templates and static files")
	}

	// Add middlewares
	router.Use(s.requestIDMeddleware)
	router.Use(s.loggingMiddleware)
	if !s.config.SkipAuth {
		router.Use(s.basicAuthMiddleware)
	} else {
		s.log.Warn("auth is disabled")
	}

	// Add API routes
	s.log.Debug("add routes")
	s.addRoutes(router)
	if s.config.EnableProfiling {
		// Enable pprof handlers
		s.log.Warn("pprof handlers are enabled")
		s.addPprofRoutes(router)
	}

	// Add File Handler
	fs := http.FS(static.New(s.config.UseEmbed))
	fileHandler := http.StripPrefix("/static/", http.FileServer(fs))
	fileHandler = cacheMiddleware(fileHandler, time.Hour*24*30, s.gitHash) // cache for 1 month
	router.PathPrefix("/static/").Handler(fileHandler)

	s.server = &http.Server{
		Addr:    ":" + strconv.Itoa(s.config.Port),
		Handler: router,
	}
}

func (s Server) ListenAndServer() error {
	s.log.Info("start server")

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.WithError(err).Error("server error")
		return errors.Wrap(err, "server error")
	}

	return nil
}

func (s Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.log.WithError(err).Errorf("couldn't shutdown server gracefully")
		return err
	}

	return nil
}
