package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer db.Shutdown()

	// Prepare
	const income = 50000
	_, err := db.AddIncome(AddIncomeArgs{MonthID: monthID, Title: "1", Income: income})
	require.Nil(err)

	payments := []struct {
		MonthlyPayment
		isError bool
	}{
		{
			MonthlyPayment: MonthlyPayment{
				ID:      1,
				MonthID: monthID,
				Title:   "Rent",
				Cost:    20000,
				Notes:   "some notes",
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID:      2,
				MonthID: monthID,
				Title:   "Loans",
				Cost:    1000,
				TypeID:  5,
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID:      3,
				MonthID: monthID,
				Title:   "Music",
				Cost:    300,
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID:      4,
				MonthID: monthID,
				Title:   "Netflix",
				Cost:    600,
			},
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID:      5,
				MonthID: monthID,
				Title:   "Patreon",
				Cost:    1000,
			},
		},
		// With errors
		{
			MonthlyPayment: MonthlyPayment{
				ID:      0,
				MonthID: monthID,
				Title:   "",
				Cost:    1000,
			},
			isError: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID:      0,
				MonthID: monthID,
				Title:   "Some name",
				Cost:    0,
			},
			isError: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID:      0,
				MonthID: monthID,
				Title:   "Another name",
				Cost:    -1000,
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
		var b int64 = income
		for _, p := range payments {
			b -= p.Cost
		}
		return b / int64(daysInMonth(time.Now().Month()))
	}()

	m := Month{ID: monthID}
	err = db.db.Select(&m)
	require.Nil(err)

	require.Equal(dailyBudget, m.DailyBudget)
}

func TestEditMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer db.Shutdown()

	// Prepare
	const income = 50000
	_, err := db.AddIncome(AddIncomeArgs{MonthID: monthID, Title: "1", Income: income})
	require.Nil(err)

	payments := []struct {
		origin  MonthlyPayment
		edited  *MonthlyPayment
		isError bool
	}{
		{
			origin: MonthlyPayment{
				ID:      1,
				MonthID: monthID,
				Title:   "test",
				Notes:   "123",
				Cost:    15000,
			},
			edited: &MonthlyPayment{
				ID:      1,
				MonthID: monthID,
				Title:   "test",
				Notes:   "123",
				Cost:    12000,
			},
		},
		{
			origin: MonthlyPayment{
				ID:      2,
				MonthID: monthID,
				Title:   "test",
				Notes:   "123",
				Cost:    15000,
			},
			edited: &MonthlyPayment{
				ID:      2,
				MonthID: monthID,
				Title:   "123",
				Notes:   "",
				Cost:    12000,
			},
		},
		{
			origin: MonthlyPayment{
				ID:      3,
				MonthID: monthID,
				Title:   "test",
				Notes:   "123",
				Cost:    15000,
			},
		},
		// With error
		{
			origin: MonthlyPayment{
				ID:      4,
				MonthID: monthID,
				Title:   "test",
				Notes:   "123",
				Cost:    15000,
			},
			edited: &MonthlyPayment{
				ID:      4,
				MonthID: monthID,
				Title:   "",
				Cost:    100,
			},
			isError: true,
		},
		{
			origin: MonthlyPayment{
				ID:      5,
				MonthID: monthID,
				Title:   "test",
				Notes:   "123",
				Cost:    15000,
			},
			edited: &MonthlyPayment{
				ID:      5,
				MonthID: monthID,
				Title:   "132",
				Cost:    0,
			},
			isError: true,
		},
		{
			origin: MonthlyPayment{
				ID:      6,
				MonthID: monthID,
				Title:   "test",
				Notes:   "123",
				Cost:    15000,
			},
			edited: &MonthlyPayment{
				ID:      6,
				MonthID: monthID,
				Title:   "Test",
				Cost:    -50,
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
		var b int64 = income
		for _, p := range payments {
			if p.edited == nil || p.isError {
				b -= p.origin.Cost
			} else {
				b -= p.edited.Cost
			}
		}
		return b / int64(daysInMonth(time.Now().Month()))
	}()

	m := Month{ID: monthID}
	err = db.db.Select(&m)
	require.Nil(err)

	require.Equal(dailyBudget, m.DailyBudget)

	// Try to edit Monthly Payment with invalid id
	title := "title"
	cost := int64(2000)
	err = db.EditMonthlyPayment(EditMonthlyPaymentArgs{ID: 20, Title: &title, Cost: &cost})
	require.NotNil(err)
}

func TestRemoveMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(require)
	defer db.Shutdown()

	payments := []struct {
		MonthlyPayment
		shouldDelete bool
	}{
		{
			MonthlyPayment: MonthlyPayment{
				ID: 1, MonthID: monthID, Title: "title 1", Notes: "123", Cost: 200,
			},
			shouldDelete: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 2, MonthID: monthID, Title: "title 2", Cost: 15,
			},
			shouldDelete: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 3, MonthID: monthID, Title: "title 3", Cost: 6000,
			},
			shouldDelete: true,
		},
		{
			MonthlyPayment: MonthlyPayment{
				ID: 4, MonthID: monthID, Title: "title 4", Notes: "1233215", Cost: 28 * 29 * 30 * 31,
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
				b -= p.Cost
			}
		}
		return b / int64(daysInMonth(time.Now().Month()))
	}()

	m := Month{ID: monthID}
	err = db.db.Select(&m)
	require.Nil(err)

	require.Equal(dailyBudget, m.DailyBudget)
}
