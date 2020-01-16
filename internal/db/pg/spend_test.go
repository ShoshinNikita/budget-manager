// +build integration

package db

import (
	"context"
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/require"

	. "github.com/ShoshinNikita/budget-manager/internal/db"
	. "github.com/ShoshinNikita/budget-manager/internal/db/models"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestAddSpend(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer cleanUp(require, db)

	spends := []struct {
		Spend
		isError bool
	}{
		{
			Spend: Spend{ID: 1, DayID: 1, Title: "first spend", Notes: "123", Cost: money.FromInt(5000)},
		},
		{
			Spend: Spend{ID: 2, DayID: 15, Title: "another spend", Cost: money.FromInt(15)},
		},
		{
			Spend: Spend{ID: 3, DayID: 22, Title: "ывоаыоварод", Cost: money.FromInt(1)},
		},
		// With errors
		{
			Spend:   Spend{ID: 4, DayID: 22, Title: "", Cost: money.FromInt(1)},
			isError: true,
		},
		{
			Spend:   Spend{ID: 5, DayID: 22, Title: "123", Cost: money.FromInt(0)},
			isError: true,
		},
		{
			Spend:   Spend{ID: 6, DayID: 22, Title: "456", Cost: money.FromInt(-50)},
			isError: true,
		},
		{
			Spend:   Spend{ID: 7, DayID: 4000, Title: "first spend", Notes: "123", Cost: money.FromInt(5000)},
			isError: true,
		},
	}

	// Add Spends
	for i, sp := range spends {
		args := AddSpendArgs{
			DayID:  sp.DayID,
			Title:  sp.Title,
			TypeID: sp.TypeID,
			Notes:  sp.Notes,
			Cost:   sp.Cost,
		}
		id, err := db.AddSpend(context.Background(), args)
		if sp.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Check Spends
	for _, sp := range spends {
		spend := &Spend{ID: sp.ID}
		err := db.db.Select(spend)
		if sp.isError {
			require.Equal(pg.ErrNoRows, err)
			continue
		}
		require.Nil(err)
		require.Equal(sp.Spend, *spend)
	}

	// Check Total Spend
	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			if sp.isError {
				continue
			}
			totalSpend -= sp.Cost.ToInt()
		}
		return totalSpend
	})
}

func TestEditSpend(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer cleanUp(require, db)

	spends := []struct {
		origin  Spend
		edited  *Spend
		isError bool
	}{
		{
			origin: Spend{
				ID: 1, DayID: 10, Title: "123", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000),
			},
			edited: &Spend{
				ID: 1, DayID: 10, Title: "123456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(15500),
			},
		},
		{
			origin: Spend{
				ID: 2, DayID: 1, Title: "123", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000),
			},
			edited: &Spend{
				ID: 2, DayID: 1, Title: "123", TypeID: 1, Notes: "", Cost: money.FromInt(150),
			},
		},
		{
			origin: Spend{
				ID: 3, DayID: 1, Title: "123", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000),
			},
			edited: &Spend{
				ID: 3, DayID: 1, Title: "123", TypeID: 0, Notes: "", Cost: money.FromInt(150),
			},
		},
		// With errors
		{
			origin: Spend{
				ID: 4, DayID: 10, Title: "456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000),
			},
			edited: &Spend{
				ID: 4, DayID: 10, Title: "", TypeID: 12, Notes: "test notes", Cost: money.FromInt(15500),
			},
			isError: true,
		},
		{
			origin: Spend{
				ID: 5, DayID: 10, Title: "456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000)},
			edited: &Spend{
				ID: 5, DayID: 10, Title: "456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(0),
			},
			isError: true,
		},
		{
			origin: Spend{
				ID: 6, DayID: 10, Title: "456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000)},
			edited: &Spend{
				ID: 6, DayID: 10, Title: "456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(-50),
			},
			isError: true,
		},
		{
			origin: Spend{
				ID: 6, DayID: 10, Title: "456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000)},
			edited: &Spend{
				ID: 200, DayID: 10, Title: "456", TypeID: 12, Notes: "test notes", Cost: money.FromInt(155000),
			},
			isError: true,
		},
	}

	// Add spends
	for i, sp := range spends {
		args := AddSpendArgs{
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
			totalSpend -= sp.origin.Cost.ToInt()
		}
		return totalSpend
	})

	// Edit spends
	for _, sp := range spends {
		if sp.edited == nil {
			continue
		}

		args := EditSpendArgs{
			ID:     sp.edited.ID,
			Title:  &sp.edited.Title,
			TypeID: &sp.edited.TypeID,
			Notes:  &sp.edited.Notes,
			Cost:   &sp.edited.Cost,
		}
		err := db.EditSpend(context.Background(), args)
		if sp.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
	}

	// Check spends
	for _, sp := range spends {
		spend := &Spend{ID: sp.origin.ID}
		err := db.db.Select(spend)
		require.Nil(err)

		if sp.edited == nil || sp.isError {
			require.Equal(sp.origin, *spend)
			continue
		}
		require.Equal(*sp.edited, *spend)
	}

	// Check Total Spend

	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			if sp.edited == nil || sp.isError {
				totalSpend -= sp.origin.Cost.ToInt()
				continue
			}
			totalSpend -= sp.edited.Cost.ToInt()
		}
		return totalSpend
	})
}

func TestDeleteSpend(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer cleanUp(require, db)

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
		args := AddSpendArgs{
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
			totalSpend -= sp.Cost.ToInt()
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
		spend := &Spend{ID: sp.ID}
		err := db.db.Select(spend)
		if sp.shouldDelete && !sp.isError {
			require.Equal(pg.ErrNoRows, err)
		}
	}

	// Check Total Spend
	checkTotalSpend(db, require, func() int64 {
		var totalSpend int64
		for _, sp := range spends {
			if !sp.shouldDelete || sp.isError {
				totalSpend -= sp.Cost.ToInt()
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
	require.Equal(totalSpend, m.TotalSpend.ToInt())
}
