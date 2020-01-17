package web

import (
	"context"
	"encoding"
	htmlTemplate "html/template"
	"io"
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
	db       Database
	tplStore TemplateStore

	server *http.Server

	config Config
}

type Database interface {
	// Month
	GetMonths(ctx context.Context, year int) ([]*db.Month, error)
	GetMonth(ctx context.Context, id uint) (*db.Month, error)
	GetMonthID(ctx context.Context, year, month int) (id uint, err error)

	// Day
	GetDay(ctx context.Context, id uint) (*db.Day, error)
	GetDayIDByDate(ctx context.Context, year, month, day int) (id uint, err error)

	// Income
	AddIncome(ctx context.Context, args db.AddIncomeArgs) (id uint, err error)
	EditIncome(ctx context.Context, args db.EditIncomeArgs) error
	RemoveIncome(ctx context.Context, id uint) error

	// Monthly Payment
	AddMonthlyPayment(ctx context.Context, args db.AddMonthlyPaymentArgs) (id uint, err error)
	EditMonthlyPayment(ctx context.Context, args db.EditMonthlyPaymentArgs) error
	RemoveMonthlyPayment(ctx context.Context, id uint) error

	// Spend
	AddSpend(ctx context.Context, args db.AddSpendArgs) (id uint, err error)
	EditSpend(ctx context.Context, args db.EditSpendArgs) error
	RemoveSpend(ctx context.Context, id uint) error

	// Spend Type
	GetSpendTypes(ctx context.Context) ([]db.SpendType, error)
	AddSpendType(ctx context.Context, name string) (id uint, err error)
	EditSpendType(ctx context.Context, id uint, newName string) error
	RemoveSpendType(ctx context.Context, id uint) error
}

type TemplateStore interface {
	Get(path string) *htmlTemplate.Template
	Execute(path string, w io.Writer, data interface{}) error
}

func NewServer(cnf Config, db Database, log *clog.Logger, debug bool) *Server {
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
