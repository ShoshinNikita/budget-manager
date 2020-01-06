package web

import (
	"context"
	"encoding"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ShoshinNikita/go-clog/v3"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/web/templates"
)

// -------------------------------------------------
// Config
// -------------------------------------------------

type Config struct {
	Port int `env:"SERVER_PORT" envDefault:"8080"`

	// CacheTemplates defines whether templates have to be loaded from disk every request.
	// It is useful during development. So, it is always false when Debug mode is on
	CacheTemplates bool `env:"SERVER_CACHE_TEMPLATES" envDefault:"true"`

	// SkipAuth disables auth. Works only in Debug mode!
	SkipAuth bool `env:"SERVER_SKIP_AUTH"`
	// Credentials is a list of pairs 'login:password' separated by comma.
	// Example: "login:password,user:qwerty"
	Credentials Credentials `env:"SERVER_CREDENTIALS"`
}

var _ encoding.TextUnmarshaler = &Credentials{}

type Credentials map[string]string

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
	log      *clog.Logger
	db       *db.DB
	tplStore *templates.TemplateStore

	server *http.Server

	config Config
}

func NewServer(cnf Config, db *db.DB, log *clog.Logger, debug bool) *Server {
	if debug {
		// Load templates every request
		cnf.CacheTemplates = false
	} else {
		// Always false when Debug mode is off
		cnf.SkipAuth = false
	}

	//nolint:gosimple
	return &Server{
		db:       db,
		log:      log,
		tplStore: templates.NewTemplateStore(log.WithPrefix("[template store]"), cnf.CacheTemplates),
		config:   cnf,
	}
}

func (s *Server) Prepare() {
	router := mux.NewRouter()
	router.StrictSlash(true)

	// Add middlewares

	if !s.config.SkipAuth {
		router.Use(s.basicAuthMiddleware)
	} else {
		s.log.Warn("auth is disabled")
	}

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
