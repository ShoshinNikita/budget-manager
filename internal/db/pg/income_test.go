// +build integration

package pg

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	common "github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestAddIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(t)
	defer cleanUp(t, db)

	incomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary", Notes: "Not big :(", Income: money.FromInt(30000)},
		{ID: 2, MonthID: monthID, Title: "Gifts", Notes: "From parents", Income: money.FromInt(5000)},
		{ID: 3, MonthID: monthID, Title: "Another birthdate gifts", Income: money.FromInt(3000)},
	}

	// Add Incomes
	for i, in := range incomes {
		args := common.AddIncomeArgs{
			MonthID: in.MonthID,
			Title:   in.Title,
			Notes:   in.Notes,
			Income:  in.Income,
		}
		id, err := db.AddIncome(context.Background(), args)
		require.Nil(err)
		require.Equal(uint(i+1), id)
	}

	// Check Incomes
	for _, in := range incomes {
		var income Income
		err := db.db.Model(&income).Where("id = ?", in.ID).Select()
		require.Nil(err)
		require.Equal(in, income)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64
		for _, in := range incomes {
			b += in.Income.Int()
		}
		now := time.Now()
		return b / int64(daysInMonth(now.Year(), now.Month()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.Int())
}

func TestEditIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(t)
	defer cleanUp(t, db)

	incomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary", Income: money.FromInt(15000)},
		{ID: 2, MonthID: monthID, Title: "Birthdate gifts", Notes: "From parents", Income: money.FromInt(5000)},
	}

	editedIncomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary++", Income: money.FromInt(20000)},
		{ID: 2, MonthID: monthID, Title: "Birthdate gifts from parents", Income: money.FromInt(5000)},
	}

	// Add Incomes
	for _, in := range incomes {
		args := common.AddIncomeArgs{
			MonthID: in.MonthID, Title: in.Title, Notes: in.Notes, Income: in.Income,
		}
		_, err := db.AddIncome(context.Background(), args)
		require.Nil(err)
	}

	// Edit Incomes
	for _, in := range editedIncomes {
		args := common.EditIncomeArgs{
			ID:     in.ID,
			Title:  &in.Title,
			Notes:  &in.Notes,
			Income: &in.Income,
		}
		err := db.EditIncome(context.Background(), args)
		require.Nil(err)
	}

	// Check Incomes
	for _, in := range editedIncomes {
		var income Income
		err := db.db.Model(&income).Where("id = ?", in.ID).Select()
		require.Nil(err)
		require.Equal(in, income)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64
		for _, in := range editedIncomes {
			b += in.Income.Int()
		}
		now := time.Now()
		return b / int64(daysInMonth(now.Year(), now.Month()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.Int())
}

func TestRemoveIncome(t *testing.T) {
	require := require.New(t)

	// Init db
	db := initDB(t)
	defer cleanUp(t, db)

	incomes := []Income{
		{ID: 1, MonthID: monthID, Title: "Salary", Notes: "Not very big :(", Income: money.FromInt(30000)},
		{ID: 2, MonthID: monthID, Title: "gifts", Notes: "From parents", Income: money.FromInt(5000)},
		{ID: 3, MonthID: monthID, Title: "gifts", Notes: "From friends", Income: money.FromInt(3000)},
	}

	// Add Incomes
	for i, in := range incomes {
		args := common.AddIncomeArgs{
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
			b += in.Income.Int()
		}

		now := time.Now()
		return b / int64(daysInMonth(now.Year(), now.Month()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.Int())

	// Try to remove Income with invalid id
	err = db.RemoveIncome(context.Background(), 100)
	require.NotNil(err)
}
