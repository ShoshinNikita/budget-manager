package pg

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/pkg/errors"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
)

// SpendType represents spend type entity in PostgreSQL db
type SpendType struct {
	tableName struct{} `pg:"spend_types"`

	ID       uint   `pg:"id,pk"`
	Name     string `pg:"name"`
	ParentID uint   `pg:"parent_id"`
}

// ToCommon converts SpendType to common SpendType structure from
// "github.com/ShoshinNikita/budget-manager/internal/db" package
//
// We return a pointer instead of a value unlike other 'ToCommon' methods because Spend Type can be optional
func (s *SpendType) ToCommon() *common.SpendType {
	if s == nil {
		return nil
	}
	return &common.SpendType{
		ID:       s.ID,
		Name:     s.Name,
		ParentID: s.ParentID,
	}
}

// GetSpendType returns Spend Type with passed id
func (db DB) GetSpendType(_ context.Context, id uint) (common.SpendType, error) {
	var spendType SpendType
	if err := db.db.Model(&spendType).Where("id = ?", id).Select(); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			err = common.ErrSpendTypeNotExist
		}
		return common.SpendType{}, err
	}

	return *spendType.ToCommon(), nil
}

// GetSpendTypes returns all Spend Types
func (db DB) GetSpendTypes(_ context.Context) ([]common.SpendType, error) {
	var spendTypes []SpendType
	err := db.db.Model(&spendTypes).Order("id ASC").Select()
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
func (db DB) AddSpendType(_ context.Context, args common.AddSpendTypeArgs) (id uint, err error) {
	err = db.db.RunInTransaction(func(tx *pg.Tx) error {
		spendType := &SpendType{
			Name:     args.Name,
			ParentID: args.ParentID,
		}
		if _, err := tx.Model(spendType).Returning("id").Insert(); err != nil {
			return err
		}
		id = spendType.ID

		return nil
	})
	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditSpendType modifies existing Spend Type
func (db DB) EditSpendType(_ context.Context, args common.EditSpendTypeArgs) error {
	if !db.checkSpendType(args.ID) {
		return common.ErrSpendTypeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		query := tx.Model((*SpendType)(nil)).Where("id = ?", args.ID)
		if args.Name != nil {
			query = query.Set("name = ?", *args.Name)
		}
		if args.ParentID != nil {
			query = query.Set("parent_id = ?", *args.ParentID)
		}

		_, err := query.Update()
		return err
	})
}

// RemoveSpendType removes Spend Type with passed id
func (db DB) RemoveSpendType(_ context.Context, id uint) error {
	if !db.checkSpendType(id) {
		return common.ErrSpendTypeNotExist
	}

	return db.db.RunInTransaction(func(tx *pg.Tx) error {
		_, err := tx.Model((*SpendType)(nil)).Where("id = ?", id).Delete()
		if err != nil {
			return err
		}

		// Reset Type IDs

		_, err = tx.Model((*MonthlyPayment)(nil)).Set("type_id = 0").Where("type_id = ?", id).Update()
		if err != nil {
			return errors.Wrap(err, "couldn't reset Type ID of Monthly Payments")
		}

		_, err = tx.Model((*Spend)(nil)).Set("type_id = 0").Where("type_id = ?", id).Update()
		if err != nil {
			return errors.Wrap(err, "couldn't reset Type ID of Spends")
		}

		return nil
	})
}
