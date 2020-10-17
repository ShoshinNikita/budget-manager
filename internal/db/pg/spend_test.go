// +build integration

package pg

import (
	"context"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/stretchr/testify/require"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestAddSpend(t *testing.T) {
	require := require.New(t)

	db := initDB(t)
	defer cleanUp(t, db)

	spends := []Spend{
		{ID: 1, DayID: 1, Title: "first spend", Notes: "123", Cost: money.FromInt(5000)},
		{ID: 2, DayID: 15, Title: "another spend", Cost: money.FromInt(15)},
		{ID: 3, DayID: 22, Title: "ывоаыоварод", Cost: money.FromInt(1)},
	}

	// Add Spends
	for i, sp := range spends {
		args := common.AddSpendArgs{
			DayID:  sp.DayID,
			Title:  sp.Title,
			TypeID: sp.TypeID,
			Notes:  sp.Notes,
			Cost:   sp.Cost,
		}
		id, err := db.AddSpend(context.Background(), args)
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Check Spends
	for _, sp := range spends {
		var spend Spend
		err := db.db.Model(&spend).Where("id = ?", sp.ID).Select()
		require.Nil(err)
		require.Equal(sp, spend)
	}

	// Check Total Spend
	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			totalSpend -= sp.Cost.Int()
		}
		return totalSpend
	})
}

func TestEditSpend(t *testing.T) {
	require := require.New(t)

	db := initDB(t)
	defer cleanUp(t, db)

	spends := []struct {
		origin Spend
		edited *Spend
	}{
		{
			origin: Spend{
				ID: 1, DayID: 10, Title: "123", TypeID: 2, Notes: "test notes", Cost: money.FromInt(155000),
			},
			edited: &Spend{
				ID: 1, DayID: 10, Title: "123456", TypeID: 2, Notes: "test notes", Cost: money.FromInt(15500),
			},
		},
		{
			origin: Spend{
				ID: 2, DayID: 1, Title: "123", TypeID: 2, Notes: "test notes", Cost: money.FromInt(155000),
			},
			edited: &Spend{
				ID: 2, DayID: 1, Title: "123", TypeID: 1, Notes: "", Cost: money.FromInt(150),
			},
		},
		{
			origin: Spend{
				ID: 3, DayID: 1, Title: "123", TypeID: 2, Notes: "test notes", Cost: money.FromInt(155000),
			},
			edited: &Spend{
				ID: 3, DayID: 1, Title: "123", TypeID: 0, Notes: "", Cost: money.FromInt(150),
			},
		},
	}

	// Add Spend Types
	for _, name := range []string{"first", "second"} {
		_, err := db.AddSpendType(context.Background(), common.AddSpendTypeArgs{Name: name})
		require.Nil(err)
	}

	// Add spends
	for i, sp := range spends {
		args := common.AddSpendArgs{
			DayID:  sp.origin.DayID,
			Title:  sp.origin.Title,
			TypeID: sp.origin.TypeID,
			Notes:  sp.origin.Notes,
			Cost:   sp.origin.Cost,
		}
		id, err := db.AddSpend(context.Background(), args)
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Check Total Spend

	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			totalSpend -= sp.origin.Cost.Int()
		}
		return totalSpend
	})

	// Edit spends
	for _, sp := range spends {
		if sp.edited == nil {
			continue
		}

		args := common.EditSpendArgs{
			ID:    sp.edited.ID,
			Title: &sp.edited.Title,
			Notes: &sp.edited.Notes,
			Cost:  &sp.edited.Cost,
		}
		if sp.origin.TypeID != sp.edited.TypeID {
			args.TypeID = &sp.edited.TypeID
		}
		err := db.EditSpend(context.Background(), args)
		require.Nil(err)
	}

	// Check spends
	for _, sp := range spends {
		var spend Spend
		err := db.db.Model(&spend).Where("id = ?", sp.origin.ID).Select()
		require.Nil(err)

		if sp.edited == nil {
			require.Equal(sp.origin, spend)
			continue
		}
		require.Equal(*sp.edited, spend)
	}

	// Check Total Spend

	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			if sp.edited == nil {
				totalSpend -= sp.origin.Cost.Int()
				continue
			}
			totalSpend -= sp.edited.Cost.Int()
		}
		return totalSpend
	})
}

func TestDeleteSpend(t *testing.T) {
	require := require.New(t)

	db := initDB(t)
	defer cleanUp(t, db)

	spends := []struct {
		Spend
		shouldDelete bool
		isError      bool
	}{
		{
			Spend: Spend{ID: 1, DayID: 1, Title: "first spend", Notes: "123", Cost: money.FromInt(5000)},
		},
		{
			Spend:        Spend{ID: 2, DayID: 15, Title: "another spend", Cost: money.FromInt(15)},
			shouldDelete: true,
		},
		{
			Spend:        Spend{ID: 3, DayID: 22, Title: "ывоаыоварод", Cost: money.FromInt(2000)},
			shouldDelete: true,
		},
		// With errors
		{
			Spend:        Spend{ID: 25, DayID: 22, Title: "ывоаыоварод", Cost: money.FromInt(2000)},
			shouldDelete: true,
			isError:      true,
		},
	}

	// Add spends
	for i, sp := range spends {
		args := common.AddSpendArgs{
			DayID:  sp.DayID,
			Title:  sp.Title,
			TypeID: sp.TypeID,
			Notes:  sp.Notes,
			Cost:   sp.Cost,
		}
		id, err := db.AddSpend(context.Background(), args)
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Check Total Spend
	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			totalSpend -= sp.Cost.Int()
		}
		return totalSpend
	})

	// Remove spends
	for _, sp := range spends {
		if !sp.shouldDelete {
			continue
		}

		err := db.RemoveSpend(context.Background(), sp.ID)
		if sp.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
	}

	// Check spends
	for _, sp := range spends {
		var spend Spend
		err := db.db.Model(&spend).Where("id = ?", sp.ID).Select()
		if sp.shouldDelete && !sp.isError {
			require.Equal(pg.ErrNoRows, err)
		}
	}

	// Check Total Spend
	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			if !sp.shouldDelete || sp.isError {
				totalSpend -= sp.Cost.Int()
			}
		}
		return totalSpend
	})
}

func checkTotalSpend(db *DB, require *require.Assertions, f func() int64) {
	totalSpend := f()
	if totalSpend > 0 {
		totalSpend = -totalSpend
	}

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(totalSpend, m.TotalSpend.Int())
}
