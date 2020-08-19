package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// AddSpend adds a new Spend
func (db DB) AddSpend(ctx context.Context, args db_common.AddSpendArgs) (id uint, err error) {
	if !db.checkDay(args.DayID) {
		return 0, db_common.ErrDayNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		spend := &Spend{
			DayID:  args.DayID,
			Title:  args.Title,
			Notes:  args.Notes,
			TypeID: args.TypeID,
			Cost:   args.Cost,
		}
		if _, err := tx.Model(spend).Returning("id").Insert(); err != nil {
			return err
		}
		id = spend.ID

		monthID, err := db.selectMonthIDByDayID(ctx, tx, args.DayID)
		if err != nil {
			return err
		}

		return db.recomputeMonth(tx, monthID)
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditSpend edits existeng Spend
func (db DB) EditSpend(ctx context.Context, args db_common.EditSpendArgs) error {
	if !db.checkSpend(args.ID) {
		return db_common.ErrSpendNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		dayID, err := db.selectSpendDayID(tx, args.ID)
		if err != nil {
			return err
		}

		query := tx.Model((*Spend)(nil)).Where("id = ?", args.ID)
		if args.Title != nil {
			query = query.Set("title = ?", *args.Title)
		}
		if args.TypeID != nil {
			if *args.TypeID == 0 {
				query = query.Set("type_id = NULL")
			} else {
				query = query.Set("type_id = ?", *args.TypeID)
			}
		}
		if args.Notes != nil {
			query = query.Set("notes = ?", *args.Notes)
		}
		if args.Cost != nil {
			query = query.Set("cost = ?", *args.Cost)
		}
		if _, err := query.Update(); err != nil {
			return err
		}

		monthID, err := db.selectMonthIDByDayID(ctx, tx, dayID)
		if err != nil {
			return err
		}
		return db.recomputeMonth(tx, monthID)
	})
}

// RemoveSpend removes Spend with passed id
func (db DB) RemoveSpend(ctx context.Context, id uint) error {
	if !db.checkSpend(id) {
		return db_common.ErrSpendNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		dayID, err := db.selectSpendDayID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.Model((*Spend)(nil)).Where("id = ?", id).Delete()
		if err != nil {
			return err
		}

		monthID, err := db.selectMonthIDByDayID(ctx, tx, dayID)
		if err != nil {
			return err
		}
		return db.recomputeMonth(tx, monthID)
	})
}

func (DB) selectSpendDayID(tx *pg.Tx, id uint) (dayID uint, err error) {
	err = tx.Model((*Spend)(nil)).Column("day_id").Where("id = ?", id).Select(&dayID)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select day id of Spend")
	}
	return dayID, nil
}

func (DB) selectMonthIDByDayID(_ context.Context, tx *pg.Tx, dayID uint) (monthID uint, err error) {
	err = tx.Model((*Day)(nil)).Column("month_id").Where("id = ?", dayID).Select(&monthID)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, db_common.ErrDayNotExist
		}
		return 0, errors.Wrap(err, "couldn't get Month which contains Day with passed id")
	}
	return monthID, nil
}
