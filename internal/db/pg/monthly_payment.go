package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	db_common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// MonthlyPayment represents monthly payment entity in PostgreSQL db
type MonthlyPayment struct {
	tableName struct{} `pg:"monthly_payments"` // nolint:structcheck,unused

	// MonthID is a foreign key to 'months' table
	MonthID uint `pg:"month_id"`

	ID uint `pg:"id,pk"`

	Title  string      `pg:"title"`
	TypeID uint        `pg:"type_id"`
	Type   *SpendType  `pg:"fk:type_id"`
	Notes  string      `pg:"notes"`
	Cost   money.Money `pg:"cost"`
}

// ToCommon converts MonthlyPayment to common MonthlyPayment structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (mp *MonthlyPayment) ToCommon(year int, month time.Month) *db_common.MonthlyPayment {
	if mp == nil {
		return nil
	}
	return &db_common.MonthlyPayment{
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
func (db DB) AddMonthlyPayment(_ context.Context, args db_common.AddMonthlyPaymentArgs) (id uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, db_common.ErrMonthNotExist
	}

	err = db.db.RunInTransaction(func(tx *pg.Tx) error {
		mp := &MonthlyPayment{
			MonthID: args.MonthID,
			Title:   args.Title,
			Notes:   args.Notes,
			TypeID:  args.TypeID,
			Cost:    args.Cost,
		}
		if _, err := tx.Model(mp).Returning("id").Insert(); err != nil {
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
func (db DB) EditMonthlyPayment(_ context.Context, args db_common.EditMonthlyPaymentArgs) error {
	if !db.checkMonthlyPayment(args.ID) {
		return db_common.ErrMonthlyPaymentNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		monthID, err := db.selectMonthlyPaymentMonthID(tx, args.ID)
		if err != nil {
			return err
		}

		query := tx.Model((*MonthlyPayment)(nil)).Where("id = ?", args.ID)
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

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

// RemoveMonthlyPayment removes Monthly Payment with passed id
func (db DB) RemoveMonthlyPayment(_ context.Context, id uint) error {
	if !db.checkMonthlyPayment(id) {
		return db_common.ErrMonthlyPaymentNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		monthID, err := db.selectMonthlyPaymentMonthID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.Model((*MonthlyPayment)(nil)).Where("id = ?", id).Delete()
		if err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectMonthlyPaymentMonthID(tx *pg.Tx, id uint) (monthID uint, err error) {
	err = tx.Model((*MonthlyPayment)(nil)).Column("month_id").Where("id = ?", id).Select(&monthID)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select month id of Monthly Payment")
	}
	return monthID, nil
}
