package base

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/types"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type Income struct {
	ID      uint         `db:"id"`
	MonthID uint         `db:"month_id"`
	Title   string       `db:"title"`
	Notes   types.String `db:"notes"`
	Income  money.Money  `db:"income"`
}

// ToCommon converts Income to common Income structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (in Income) ToCommon(year int, month time.Month) common.Income {
	return common.Income{
		ID:     in.ID,
		Year:   year,
		Month:  month,
		Title:  in.Title,
		Notes:  string(in.Notes),
		Income: in.Income,
	}
}

// AddIncome adds a new income with passed params
func (db DB) AddIncome(ctx context.Context, args common.AddIncomeArgs) (id uint, err error) {
	err = db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) (err error) {
		if !checkMonth(tx, args.MonthID) {
			return common.ErrMonthNotExist
		}

		err = tx.Get(
			&id,
			`INSERT INTO incomes(month_id, title, notes, income) VALUES(?, ?, ?, ?) RETURNING id`,
			args.MonthID, args.Title, args.Notes, args.Income,
		)
		if err != nil {
			return err
		}
		return db.recomputeAndUpdateMonth(tx, args.MonthID)
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditIncome edits income with passed id, nil args are ignored
func (db DB) EditIncome(ctx context.Context, args common.EditIncomeArgs) error {
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkIncome(tx, args.ID) {
			return common.ErrIncomeNotExist
		}

		monthID, err := db.selectIncomeMonthID(tx, args.ID)
		if err != nil {
			return err
		}

		query := squirrel.Update("incomes").Where("id = ?", args.ID)
		if args.Title != nil {
			query = query.Set("title", *args.Title)
		}
		if args.Notes != nil {
			query = query.Set("notes", *args.Notes)
		}
		if args.Income != nil {
			query = query.Set("income", *args.Income)
		}
		if _, err := tx.ExecQuery(query); err != nil {
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
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkIncome(tx, id) {
			return common.ErrIncomeNotExist
		}

		monthID, err := db.selectIncomeMonthID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`DELETE FROM incomes WHERE id = ?`, id)
		if err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectIncomeMonthID(tx *sqlx.Tx, id uint) (monthID uint, err error) {
	err = tx.Get(&monthID, `SELECT month_id FROM incomes WHERE id = ?`, id)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select month id of Income")
	}
	return monthID, nil
}
