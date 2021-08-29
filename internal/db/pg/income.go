package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
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

func (db DB) GetAllIncomes(ctx context.Context) ([]common.Income, error) {
	var incomes []struct {
		Income

		Year  int        `pg:"year"`
		Month time.Month `pg:"month"`
	}
	_, err := db.db.Query(&incomes, `
		SELECT
			incomes.*,
			months.year,
			months.month
		FROM incomes
		LEFT JOIN months ON months.id = incomes.month_id
		ORDER BY incomes.id ASC`,
	)
	if err != nil {
		return nil, err
	}

	res := make([]common.Income, 0, len(incomes))
	for _, i := range incomes {
		res = append(res, i.ToCommon(i.Year, i.Month))
	}
	return res, nil
}

// AddIncome adds a new income with passed params
func (db DB) AddIncome(ctx context.Context, args common.AddIncomeArgs) (id uint, err error) {
	err = db.db.RunInTransaction(ctx, func(tx *pg.Tx) (err error) {
		if !checkMonth(ctx, tx, args.MonthID) {
			return common.ErrMonthNotExist
		}

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
	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		if !checkIncome(ctx, tx, args.ID) {
			return common.ErrIncomeNotExist
		}

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
	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		if !checkIncome(ctx, tx, id) {
			return common.ErrIncomeNotExist
		}

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
