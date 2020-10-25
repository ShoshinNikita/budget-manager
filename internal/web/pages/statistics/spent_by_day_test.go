package statistics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestCalculateSpentByDay(t *testing.T) {
	t.Parallel()

	date := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}

	tests := []struct {
		spends    []db.Spend
		startDate time.Time
		endDate   time.Time
		//
		want SpentByDayDataset
	}{
		{
			spends: []db.Spend{
				{
					Year: 2020, Month: 10, Day: 25,
					Cost: money.FromInt(100),
				},
				{
					Year: 2020, Month: 10, Day: 25,
					Cost: money.FromInt(150),
				},
				{
					Year: 2020, Month: 10, Day: 27,
					Cost: money.FromInt(1),
				},
				{
					Year: 2020, Month: 10, Day: 30,
					Cost: money.FromInt(33),
				},
			},
			startDate: date(2020, 10, 25),
			endDate:   date(2020, 10, 30),
			//
			want: SpentByDayDataset{
				{Year: 2020, Month: 10, Day: 25, Spent: money.FromInt(250)},
				{Year: 2020, Month: 10, Day: 26, Spent: 0},
				{Year: 2020, Month: 10, Day: 27, Spent: money.FromInt(1)},
				{Year: 2020, Month: 10, Day: 28, Spent: 0},
				{Year: 2020, Month: 10, Day: 29, Spent: 0},
				{Year: 2020, Month: 10, Day: 30, Spent: money.FromInt(33)},
			},
		},
		{
			spends: []db.Spend{
				{
					Year: 2020, Month: 10, Day: 30,
					Cost: money.FromInt(33),
				},
				{
					Year: 2020, Month: 11, Day: 03,
					Cost: money.FromInt(55),
				},
			},
			startDate: date(2020, 10, 28),
			endDate:   date(2020, 11, 04),
			//
			want: SpentByDayDataset{
				{Year: 2020, Month: 10, Day: 28, Spent: 0},
				{Year: 2020, Month: 10, Day: 29, Spent: 0},
				{Year: 2020, Month: 10, Day: 30, Spent: money.FromInt(33)},
				{Year: 2020, Month: 10, Day: 31, Spent: 0},
				{Year: 2020, Month: 11, Day: 01, Spent: 0},
				{Year: 2020, Month: 11, Day: 02, Spent: 0},
				{Year: 2020, Month: 11, Day: 03, Spent: money.FromInt(55)},
				{Year: 2020, Month: 11, Day: 04, Spent: 0},
			},
		},
		{
			spends: []db.Spend{
				{
					Year: 2020, Month: 10, Day: 30,
					Cost: money.FromInt(33),
				},
				{
					Year: 2020, Month: 11, Day: 03,
					Cost: money.FromInt(55),
				},
			},
			//
			want: SpentByDayDataset{
				{Year: 2020, Month: 10, Day: 30, Spent: money.FromInt(33)},
				{Year: 2020, Month: 10, Day: 31, Spent: 0},
				{Year: 2020, Month: 11, Day: 01, Spent: 0},
				{Year: 2020, Month: 11, Day: 02, Spent: 0},
				{Year: 2020, Month: 11, Day: 03, Spent: money.FromInt(55)},
			},
		},
		{
			spends: []db.Spend{
				{
					Year: 2019, Month: 12, Day: 30,
					Cost: money.FromInt(1000),
				},
				{
					Year: 2020, Month: 01, Day: 03,
					Cost: money.FromInt(99),
				},
			},
			//
			want: SpentByDayDataset{
				{Year: 2019, Month: 12, Day: 30, Spent: money.FromInt(1000)},
				{Year: 2019, Month: 12, Day: 31, Spent: 0},
				{Year: 2020, Month: 01, Day: 01, Spent: 0},
				{Year: 2020, Month: 01, Day: 02, Spent: 0},
				{Year: 2020, Month: 01, Day: 03, Spent: money.FromInt(99)},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			got := CalculateSpentByDay(tt.spends, tt.startDate, tt.endDate)
			require.Equal(t, tt.want, got)
		})
	}
}
