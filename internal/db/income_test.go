package db

import (
	"testing"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/require"
)

func TestAddIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer db.Shutdown()

	incomes := []struct {
		Income
		isError bool
	}{
		{
			Income: Income{
				ID: 1, MonthID: monthID, Title: "Salary", Notes: "Not big :(", Income: 30000,
			},
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Gifts", Notes: "From parents", Income: 5000,
			},
		},
		{
			Income: Income{
				ID: 3, MonthID: monthID, Title: "Another birthdate gifts", Income: 3000,
			},
		},
		// With errors
		{
			Income: Income{
				ID: 0, MonthID: monthID, Title: "", Notes: "From friends", Income: 3000,
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 0, MonthID: monthID, Title: "Birthdate gifts", Income: 0,
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 0, MonthID: monthID, Title: "Gifts 2", Notes: "From friends", Income: -50,
			},
			isError: true,
		},
	}

	// Add Incomes
	for i, in := range incomes {
		args := AddIncomeArgs{
			MonthID: in.MonthID,
			Title:   in.Title,
			Notes:   in.Notes,
			Income:  in.Income.Income,
		}
		id, err := db.AddIncome(args)
		if in.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Check Incomes
	for _, in := range incomes {
		income := &Income{ID: in.ID}
		err := db.db.Select(income)
		if in.isError {
			require.Equal(pg.ErrNoRows, err)
			continue
		}
		require.Nil(err)
		require.Equal(in.Income, *income)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64
		for _, in := range incomes {
			if in.isError {
				continue
			}
			b += in.Income.Income
		}
		return b / int64(daysInMonth(time.Now().Month()))
	}()

	m := &Month{ID: monthID}
	db.db.Model(m).Column("daily_budget").WherePK().Select()
	require.Equal(dailyBudget, m.DailyBudget)
}

func TestEditIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer db.Shutdown()

	incomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary", Income: 15000},
		{ID: 2, MonthID: monthID, Title: "Birthdate gifts", Notes: "From parents", Income: 5000},
	}

	editedIncomes := []struct {
		Income
		isError bool
	}{
		{
			Income: Income{
				ID: 1, MonthID: monthID, Title: "Salary++", Income: 20000,
			},
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Birthdate gifts from parents", Income: 5000,
			},
		},
		// With errors
		{
			Income: Income{
				ID: 100, MonthID: monthID, Title: "Valid title", Income: 5000,
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 1, MonthID: monthID, Title: "", Income: 5000,
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Birthdate gifts from parents", Income: 0,
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Birthdate gifts from parents", Income: -100,
			},
			isError: true,
		},
	}

	// Add Incomes
	for _, in := range incomes {
		args := AddIncomeArgs{
			MonthID: in.MonthID, Title: in.Title, Notes: in.Notes, Income: in.Income,
		}
		db.AddIncome(args)
	}

	// Edit Incomes
	for _, in := range editedIncomes {
		args := EditIncomeArgs{
			ID:     in.ID,
			Title:  &in.Title,
			Notes:  &in.Notes,
			Income: &in.Income.Income,
		}
		err := db.EditIncome(args)
		if in.isError {
			require.NotNil(err)
			continue
		}
		require.Nil(err)
	}

	// Check Incomes
	for _, in := range editedIncomes {
		if in.isError {
			continue
		}

		income := &Income{ID: in.ID}
		err := db.db.Select(income)
		require.Nil(err)
		require.Equal(in.Income, *income)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64
		for _, in := range editedIncomes {
			if in.isError {
				continue
			}
			b += in.Income.Income
		}
		return b / int64(daysInMonth(time.Now().Month()))
	}()

	m := &Month{ID: monthID}
	db.db.Model(m).Column("daily_budget").WherePK().Select()
	require.Equal(dailyBudget, m.DailyBudget)
}

func TestRemoveIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer db.Shutdown()

	incomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary", Notes: "Not very big :(", Income: 30000},
		{ID: 2, MonthID: monthID, Title: "Birthdate gifts", Notes: "From parents", Income: 5000},
		{ID: 3, MonthID: monthID, Title: "Birthdate gifts", Notes: "From friends", Income: 3000},
	}

	// Add Incomes
	for i, in := range incomes {
		args := AddIncomeArgs{
			MonthID: in.MonthID,
			Title:   in.Title,
			Notes:   in.Notes,
			Income:  in.Income,
		}
		id, err := db.AddIncome(args)
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Remove Income with id = 1
	err := db.RemoveIncome(1)
	require.Nil(err)

	// Check daily budget (without Income with id = 1)
	dailyBudget := func() int64 {
		var b int64
		for _, in := range incomes {
			if in.ID == 1 {
				continue
			}
			b += in.Income
		}

		return b / int64(daysInMonth(time.Now().Month()))
	}()

	m := &Month{ID: monthID}
	db.db.Model(m).Column("daily_budget").WherePK().Select()
	require.Equal(dailyBudget, m.DailyBudget)

	// Try to remove Income with invalid id
	err = db.RemoveIncome(100)
	require.NotNil(err)
}
