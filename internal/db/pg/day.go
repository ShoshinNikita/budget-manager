package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

func (db DB) GetDay(_ context.Context, id uint) (*db_common.Day, error) {
	var (
		day   Day
		year  int
		month time.Month
	)
	err := db.db.RunInTransaction(func(tx *pg.Tx) error {
		query := tx.Model(&day).Where("id = ?", id).
			Relation("Spends", orderByID).
			Relation("Spends.Type")
		if err := query.Select(); err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return db_common.ErrDayNotExist
			}
			return err
		}

		// Get year and month
		query = tx.Model((*Month)(nil)).Column("year", "month").Where("id = ?", day.MonthID)
		if err := query.Select(&year, &month); err != nil {
			return errors.Wrap(err, "couldn't get year and month for Day")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return day.ToCommon(year, month), nil
}

func (db DB) GetDayIDByDate(ctx context.Context, year int, month int, day int) (id uint, err error) {
	monthID, err := db.GetMonthID(ctx, year, month)
	if err != nil {
		if errors.Is(err, db_common.ErrMonthNotExist) {
			return 0, db_common.ErrMonthNotExist
		}
		return 0, errors.Wrap(err, "couldn't define month id with passed year and month")
	}

	query := db.db.Model((*Day)(nil)).Column("id").Where("month_id = ? AND day = ?", monthID, day)
	if err = query.Select(&id); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, db_common.ErrDayNotExist
		}
		return 0, err
	}

	return id, nil
}
