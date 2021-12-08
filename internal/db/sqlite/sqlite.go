package sqlite

//nolint:gci
import (
	"context"
	"os/exec"

	_ "github.com/mattn/go-sqlite3" // register SQLite driver

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/base"
	"github.com/ShoshinNikita/budget-manager/internal/db/sqlite/migrations"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
)

type DB struct {
	*base.DB

	cfg Config
}

type Config struct {
	Path string
}

func NewDB(config Config, log logger.Logger) (*DB, error) {
	db, err := base.NewDB("sqlite3", config.Path, base.Question, migrations.GetMigrations(), log)
	if err != nil {
		return nil, err
	}
	return &DB{db, config}, nil
}

func (db *DB) GetType() common.Type {
	return common.Sqlite3
}

//nolint:gosec
func (db *DB) Backup(ctx context.Context) ([]byte, error) {
	return exec.CommandContext(ctx, "sqlite3", db.cfg.Path, ".dump").Output()
}
