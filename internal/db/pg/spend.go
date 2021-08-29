package pg

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
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

func (db DB) GetAllSpends(ctx context.Context) ([]common.Spend, error) {
	var spends []struct {
		Spend

		Year  int        `pg:"year"`
		Month time.Month `pg:"month"`
		Day   int        `pg:"day"`
	}
	_, err := db.db.Query(&spends, `
			SELECT
				spends.*,
				months.year,
				months.month,
				days.day,
				spend_types.id AS type__id,
				spend_types.name AS type__name,
				spend_types.parent_id AS type__parent_id
			FROM spends
			LEFT JOIN days ON days.id = spends.day_id
			LEFT JOIN months ON months.id = days.month_id
			LEFT JOIN spend_types AS spend_types ON spend_types.id = spends.type_id
			ORDER BY spends.id ASC`,
	)
	if err != nil {
		return nil, err
	}

	res := make([]common.Spend, 0, len(spends))
	for _, s := range spends {
		res = append(res, s.ToCommon(s.Year, s.Month, s.Day))
	}
	return res, nil
}

// AddSpend adds a new Spend
func (db DB) AddSpend(ctx context.Context, args common.AddSpendArgs) (id uint, err error) {
	err = db.db.RunInTransaction(ctx, func(tx *pg.Tx) (err error) {
		if !checkDay(ctx, tx, args.DayID) {
			return common.ErrDayNotExist
		}
		if args.TypeID != 0 && !checkSpendType(ctx, tx, args.TypeID) {
			return common.ErrSpendTypeNotExist
		}

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
	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		if !checkSpend(ctx, tx, args.ID) {
			return common.ErrSpendNotExist
		}
		if args.TypeID != nil && *args.TypeID != 0 && !checkSpendType(ctx, tx, *args.TypeID) {
			return common.ErrSpendTypeNotExist
		}

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
	return db.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		if !checkSpend(ctx, tx, id) {
			return common.ErrSpendNotExist
		}

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
