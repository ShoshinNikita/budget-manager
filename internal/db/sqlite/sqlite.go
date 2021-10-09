package sqlite

//nolint:gci
import (
	_ "github.com/mattn/go-sqlite3" // register SQLite driver

	"github.com/ShoshinNikita/budget-manager/internal/db/base"
	"github.com/ShoshinNikita/budget-manager/internal/db/sqlite/migrations"
	"github.com/ShoshinNikita/budget-manager/internal/logger"
)

type DB struct {
	*base.DB
}

type Config struct {
	Path string
}

func NewDB(config Config, log logger.Logger) (*DB, error) {
	db, err := base.NewDB("sqlite3", config.Path, base.Question, migrations.GetMigrations(), log)
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
