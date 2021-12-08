package pg

//nolint:gci
import (
	"context"
	"fmt"
	"os/exec"

	_ "github.com/lib/pq" // register PostgreSQL driver

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/base"
	"github.com/ShoshinNikita/budget-manager/internal/db/pg/migrations"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (c Config) toURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, c.Password, c.Host, c.Port, c.Database)
}

type DB struct {
	*base.DB

	cfg Config
}

func NewDB(config Config, log logger.Logger) (*DB, error) {
	db, err := base.NewDB("postgres", config.toURL(), base.Dollar, migrations.GetMigrations(), log)
	if err != nil {
		return nil, err
	}
	return &DB{db, config}, nil
}

func (db *DB) GetType() common.Type {
	return common.Postgres
}

func (db *DB) Backup(ctx context.Context) ([]byte, error) {
	args := []string{
		"--dbname=" + db.cfg.toURL(),
	}
	return exec.CommandContext(ctx, "pg_dump", args...).Output()
}
