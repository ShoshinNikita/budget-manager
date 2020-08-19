package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// AddIncome adds a new income with passed params
func (db DB) AddIncome(_ context.Context, args db_common.AddIncomeArgs) (id uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, db_common.ErrMonthNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) (err error) {
		income := &Income{
			MonthID: args.MonthID,
			//
			Title:  args.Title,
			Notes:  args.Notes,
			Income: args.Income,
		}
		if _, err = tx.Model(income).Returning("id").Insert(); err != nil {
			return err
		}
		id = income.ID

		return db.recomputeAndUpdateMonth(tx, args.MonthID)
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditIncome edits income with passed id, nil args are ignored
func (db DB) EditIncome(_ context.Context, args db_common.EditIncomeArgs) error {
	if !db.checkIncome(args.ID) {
		return db_common.ErrIncomeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		monthID, err := db.selectIncomeMonthID(tx, args.ID)
		if err != nil {
			return err
		}

		query := tx.Model((*Income)(nil)).Where("id = ?", args.ID)
		if args.Title != nil {
			query = query.Set("title = ?", *args.Title)
		}
		if args.Notes != nil {
			query = query.Set("notes = ?", *args.Notes)
		}
		if args.Income != nil {
			query = query.Set("income = ?", *args.Income)
		}
		if _, err := query.Update(); err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

// RemoveIncome removes income with passed id
func (db DB) RemoveIncome(_ context.Context, id uint) error {
	if !db.checkIncome(id) {
		return db_common.ErrIncomeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		monthID, err := db.selectIncomeMonthID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.Model((*Income)(nil)).Where("id = ?", id).Delete()
		if err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectIncomeMonthID(tx *pg.Tx, id uint) (monthID uint, err error) {
	err = tx.Model((*Income)(nil)).Column("month_id").Where("id = ?", id).Select(&monthID)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select month id of Income")
	}
	return monthID, nil
}
