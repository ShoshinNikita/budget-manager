package pg

//nolint:gci
import (
	"fmt"

	_ "github.com/lib/pq" // register PostgreSQL driver

	"github.com/ShoshinNikita/budget-manager/internal/db/base"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg/migrations"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
)

type Config struct {
	Host     string `env:"DB_PG_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PG_PORT" envDefault:"5432"`
	User     string `env:"DB_PG_USER" envDefault:"postgres"`
	Password string `env:"DB_PG_PASSWORD"`
	Database string `env:"DB_PG_DATABASE" envDefault:"postgres"`
}

func (c Config) toURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.Database)
}

type DB struct {
	*base.DB
}

func NewDB(config Config, log logger.Logger) (*DB, error) {
	db, err := base.NewDB("postgres", config.toURL(), base.Dollar, migrations.GetMigrations(), log)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
