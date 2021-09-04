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

type MonthlyPayment struct {
	ID      uint        `db:"id"`
	MonthID uint        `db:"month_id"`
	Title   string      `db:"title"`
	TypeID  types.Uint  `db:"type_id"`
	Notes   string      `db:"notes"`
	Cost    money.Money `db:"cost"`

	Type *SpendType `db:"type"`
}

// ToCommon converts MonthlyPayment to common MonthlyPayment structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (mp MonthlyPayment) ToCommon(year int, month time.Month) common.MonthlyPayment {
	return common.MonthlyPayment{
		ID:    mp.ID,
		Year:  year,
		Month: month,
		Title: mp.Title,
		Type:  mp.Type.ToCommon(),
		Notes: mp.Notes,
		Cost:  mp.Cost,
	}
}

// AddMonthlyPayment adds new Monthly Payment
func (db DB) AddMonthlyPayment(ctx context.Context, args common.AddMonthlyPaymentArgs) (id uint, err error) {
	err = db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkMonth(tx, args.MonthID) {
			return common.ErrMonthNotExist
		}
		if args.TypeID != 0 && !checkSpendType(tx, args.TypeID) {
			return common.ErrSpendTypeNotExist
		}

		err = tx.Get(
			&id,
			`INSERT INTO monthly_payments(month_id, title, notes, type_id, cost) VALUES(?, ?, ?, ?, ?) RETURNING id`,
			args.MonthID, args.Title, args.Notes, types.Uint(args.TypeID), args.Cost,
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

// EditMonthlyPayment modifies existing Monthly Payment
func (db DB) EditMonthlyPayment(ctx context.Context, args common.EditMonthlyPaymentArgs) error {
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkMonthlyPayment(tx, args.ID) {
			return common.ErrMonthlyPaymentNotExist
		}
		if args.TypeID != nil && *args.TypeID != 0 && !checkSpendType(tx, *args.TypeID) {
			return common.ErrSpendTypeNotExist
		}

		monthID, err := db.selectMonthlyPaymentMonthID(tx, args.ID)
		if err != nil {
			return err
		}

		query := squirrel.Update("monthly_payments").Where("id = ?", args.ID)
		if args.Title != nil {
			query = query.Set("title", *args.Title)
		}
		if args.TypeID != nil {
			if *args.TypeID == 0 {
				query = query.Set("type_id", nil)
			} else {
				query = query.Set("type_id", *args.TypeID)
			}
		}
		if args.Notes != nil {
			query = query.Set("notes", *args.Notes)
		}
		if args.Cost != nil {
			query = query.Set("cost", *args.Cost)
		}
		if _, err := tx.ExecQuery(query); err != nil {
			return err
		}

		if args.Cost != nil {
			// Recompute month only when cost has been changed
			return db.recomputeAndUpdateMonth(tx, monthID)
		}
		return nil
	})
}

// RemoveMonthlyPayment removes Monthly Payment with passed id
func (db DB) RemoveMonthlyPayment(ctx context.Context, id uint) error {
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkMonthlyPayment(tx, id) {
			return common.ErrMonthlyPaymentNotExist
		}

		monthID, err := db.selectMonthlyPaymentMonthID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`DELETE FROM monthly_payments WHERE id = ?`, id)
		if err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectMonthlyPaymentMonthID(tx *sqlx.Tx, id uint) (monthID uint, err error) {
	err = tx.Get(&monthID, `SELECT month_id FROM monthly_payments WHERE id = ?`, id)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select month id of Monthly Payment")
	}
	return monthID, nil
}
