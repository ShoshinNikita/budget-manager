package db

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/pkg/errors"
)

var (
	_ orm.BeforeInsertHook = (*Income)(nil)
	_ orm.BeforeUpdateHook = (*Income)(nil)
)

// Income contains information about incomes (salary, gifts and etc.)
type Income struct {
	// MonthID is a foreign key to Months table
	MonthID uint

	ID uint `pg:",pk"`

	Title  string
	Notes  string
	Income int64 // multiplied by 100
}

func (in *Income) BeforeInsert(ctx context.Context) (context.Context, error) {
	// Check Title
	if in.Title == "" {
		return ctx, errors.New("title can't be empty")
	}

	// Skip Notes

	// Check Income
	if in.Income <= 0 {
		return ctx, errors.Errorf("invalid income: '%d'", in.Income)
	}

	return ctx, nil
}

func (in *Income) BeforeUpdate(ctx context.Context) (context.Context, error) {
	// Can use BeforeInsert hook
	return in.BeforeInsert(ctx)
}

// -----------------------------------------------------------------------------

type AddIncomeArgs struct {
	MonthID uint
	Title   string
	Notes   string
	Income  int64
}

// AddIncome adds a new income with passed params
func (db DB) AddIncome(args AddIncomeArgs) (incomeID uint, err error) {
	if !db.checkMonth(args.MonthID) {
		return 0, ErrMonthNotExist
	}

	tx, err := db.db.Begin()
	if err != nil {
		err = errors.Wrap(err, errBeginTransaction)
		db.log.Error(err)
		return 0, err
	}
	defer tx.Rollback()

	// Add a new Income
	id, err := db.addIncome(tx, args)
	if err != nil {
		err = errors.Wrap(err, "can't add income")
		db.log.Error(err)
		return 0, err
	}

	// Recompute Month budget
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

func (_ DB) addIncome(tx *pg.Tx, args AddIncomeArgs) (incomeID uint, err error) {
	in := &Income{
		MonthID: args.MonthID,

		Title:  args.Title,
		Notes:  args.Notes,
		Income: args.Income,
	}

	err = tx.Insert(in)
	if err != nil {
		err = errors.Wrap(err, "can't insert a new row")
		return 0, err
	}

	return in.ID, nil
}

type EditIncomeArgs struct {
	ID     uint
	Title  *string
	Notes  *string
	Income *int64
}

// EditIncome edits income with passed id, nil args are ignored
func (db DB) EditIncome(args EditIncomeArgs) error {
	if !db.checkIncome(args.ID) {
		return ErrIncomeNotExist
	}

	tx, err := db.db.Begin()
	if err != nil {
		err = errors.Wrap(err, errBeginTransaction)
		db.log.Error(err)
		return err
	}
	defer tx.Rollback()

	// Select Income.
	in := &Income{ID: args.ID}
	err = tx.Select(in)
	if err != nil {
		if err == pg.ErrNoRows {
			err = errors.New("wrong id")
		} else {
			err = errors.Wrapf(err, "can't select Income with passed id '%d'", args.ID)
		}
		db.log.Error(err)
		return err
	}

	// Edit Income
	err = db.editIncome(tx, in, args)
	if err != nil {
		err = errors.Wrap(err, "can't edit Income")
		db.log.Error(err)
		return err
	}

	// Recompute Month budget
	err = db.recomputeMonth(tx, in.MonthID)
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

func (_ DB) editIncome(tx *pg.Tx, in *Income, args EditIncomeArgs) error {
	if args.Title != nil {
		in.Title = *args.Title
	}
	if args.Notes != nil {
		in.Notes = *args.Notes
	}
	if args.Income != nil {
		in.Income = *args.Income
	}

	err := tx.Update(in)
	if err != nil {
		return errors.Wrap(err, "can't update Income")
	}

	return nil
}

// RemoveIncome removes income with passed id
func (db DB) RemoveIncome(id uint) error {
	if !db.checkIncome(id) {
		return ErrIncomeNotExist
	}

	tx, err := db.db.Begin()
	if err != nil {
		err = errors.Wrap(err, errBeginTransaction)
		return err
	}
	defer tx.Rollback()

	monthID, err := func() (uint, error) {
		in := &Income{ID: id}
		err = tx.Model(in).Column("month_id").WherePK().Select(in)
		if err != nil {
			return 0, err
		}

		return in.MonthID, nil
	}()
	if err != nil {
		err = errors.Wrap(err, "can't select Income with passed id")
		db.log.Error(err)
		return err
	}

	// Remove income
	err = db.removeIncome(tx, id)
	if err != nil {
		err = errors.Wrap(err, "can't remove Income with passed id")
		db.log.Error(err)
		return err
	}

	// Recompute Month budget
	err = db.recomputeMonth(tx, monthID)
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

func (_ DB) removeIncome(tx *pg.Tx, id uint) error {
	in := &Income{ID: id}
	return tx.Delete(in)
}
