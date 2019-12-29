// +build integration

package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget_manager/internal/pkg/money"
)

func TestAddMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer cleanUp(require, db)

	// Prepare
	var income = money.FromInt(50000)
	_, err := db.AddIncome(AddIncomeArgs{MonthID: monthID, Title: "1", Income: income})
	require.Nil(err)

	payments := []struct {
		MonthlyPayment
		isError bool
	}{
		{
			MonthlyPayment: MonthlyPayment{
				ID: 1, MonthID: monthID, Title: "Rent", Cost: money.FromInt(20000), Notes: "some notes",
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 2, MonthID: monthID, Title: "Loans", Cost: money.FromInt(1000), TypeID: 5,
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 3, MonthID: monthID, Title: "Music", Cost: money.FromInt(300),
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 4, MonthID: monthID, Title: "Netflix", Cost: money.FromInt(600),
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 5, MonthID: monthID, Title: "Patreon", Cost: money.FromInt(1000),
			},
		},
		// With errors
		{
			MonthlyPayment: MonthlyPayment{
				ID: 0, MonthID: monthID, Title: "", Cost: money.FromInt(1000),
			},
			isError: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 0, MonthID: monthID, Title: "Some name", Cost: money.FromInt(0),
			},
			isError: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 0, MonthID: monthID, Title: "Another name", Cost: money.FromInt(-1000),
			},
			isError: true,
		},
	}

	// Add Monthly Payments
	for _, p := range payments {
		args := AddMonthlyPaymentArgs{
			MonthID: p.MonthID,
			Title:   p.Title,
			TypeID:  p.TypeID,
			Notes:   p.Notes,
			Cost:    p.Cost,
		}
		id, err := db.AddMonthlyPayment(args)
		if p.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
		require.Equal(p.ID, id)
	}

	// Check Monthly Payments
	for _, p := range payments {
		if p.isError {
			continue
		}

		mp := &MonthlyPayment{ID: p.ID}
		err := db.db.Select(mp)
		require.Nil(err)
		require.Equal(p.MonthlyPayment, *mp)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64 = income.ToInt()
		for _, p := range payments {
			b -= p.Cost.ToInt()
		}
		return b / int64(daysInMonth(time.Now()))
	}()

	m, err := db.GetMonth(monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.ToInt())
}

func TestEditMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer cleanUp(require, db)

	// Prepare
	var income = money.FromInt(50000)
	_, err := db.AddIncome(AddIncomeArgs{MonthID: monthID, Title: "1", Income: income})
	require.Nil(err)

	payments := []struct {
		origin  MonthlyPayment
		edited  *MonthlyPayment
		isError bool
	}{
		{
			origin: MonthlyPayment{
				ID: 1, MonthID: monthID, Title: "test", Notes: "123", Cost: money.FromInt(15000),
			},
			edited: &MonthlyPayment{
				ID: 1, MonthID: monthID, Title: "test", Notes: "123", Cost: money.FromInt(12000),
			},
		},
		{
			origin: MonthlyPayment{
				ID: 2, MonthID: monthID, Title: "test", Notes: "123", Cost: money.FromInt(15000),
			},
			edited: &MonthlyPayment{
				ID: 2, MonthID: monthID, Title: "123", Notes: "", Cost: money.FromInt(12000),
			},
		},
		{
			origin: MonthlyPayment{
				ID: 3, MonthID: monthID, Title: "test", Notes: "123", Cost: money.FromInt(15000),
			},
		},
		// With error
		{
			origin: MonthlyPayment{
				ID: 4, MonthID: monthID, Title: "test", Notes: "123", Cost: money.FromInt(15000),
			},
			edited: &MonthlyPayment{
				ID: 4, MonthID: monthID, Title: "", Cost: money.FromInt(100),
			},
			isError: true,
		},
		{
			origin: MonthlyPayment{
				ID: 5, MonthID: monthID, Title: "test", Notes: "123", Cost: money.FromInt(15000),
			},
			edited: &MonthlyPayment{
				ID: 5, MonthID: monthID, Title: "132", Cost: money.FromInt(0),
			},
			isError: true,
		},
		{
			origin: MonthlyPayment{
				ID: 6, MonthID: monthID, Title: "test", Notes: "123", Cost: money.FromInt(15000),
			},
			edited: &MonthlyPayment{
				ID: 6, MonthID: monthID, Title: "Test", Cost: money.FromInt(-50),
			},
			isError: true,
		},
	}

	// Add Monthly Payments
	for _, p := range payments {
		args := AddMonthlyPaymentArgs{
			MonthID: p.origin.MonthID,
			Title:   p.origin.Title,
			TypeID:  p.origin.TypeID,
			Notes:   p.origin.Notes,
			Cost:    p.origin.Cost,
		}
		id, err := db.AddMonthlyPayment(args)
		require.Nil(err)
		require.Equal(p.origin.ID, id)
	}

	// Edit Monthly Payments
	for _, p := range payments {
		if p.edited == nil {
			continue
		}

		args := EditMonthlyPaymentArgs{
			ID:     p.edited.ID,
			Title:  &p.edited.Title,
			TypeID: &p.edited.TypeID,
			Notes:  &p.edited.Notes,
			Cost:   &p.edited.Cost,
		}
		err := db.EditMonthlyPayment(args)
		if p.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
	}

	// Check Monthly Payments
	for _, p := range payments {
		mp := &MonthlyPayment{ID: p.origin.ID}
		err = db.db.Select(mp)
		require.Nil(err)
		if p.edited == nil || p.isError {
			require.Equal(p.origin, *mp)
			continue
		}
		require.Equal(*p.edited, *mp)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64 = income.ToInt()
		for _, p := range payments {
			if p.edited == nil || p.isError {
				b -= p.origin.Cost.ToInt()
			} else {
				b -= p.edited.Cost.ToInt()
			}
		}
		return b / int64(daysInMonth(time.Now()))
	}()

	m, err := db.GetMonth(monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.ToInt())

	// Try to edit Monthly Payment with invalid id
	title := "title"
	cost := money.FromInt(2000)
	err = db.EditMonthlyPayment(EditMonthlyPaymentArgs{ID: 20, Title: &title, Cost: &cost})
	require.NotNil(err)
}

func TestRemoveMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer cleanUp(require, db)

	payments := []struct {
		MonthlyPayment
		shouldDelete bool
	}{
		{
			MonthlyPayment: MonthlyPayment{
				ID: 1, MonthID: monthID, Title: "title 1", Notes: "123", Cost: money.FromInt(200),
			},
			shouldDelete: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 2, MonthID: monthID, Title: "title 2", Cost: money.FromInt(15),
			},
			shouldDelete: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 3, MonthID: monthID, Title: "title 3", Cost: money.FromInt(6000),
			},
			shouldDelete: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 4, MonthID: monthID, Title: "title 4", Notes: "1233215", Cost: money.FromInt(28 * 29 * 30 * 31),
			},
			shouldDelete: false,
		},
	}

	// Add Monthly Payments
	for _, p := range payments {
		args := AddMonthlyPaymentArgs{
			MonthID: p.MonthID,
			Title:   p.Title,
			TypeID:  p.TypeID,
			Notes:   p.Notes,
			Cost:    p.Cost,
		}
		id, err := db.AddMonthlyPayment(args)
		require.Nil(err)
		require.Equal(p.ID, id)
	}

	// Remove Monthly Payments
	for _, p := range payments {
		if !p.shouldDelete {
			continue
		}
		err := db.RemoveMonthlyPayment(p.ID)
		require.Nil(err)
	}
	// Try to remove with invalid id
	err := db.RemoveMonthlyPayment(10)
	require.NotNil(err)

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64
		for _, p := range payments {
			if !p.shouldDelete {
				b -= p.Cost.ToInt()
			}
		}
		return b / int64(daysInMonth(time.Now()))
	}()

	m, err := db.GetMonth(monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.ToInt())
}
