// +build integration

package pg

import (
	"context"
	"testing"
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/require"

	. "github.com/ShoshinNikita/budget-manager/internal/db"
	. "github.com/ShoshinNikita/budget-manager/internal/db/models"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestAddIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer cleanUp(require, db)

	incomes := []struct {
		Income
		isError bool
	}{
		{
			Income: Income{
				ID: 1, MonthID: monthID, Title: "Salary", Notes: "Not big :(", Income: money.FromInt(30000),
			},
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Gifts", Notes: "From parents", Income: money.FromInt(5000),
			},
		},
		{
			Income: Income{
				ID: 3, MonthID: monthID, Title: "Another birthdate gifts", Income: money.FromInt(3000),
			},
		},
		// With errors
		{
			Income: Income{
				ID: 0, MonthID: monthID, Title: "", Notes: "From friends", Income: money.FromInt(3000),
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 0, MonthID: monthID, Title: "Birthdate gifts", Income: money.FromInt(0),
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 0, MonthID: monthID, Title: "Gifts 2", Notes: "From friends", Income: money.FromInt(-500),
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
		id, err := db.AddIncome(context.Background(), args)
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
			b += in.Income.Income.ToInt()
		}
		return b / int64(daysInMonth(time.Now()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.ToInt())
}

func TestEditIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer cleanUp(require, db)

	incomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary", Income: money.FromInt(15000)},
		{ID: 2, MonthID: monthID, Title: "Birthdate gifts", Notes: "From parents", Income: money.FromInt(5000)},
	}

	editedIncomes := []struct {
		Income
		isError bool
	}{
		{
			Income: Income{
				ID: 1, MonthID: monthID, Title: "Salary++", Income: money.FromInt(20000),
			},
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Birthdate gifts from parents", Income: money.FromInt(5000),
			},
		},
		// With errors
		{
			Income: Income{
				ID: 100, MonthID: monthID, Title: "Valid title", Income: money.FromInt(5000),
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 1, MonthID: monthID, Title: "", Income: money.FromInt(5000),
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Birthdate gifts from parents", Income: money.FromInt(0),
			},
			isError: true,
		},
		{
			Income: Income{
				ID: 2, MonthID: monthID, Title: "Birthdate gifts from parents", Income: money.FromInt(-100),
			},
			isError: true,
		},
	}

	// Add Incomes
	for _, in := range incomes {
		args := AddIncomeArgs{
			MonthID: in.MonthID, Title: in.Title, Notes: in.Notes, Income: in.Income,
		}
		_, err := db.AddIncome(context.Background(), args)
		require.Nil(err)
	}

	// Edit Incomes
	for _, in := range editedIncomes {
		args := EditIncomeArgs{
			ID:     in.ID,
			Title:  &in.Title,
			Notes:  &in.Notes,
			Income: &in.Income.Income,
		}
		err := db.EditIncome(context.Background(), args)
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
			b += in.Income.Income.ToInt()
		}
		return b / int64(daysInMonth(time.Now()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.ToInt())
}

func TestRemoveIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(require)
	defer cleanUp(require, db)

	incomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary", Notes: "Not very big :(", Income: money.FromInt(30000)},
		{ID: 2, MonthID: monthID, Title: "gifts", Notes: "From parents", Income: money.FromInt(5000)},
		{ID: 3, MonthID: monthID, Title: "gifts", Notes: "From friends", Income: money.FromInt(3000)},
	}

	// Add Incomes
	for i, in := range incomes {
		args := AddIncomeArgs{
			MonthID: in.MonthID,
			Title:   in.Title,
			Notes:   in.Notes,
			Income:  in.Income,
		}
		id, err := db.AddIncome(context.Background(), args)
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Remove Income with id = 1
	err := db.RemoveIncome(context.Background(), 1)
	require.Nil(err)

	// Check daily budget (without Income with id = 1)
	dailyBudget := func() int64 {
		var b int64
		for _, in := range incomes {
			if in.ID == 1 {
				continue
			}
			b += in.Income.ToInt()
		}

		return b / int64(daysInMonth(time.Now()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.ToInt())

	// Try to remove Income with invalid id
	err = db.RemoveIncome(context.Background(), 100)
	require.NotNil(err)
}
