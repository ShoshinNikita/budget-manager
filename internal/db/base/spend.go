package base

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/types"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

type Spend struct {
	ID     uint         `db:"id"`
	DayID  uint         `db:"day_id"`
	Title  string       `db:"title"`
	TypeID types.Uint   `db:"type_id"`
	Notes  types.String `db:"notes"`
	Cost   money.Money  `db:"cost"`

	Type *SpendType `db:"type"`
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
		Notes: string(s.Notes),
		Cost:  s.Cost,
	}
}

// AddSpend adds a new Spend
func (db DB) AddSpend(ctx context.Context, args common.AddSpendArgs) (id uint, err error) {
	err = db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) (err error) {
		if !checkDay(tx, args.DayID) {
			return common.ErrDayNotExist
		}
		if args.TypeID != 0 && !checkSpendType(tx, args.TypeID) {
			return common.ErrSpendTypeNotExist
		}

		err = tx.Get(
			&id,
			`INSERT INTO spends(day_id, title, notes, type_id, cost) VALUES(?, ?, ?, ?, ?) RETURNING id`,
			args.DayID, args.Title, args.Notes, types.Uint(args.TypeID), args.Cost,
		)
		if err != nil {
			return err
		}

		monthID, err := db.selectMonthIDByDayID(tx, args.DayID)
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
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkSpend(tx, args.ID) {
			return common.ErrSpendNotExist
		}
		if args.TypeID != nil && *args.TypeID != 0 && !checkSpendType(tx, *args.TypeID) {
			return common.ErrSpendTypeNotExist
		}

		dayID, err := db.selectSpendDayID(tx, args.ID)
		if err != nil {
			return err
		}

		query := squirrel.Update("spends").Where("id = ?", args.ID)
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
			monthID, err := db.selectMonthIDByDayID(tx, dayID)
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
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkSpend(tx, id) {
			return common.ErrSpendNotExist
		}

		dayID, err := db.selectSpendDayID(tx, id)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`DELETE FROM spends WHERE id = ?`, id)
		if err != nil {
			return err
		}

		monthID, err := db.selectMonthIDByDayID(tx, dayID)
		if err != nil {
			return err
		}
		return db.recomputeAndUpdateMonth(tx, monthID)
	})
}

func (DB) selectSpendDayID(tx *sqlx.Tx, id uint) (dayID uint, err error) {
	err = tx.Get(&dayID, `SELECT day_id FROM spends WHERE id = ?`, id)
	if err != nil {
		return 0, errors.Wrap(err, "couldn't select day id of Spend")
	}
	return dayID, nil
}

func (DB) selectMonthIDByDayID(tx *sqlx.Tx, dayID uint) (monthID uint, err error) {
	err = tx.Get(&monthID, `SELECT month_id FROM days WHERE id = ?`, dayID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, common.ErrDayNotExist
		}
		return 0, errors.Wrap(err, "couldn't get Month which contains Day with passed id")
	}
	return monthID, nil
}
