package web

import (
	"context"
	"encoding"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/logger"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/web/api"
	"github.com/ShoshinNikita/budget-manager/internal/web/pages"
	"github.com/ShoshinNikita/budget-manager/static"
)

// -------------------------------------------------
// Config
// -------------------------------------------------

type Config struct { //nolint:maligned
	Port int

	// UseEmbed defines whether server should use embedded templates and static files
	UseEmbed bool

	// SkipAuth disables auth
	SkipAuth bool

	// Credentials is a list of pairs 'login:password' separated by comma.
	// Passwords must be hashed using BCrypt
	Credentials Credentials

	EnableProfiling bool
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
	fileHandler = cacheMiddleware(fileHandler, time.Hour*24*30, s.gitHash) // cache for 1 month
	router.Handle("/static/", fileHandler)

	// Wrap the handler in middlewares. The last middleware will be called first and so on
	var handler http.Handler = router
	if !s.config.SkipAuth {
		handler = s.basicAuthMiddleware(handler)
	} else {
		s.log.Warn("auth is disabled")
	}
	handler = s.loggingMiddleware(handler)
	handler = s.requestIDMeddleware(handler)

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
