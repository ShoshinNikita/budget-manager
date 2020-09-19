package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// Income represents income entity in PostgreSQL db
type Income struct {
	tableName struct{} `pg:"incomes"`

	ID uint `pg:"id,pk"`

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	Title  string      `pg:"title"`
	Notes  string      `pg:"notes"`
	Income money.Money `pg:"income"`
}

// ToCommon converts Income to common Income structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (in Income) ToCommon(year int, month time.Month) common.Income {
	return common.Income{
		ID:     in.ID,
		Year:   year,
		Month:  month,
		Title:  in.Title,
		Notes:  in.Notes,
		Income: in.Income,
	}
}

// AddIncome adds a new income with passed params
func (db DB) AddIncome(ctx context.Context, args common.AddIncomeArgs) (id uint, err error) {
	if !db.checkMonth(ctx, args.MonthID) {
		return 0, common.ErrMonthNotExist
	}

	err = db.db.RunInTransaction(ctx, func(tx *pg.Tx) (err error) {
		income := &Income{
			MonthID: args.MonthID,
			//
			Title:  args.Title,
			Notes:  args.Notes,
			Income: args.Income,
		}
		if _, err = tx.ModelContext(ctx, income).Returning("id").Insert(); err != nil {
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
func (db DB) EditIncome(ctx context.Context, args common.EditIncomeArgs) error {
	if !db.checkIncome(ctx, args.ID) {
		return common.ErrIncomeNotExist
	}

	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		monthID, err := db.selectIncomeMonthID(tx, args.ID)
		if err != nil {
			return err
		}

		query := tx.ModelContext(ctx, (*Income)(nil)).Where("id = ?", args.ID)
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

		if args.Income != nil {
			// Recompute month only when income has been changed
			return db.recomputeAndUpdateMonth(tx, monthID)
		}
		return nil
	})
}

// RemoveIncome removes income with passed id
func (db DB) RemoveIncome(ctx context.Context, id uint) error {
	if !db.checkIncome(ctx, id) {
		return common.ErrIncomeNotExist
	}

	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		monthID, err := db.selectIncomeMonthID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.ModelContext(ctx, (*Income)(nil)).Where("id = ?", id).Delete()
		if err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectIncomeMonthID(tx *pg.Tx, id uint) (monthID uint, err error) {
	ctx := tx.Context()
	err = tx.ModelContext(ctx, (*Income)(nil)).Column("month_id").Where("id = ?", id).Select(&monthID)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select month id of Income")
	}
	return monthID, nil
}
