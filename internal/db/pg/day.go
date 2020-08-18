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
		day   *Day
		year  int
		month time.Month
	)
	err := db.db.RunInTransaction(func(tx *pg.Tx) error {
		day = &Day{ID: id}
		err := tx.Model(day).
			Relation("Spends", orderByID).
			Relation("Spends.Type").
			WherePK().Select()
		if err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				return db_common.ErrDayNotExist
			}
			return err
		}

		// Get year and month
		err = tx.Model((*Month)(nil)).
			Column("year", "month").
			Where("id = ?", day.MonthID).
			Select(&year, &month)
		if err != nil {
			return errors.Wrap(err, "couldn't get year and month for Day")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return day.ToCommon(year, month), nil
}

func (db DB) GetDayIDByDate(ctx context.Context, year int, month int, day int) (uint, error) {
	monthID, err := db.GetMonthID(ctx, year, month)
	if err != nil {
		if errors.Is(err, db_common.ErrMonthNotExist) {
			return 0, db_common.ErrMonthNotExist
		}
		return 0, errors.Wrap(err, "couldn't define month id with passed year and month")
	}

	d := &Day{}
	err = db.db.Model(d).
		Column("id").
		Where("month_id = ? AND day = ?", monthID, day).
		Select()
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, db_common.ErrDayNotExist
		}
		return 0, err
	}

	return d.ID, nil
}
