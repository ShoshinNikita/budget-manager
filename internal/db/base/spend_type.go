package base

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/internal/sqlx"
	"github.com/ShoshinNikita/budget-manager/internal/db/base/types"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// SpendType represents spend type entity in PostgreSQL db
type SpendType struct {
	ID       types.Uint   `db:"id"`
	Name     types.String `db:"name"`
	ParentID types.Uint   `db:"parent_id"`
}

// ToCommon converts SpendType to common SpendType structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
//
// We return a pointer instead of a value unlike other 'ToCommon' methods because Spend Type can be optional
func (s *SpendType) ToCommon() *common.SpendType {
	if s == nil {
		return nil
	}
	if s.ID == 0 {
		// Valid Spend Type must have have id != 0
		return nil
	}

	return &common.SpendType{
		ID:       uint(s.ID),
		Name:     string(s.Name),
		ParentID: uint(s.ParentID),
	}
}

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(ctx context.Context, id uint) (common.SpendType, error) {
	var spendType SpendType
	err := db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		return tx.Get(&spendType, `SELECT * from spend_types WHERE id = ?`, id)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = common.ErrSpendTypeNotExist
		}
		return common.SpendType{}, err
	}

	return *spendType.ToCommon(), nil
}

// GetSpendTypes returns all Spend Types
func (db DB) GetSpendTypes(ctx context.Context) ([]common.SpendType, error) {
	var spendTypes []SpendType
	err := db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		return tx.Select(&spendTypes, `SELECT * from spend_types ORDER BY id ASC`)
	})
	if err != nil {
		return nil, err
	}

	res := make([]common.SpendType, 0, len(spendTypes))
	for i := range spendTypes {
		res = append(res, *spendTypes[i].ToCommon())
	}
	return res, nil
}

// AddSpendType adds new Spend Type
func (db DB) AddSpendType(ctx context.Context, args common.AddSpendTypeArgs) (id uint, err error) {
	err = db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		return tx.Get(
			&id,
			`INSERT INTO spend_types(name, parent_id) VALUES(?, ?) RETURNING id`,
			args.Name, types.Uint(args.ParentID),
		)
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditSpendType modifies existing Spend Type
func (db DB) EditSpendType(ctx context.Context, args common.EditSpendTypeArgs) error {
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkSpendType(tx, args.ID) {
			return common.ErrSpendTypeNotExist
		}

		query := squirrel.Update("spend_types").Where("id = ?", args.ID)
		if args.Name != nil {
			query = query.Set("name", *args.Name)
		}
		if args.ParentID != nil {
			if *args.ParentID == 0 {
				query = query.Set("parent_id", nil)
			} else {
				query = query.Set("parent_id", *args.ParentID)
			}
		}
		_, err := tx.ExecQuery(query)
		return err
	})
}

// RemoveSpendType removes Spend Type with passed id
func (db DB) RemoveSpendType(ctx context.Context, id uint) error {
	return db.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
		if !checkSpendType(tx, id) {
			return common.ErrSpendTypeNotExist
		}

		// TODO: maybe we should return an error if the spend type is used?

		// Reset Spend Type
		_, err := tx.Exec(`UPDATE monthly_payments SET type_id = NULL WHERE type_id = ?`, id)
		if err != nil {
			return errors.Wrap(err, "couldn't reset Spend Type for Monthly Payments")
		}
		_, err = tx.Exec(`UPDATE spends SET type_id = NULL WHERE type_id = ?`, id)
		if err != nil {
			return errors.Wrap(err, "couldn't reset Spend Type for Spends")
		}

		// Remove Spend Type
		_, err = tx.Exec(`DELETE FROM spend_types WHERE id = ?`, id)
		if err != nil {
			return err
		}

		return nil
	})
}
