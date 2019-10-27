package db

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

var (
	_ orm.BeforeInsertHook = (*MonthlyPayment)(nil)
	_ orm.BeforeUpdateHook = (*MonthlyPayment)(nil)
)

// MonthlyPayment contains information about monthly payments (rent, Patreon and etc.)
type MonthlyPayment struct {
	// MonthID is a foreign key to Monthes table
	MonthID uint

	ID uint `pg:",pk"`

	Title  string
	TypeID uint
	Type   *SpendType `pg:"fk:type_id"`
	Notes  string
	Cost   int64 // multiplied by 100
}

func (in *MonthlyPayment) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Check Title
	if in.Title == "" {
		return ctx, errors.New("title can't be empty")
	}

	// Skip Type

	// Skip Notes

	// Check Cost
	if in.Cost <= 0 {
		return ctx, errors.Errorf("invalid income: '%d'", in.Cost)
	}

	return ctx, nil
}

func (in *MonthlyPayment) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return in.BeforeInsert(ctx)
}

// -----------------------------------------------------------------------------

type AddMonthlyPaymentArgs struct {
	MonthID uint
	Title   string
	TypeID  uint
	Notes   string
	Cost    int64
}

// AddMonthlyPayment adds new Monthly Payment
func (db DB) AddMonthlyPayment(args AddMonthlyPaymentArgs) (id uint, err error) {
	tx, err := db.db.Begin()
	if err != nil {
		err = errors.Wrap(err, errBeginTransaction)
		db.log.Error(err)
		return 0, err
	}
	defer tx.Rollback()

	// Add Monthly Payment
	id, err = db.addMonthlyPayment(tx, args)
	if err != nil {
		err = errors.Wrap(err, "can't add new MonthlyPayment")
		db.log.Error(err)
		return 0, err
	}

	// Recompute month budget
	err = db.recomputeMonth(tx, args.MonthID)
	if err != nil {
		err = errors.Wrap(err, errRecomputeBudget)
		db.log.Error(err)
		return 0, err
	}

	// Commit changes
	err = tx.Commit()
	if err != nil {
		err = errors.Wrap(err, errCommitChanges)
		db.log.Error(err)
		return 0, err
	}

	return id, nil
}

func (_ DB) addMonthlyPayment(tx *pg.Tx, args AddMonthlyPaymentArgs) (id uint, err error) {
	mp := &MonthlyPayment{
		MonthID: args.MonthID,
		Title:   args.Title,
		Notes:   args.Notes,
		TypeID:  args.TypeID,
		Cost:    args.Cost,
	}

	err = tx.Insert(mp)
	if err != nil {
		return 0, errors.Wrap(err, "can't insert MonthlyPayment")
	}

	return mp.ID, nil
}

type EditMonthlyPaymentArgs struct {
	ID uint

	Title  *string
	TypeID *uint
	Notes  *string
	Cost   *int64
}

// EditMonthlyPayment modifies existing Monthly Payment
func (db DB) EditMonthlyPayment(args EditMonthlyPaymentArgs) error {
	tx, err := db.db.Begin()
	if err != nil {
		return errors.Wrap(err, errBeginTransaction)
	}
	defer tx.Rollback()

	mp := &MonthlyPayment{ID: args.ID}
	err = tx.Select(mp)
	if err != nil {
		if err == pg.ErrNoRows {
			err = errors.New("wrong id")
		} else {
			errors.Wrap(err, "can't get Monthly Payment with passed id")
		}

		db.log.Error(err)
		return err
	}

	// Edit Monthly Payment
	err = db.editMonthlyPayment(tx, mp, args)
	if err != nil {
		err = errors.Wrapf(err, "can't edit Monthly Payment with id '%d'", args.ID)
		db.log.Error(err)
		return err
	}

	// Recompute month budget
	err = db.recomputeMonth(tx, mp.MonthID)
	if err != nil {
		err = errors.Wrap(err, errRecomputeBudget)
		db.log.Error(err)
		return err
	}

	// Commit changes
	err = tx.Commit()
	if err != nil {
		err = errors.Wrap(err, errCommitChanges)
		db.log.Error(err)
		return err
	}

	return nil
}

func (_ DB) editMonthlyPayment(tx *pg.Tx, mp *MonthlyPayment, args EditMonthlyPaymentArgs) error {
	if args.Title != nil {
		mp.Title = *args.Title
	}
	if args.TypeID != nil {
		mp.TypeID = *args.TypeID
	}
	if args.Notes != nil {
		mp.Notes = *args.Notes
	}
	if args.Cost != nil {
		mp.Cost = *args.Cost
	}

	err := tx.Update(mp)
	return errors.Wrap(err, "can't update Monthly Payment")
}

func (db DB) RemoveMonthlyPayment(id uint) error {
	tx, err := db.db.Begin()
	if err != nil {
		return errors.Wrap(err, errBeginTransaction)
	}
	defer tx.Rollback()

	// Remove Monthly Payment
	mp := &MonthlyPayment{ID: id}
	err = db.db.Delete(mp)
	if err != nil {
		err = errors.Wrapf(err, "can't remove Monthly Payment with id '%d'", id)
		db.log.Error(err)
		return err
	}

	// Recompute month budget
	err = db.recomputeMonth(tx, mp.MonthID)
	if err != nil {
		err = errors.Wrap(err, errRecomputeBudget)
		db.log.Error(err)
		return err
	}

	// Commit changes
	err = tx.Commit()
	if err != nil {
		err = errors.Wrap(err, errCommitChanges)
		db.log.Error(err)
		return err
	}

	return nil
}
