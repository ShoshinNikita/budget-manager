package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// MonthlyPayment represents monthly payment entity in PostgreSQL db
type MonthlyPayment struct {
	tableName struct{} `pg:"monthly_payments"`

	ID uint `pg:"id,pk"`

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	Title  string      `pg:"title"`
	TypeID uint        `pg:"type_id"`
	Type   *SpendType  `pg:"rel:has-one,fk:type_id"`
	Notes  string      `pg:"notes"`
	Cost   money.Money `pg:"cost"`
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
	if !db.checkMonth(ctx, args.MonthID) {
		return 0, common.ErrMonthNotExist
	}

	err = db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		mp := &MonthlyPayment{
			MonthID: args.MonthID,
			Title:   args.Title,
			Notes:   args.Notes,
			TypeID:  args.TypeID,
			Cost:    args.Cost,
		}
		if _, err := tx.ModelContext(ctx, mp).Returning("id").Insert(); err != nil {
			return err
		}
		id = mp.ID

		return db.recomputeAndUpdateMonth(tx, args.MonthID)
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditMonthlyPayment modifies existing Monthly Payment
func (db DB) EditMonthlyPayment(ctx context.Context, args common.EditMonthlyPaymentArgs) error {
	if !db.checkMonthlyPayment(ctx, args.ID) {
		return common.ErrMonthlyPaymentNotExist
	}

	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		monthID, err := db.selectMonthlyPaymentMonthID(tx, args.ID)
		if err != nil {
			return err
		}

		query := tx.ModelContext(ctx, (*MonthlyPayment)(nil)).Where("id = ?", args.ID)
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

		if args.Cost != nil {
			// Recompute month only when cost has been changed
			return db.recomputeAndUpdateMonth(tx, monthID)
		}
		return nil
	})
}

// RemoveMonthlyPayment removes Monthly Payment with passed id
func (db DB) RemoveMonthlyPayment(ctx context.Context, id uint) error {
	if !db.checkMonthlyPayment(ctx, id) {
		return common.ErrMonthlyPaymentNotExist
	}

	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		monthID, err := db.selectMonthlyPaymentMonthID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.ModelContext(ctx, (*MonthlyPayment)(nil)).Where("id = ?", id).Delete()
		if err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectMonthlyPaymentMonthID(tx *pg.Tx, id uint) (monthID uint, err error) {
	ctx := tx.Context()
	err = tx.ModelContext(ctx, (*MonthlyPayment)(nil)).Column("month_id").Where("id = ?", id).Select(&monthID)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select month id of Monthly Payment")
	}
	return monthID, nil
}
