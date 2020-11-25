package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

// Spend represents spend entity in PostgreSQL db
type Spend struct {
	tableName struct{} `pg:"spends"`

	// DayID is a foreign key to 'days' table
	DayID uint `pg:"day_id"`

	ID uint `pg:"id,pk"`

	Title  string      `pg:"title"`
	TypeID uint        `pg:"type_id"`
	Type   *SpendType  `pg:"rel:has-one,fk:type_id"`
	Notes  string      `pg:"notes"`
	Cost   money.Money `pg:"cost,use_zero"`
}

// ToCommon converts Spend to common Spend structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
func (s Spend) ToCommon(year int, month time.Month, day int) common.Spend {
	return common.Spend{
		ID:    s.ID,
		Year:  year,
		Month: month,
		Day:   day,
		Title: s.Title,
		Type:  s.Type.ToCommon(),
		Notes: s.Notes,
		Cost:  s.Cost,
	}
}

// AddSpend adds a new Spend
func (db DB) AddSpend(ctx context.Context, args common.AddSpendArgs) (id uint, err error) {
	if !db.checkDay(ctx, args.DayID) {
		return 0, common.ErrDayNotExist
	}

	err = db.db.RunInTransaction(ctx, func(tx *pg.Tx) (err error) {
		spend := &Spend{
			DayID:  args.DayID,
			Title:  args.Title,
			Notes:  args.Notes,
			TypeID: args.TypeID,
			Cost:   args.Cost,
		}
		if _, err := tx.ModelContext(ctx, spend).Returning("id").Insert(); err != nil {
			return err
		}
		id = spend.ID

		monthID, err := db.selectMonthIDByDayID(ctx, tx, args.DayID)
		if err != nil {
			return err
		}

		return db.recomputeAndUpdateMonth(tx, monthID)
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditSpend edits existeng Spend
func (db DB) EditSpend(ctx context.Context, args common.EditSpendArgs) error {
	if !db.checkSpend(ctx, args.ID) {
		return common.ErrSpendNotExist
	}

	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		dayID, err := db.selectSpendDayID(ctx, tx, args.ID)
		if err != nil {
			return err
		}

		query := tx.ModelContext(ctx, (*Spend)(nil)).Where("id = ?", args.ID)
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
			monthID, err := db.selectMonthIDByDayID(ctx, tx, dayID)
			if err != nil {
				return err
			}
			return db.recomputeAndUpdateMonth(tx, monthID)
		}
		return nil
	})
}

// RemoveSpend removes Spend with passed id
func (db DB) RemoveSpend(ctx context.Context, id uint) error {
	if !db.checkSpend(ctx, id) {
		return common.ErrSpendNotExist
	}

	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		dayID, err := db.selectSpendDayID(ctx, tx, id)
		if err != nil {
			return err
		}

		_, err = tx.ModelContext(ctx, (*Spend)(nil)).Where("id = ?", id).Delete()
		if err != nil {
			return err
		}

		monthID, err := db.selectMonthIDByDayID(ctx, tx, dayID)
		if err != nil {
			return err
		}
		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectSpendDayID(ctx context.Context, tx *pg.Tx, id uint) (dayID uint, err error) {
	query := tx.ModelContext(ctx, (*Spend)(nil)).Column("day_id").Where("id = ?", id)
	err = query.Select(&dayID)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select day id of Spend")
	}
	return dayID, nil
}

func (DB) selectMonthIDByDayID(ctx context.Context, tx *pg.Tx, dayID uint) (monthID uint, err error) {
	query := tx.ModelContext(ctx, (*Day)(nil)).Column("month_id").Where("id = ?", dayID)
	err = query.Select(&monthID)
	if err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, common.ErrDayNotExist
		}
		return 0, errors.Wrap(err, "couldn't get Month which contains Day with passed id")
	}
	return monthID, nil
}
