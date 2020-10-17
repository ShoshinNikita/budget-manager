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

func TestAddMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(t)
	defer cleanUp(t, db)

	// Prepare
	income := money.FromInt(50000)
	_, err := db.AddIncome(
		context.Background(), common.AddIncomeArgs{MonthID: monthID, Title: "1", Income: income},
	)
	require.Nil(err)

	_, err = db.AddSpendType(context.Background(), common.AddSpendTypeArgs{Name: "spend type"})
	require.Nil(err)

	payments := []MonthlyPayment{
		{ID: 1, MonthID: monthID, Title: "Rent", Cost: money.FromInt(20000), Notes: "some notes"},
		{ID: 2, MonthID: monthID, Title: "Loans", Cost: money.FromInt(1000), TypeID: 1},
		{ID: 3, MonthID: monthID, Title: "Music", Cost: money.FromInt(300)},
		{ID: 4, MonthID: monthID, Title: "Netflix", Cost: money.FromInt(600)},
		{ID: 5, MonthID: monthID, Title: "Patreon", Cost: money.FromInt(1000)},
	}

	// Add Monthly Payments
	for _, p := range payments {
		args := common.AddMonthlyPaymentArgs{
			MonthID: p.MonthID,
			Title:   p.Title,
			TypeID:  p.TypeID,
			Notes:   p.Notes,
			Cost:    p.Cost,
		}
		id, err := db.AddMonthlyPayment(context.Background(), args)
		require.Nil(err)
		require.Equal(p.ID, id)
	}

	// Check Monthly Payments
	for _, p := range payments {
		var mp MonthlyPayment
		err := db.db.Model(&mp).Where("id = ?", p.ID).Select()
		require.Nil(err)
		require.Equal(p, mp)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64 = income.Int()
		for _, p := range payments {
			b -= p.Cost.Int()
		}
		now := time.Now()
		return b / int64(daysInMonth(now.Year(), now.Month()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.Int())
}

func TestEditMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(t)
	defer cleanUp(t, db)

	// Prepare
	income := money.FromInt(50000)
	_, err := db.AddIncome(
		context.Background(), common.AddIncomeArgs{MonthID: monthID, Title: "1", Income: income},
	)
	require.Nil(err)

	payments := []struct {
		origin MonthlyPayment
		edited *MonthlyPayment
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
	}

	// Add Monthly Payments
	for _, p := range payments {
		args := common.AddMonthlyPaymentArgs{
			MonthID: p.origin.MonthID,
			Title:   p.origin.Title,
			TypeID:  p.origin.TypeID,
			Notes:   p.origin.Notes,
			Cost:    p.origin.Cost,
		}
		id, err := db.AddMonthlyPayment(context.Background(), args)
		require.Nil(err)
		require.Equal(p.origin.ID, id)
	}

	// Edit Monthly Payments
	for _, p := range payments {
		if p.edited == nil {
			continue
		}

		args := common.EditMonthlyPaymentArgs{
			ID:    p.edited.ID,
			Title: &p.edited.Title,
			Notes: &p.edited.Notes,
			Cost:  &p.edited.Cost,
		}
		if p.origin.TypeID != p.edited.TypeID {
			args.TypeID = &p.edited.TypeID
		}
		err := db.EditMonthlyPayment(context.Background(), args)
		require.Nil(err)
	}

	// Check Monthly Payments
	for _, p := range payments {
		var mp MonthlyPayment
		err = db.db.Model(&mp).Where("id = ?", p.origin.ID).Select()
		require.Nil(err)
		if p.edited == nil {
			require.Equal(p.origin, mp)
			continue
		}
		require.Equal(*p.edited, mp)
	}

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64 = income.Int()
		for _, p := range payments {
			if p.edited == nil {
				b -= p.origin.Cost.Int()
			} else {
				b -= p.edited.Cost.Int()
			}
		}
		now := time.Now()
		return b / int64(daysInMonth(now.Year(), now.Month()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.Int())

	// Try to edit Monthly Payment with invalid id
	title := "title"
	cost := money.FromInt(2000)
	err = db.EditMonthlyPayment(
		context.Background(), common.EditMonthlyPaymentArgs{ID: 20, Title: &title, Cost: &cost},
	)
	require.NotNil(err)
}

func TestRemoveMonthlyPayment(t *testing.T) {
	require := require.New(t)

	db := initDB(t)
	defer cleanUp(t, db)

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
		args := common.AddMonthlyPaymentArgs{
			MonthID: p.MonthID,
			Title:   p.Title,
			TypeID:  p.TypeID,
			Notes:   p.Notes,
			Cost:    p.Cost,
		}
		id, err := db.AddMonthlyPayment(context.Background(), args)
		require.Nil(err)
		require.Equal(p.ID, id)
	}

	// Remove Monthly Payments
	for _, p := range payments {
		if !p.shouldDelete {
			continue
		}
		err := db.RemoveMonthlyPayment(context.Background(), p.ID)
		require.Nil(err)
	}
	// Try to remove with invalid id
	err := db.RemoveMonthlyPayment(context.Background(), 10)
	require.NotNil(err)

	// Check daily budget
	dailyBudget := func() int64 {
		var b int64
		for _, p := range payments {
			if !p.shouldDelete {
				b -= p.Cost.Int()
			}
		}
		now := time.Now()
		return b / int64(daysInMonth(now.Year(), now.Month()))
	}()

	m, err := db.GetMonth(context.Background(), monthID)
	require.Nil(err)
	require.Equal(dailyBudget, m.DailyBudget.Int())
}
