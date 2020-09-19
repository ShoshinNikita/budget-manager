package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// Day represents day entity in PostgreSQL db
type Day struct {
	tableName struct{} `pg:"days"`

	ID uint `pg:"id,pk"`

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	Day int `pg:"day"`
	// Saldo is a DailyBudget - Cost of all Spends multiplied by 100 (can be negative)
	Saldo  money.Money `pg:"saldo,use_zero"`
	Spends []Spend     `pg:"rel:has-many,join_fk:day_id"`
}

// ToCommon converts Day to common Day structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (d Day) ToCommon(year int, month time.Month) common.Day {
	return common.Day{
		ID:    d.ID,
		Year:  year,
		Month: month,
		Day:   d.Day,
		Saldo: d.Saldo,
		Spends: func() []common.Spend {
			spends := make([]common.Spend, 0, len(d.Spends))
			for i := range d.Spends {
				spends = append(spends, d.Spends[i].ToCommon(year, month, d.Day))
			}
			return spends
		}(),
	}
}

func (db DB) GetDay(ctx context.Context, id uint) (common.Day, error) {
	var (
		day   Day
		year  int
		month time.Month
	)
	err := db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		query := tx.ModelContext(ctx, &day).Where("id = ?", id).
			Relation("Spends", orderByID).
			Relation("Spends.Type")
		if err := query.Select(); err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return common.ErrDayNotExist
			}
			return err
		}

		// Get year and month
		query = tx.ModelContext(ctx, (*Month)(nil)).Column("year", "month").Where("id = ?", day.MonthID)
		if err := query.Select(&year, &month); err != nil {
			return errors.Wrap(err, "couldn't get year and month for Day")
		}

		return nil
	})
	if err != nil {
		return common.Day{}, err
	}

	return day.ToCommon(year, month), nil
}

func (db DB) GetDayIDByDate(ctx context.Context, year int, month int, day int) (id uint, err error) {
	monthID, err := db.GetMonthID(ctx, year, month)
	if err != nil {
		if errors.Is(err, common.ErrMonthNotExist) {
			return 0, common.ErrMonthNotExist
		}
		return 0, errors.Wrap(err, "couldn't define month id with passed year and month")
	}

	query := db.db.ModelContext(ctx, (*Day)(nil)).Column("id").Where("month_id = ? AND day = ?", monthID, day)
	err = query.Select(&id)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, common.ErrDayNotExist
		}
		return 0, err
	}

	return id, nil
}
