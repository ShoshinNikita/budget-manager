package statistics

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/db"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestCalculateCostIntervals(t *testing.T) {
	t.Parallel()

	m := money.FromInt

	tests := []struct {
		spends    []db.Spend
		intervals int
		//
		want []CostInterval
	}{
		{
			spends:    []db.Spend{},
			intervals: 5,
			//
			want: nil,
		},
		{
			spends: []db.Spend{
				{Cost: m(20)},
				{Cost: m(30)},
				{Cost: m(35)},
				{Cost: m(17)},
			},
			intervals: 5,
			//
			want: []CostInterval{
				{From: m(17), To: m(21) - 1, Count: 2, Total: m(37)},
				{From: m(21), To: m(25) - 1, Count: 0, Total: 0},
				{From: m(25), To: m(29) - 1, Count: 0, Total: 0},
				{From: m(29), To: m(33) - 1, Count: 1, Total: m(30)},
				{From: m(33), To: m(35), Count: 1, Total: m(35)},
			},
		},
		// Too many intervals
		{
			spends: []db.Spend{
				{Cost: m(1)},
				{Cost: m(2)},
				{Cost: m(3)},
				{Cost: m(4)},
			},
			intervals: 10,
			//
			want: []CostInterval{
				{From: m(1), To: m(2) - 1, Count: 1, Total: m(1)},
				{From: m(2), To: m(3) - 1, Count: 1, Total: m(2)},
				{From: m(3), To: m(4), Count: 2, Total: m(7)},
			},
		},
		{
			spends: []db.Spend{
				{Cost: m(20)},
				{Cost: m(30)},
				{Cost: m(35)},
				{Cost: m(17)},
			},
			intervals: 1,
			//
			want: []CostInterval{
				{From: m(17), To: m(35), Count: 4, Total: m(102)},
			},
		},
		{
			spends: []db.Spend{
				{Cost: m(20)},
				{Cost: m(30)},
				{Cost: m(35)},
				{Cost: m(17)},
			},
			intervals: 10,
			//
			want: []CostInterval{
				{From: m(17), To: m(20) - 1, Count: 1, Total: m(17)},
				{From: m(20), To: m(23) - 1, Count: 1, Total: m(20)},
				{From: m(23), To: m(26) - 1, Count: 0, Total: 0},
				{From: m(26), To: m(29) - 1, Count: 0, Total: 0},
				{From: m(29), To: m(32) - 1, Count: 1, Total: m(30)},
				{From: m(32), To: m(35), Count: 1, Total: m(35)},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			got := CalculateCostIntervals(tt.spends, tt.intervals)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPrepareIntervals(t *testing.T) {
	t.Parallel()

	m := money.FromInt

	tests := []struct {
		costs     []money.Money
		intervals int
		//
		want []CostInterval
	}{
		{
			costs:     []money.Money{m(1), m(2), m(3), m(4), m(5), m(6), m(7)},
			intervals: 2,
			//
			want: []CostInterval{
				{From: m(1), To: m(4) - 1},
				{From: m(4), To: m(7)},
			},
		},
		{
			costs:     []money.Money{m(1), m(2), m(3), m(4), m(5), m(6), m(7)},
			intervals: 3,
			//
			want: []CostInterval{
				{From: m(1), To: m(3) - 1},
				{From: m(3), To: m(5) - 1},
				{From: m(5), To: m(7)},
			},
		},
		{
			costs: []money.Money{
				// p5
				m(1),
				//
				m(2), m(3), m(4), m(5), m(6), m(7), m(8), m(9), m(10), m(11), m(12), m(13), m(14), m(15), m(16),
				m(17), m(18), m(19), m(20), m(21), m(22), m(23), m(24), m(25), m(26), m(27), m(28), m(29),
				// p95
				m(30),
			},
			intervals: 2,
			//
			want: []CostInterval{
				{From: m(2), To: m(16) - 1},
				{From: m(16), To: m(29)},
			},
		},
		{
			costs: []money.Money{
				// p5
				m(1), m(2),
				//
				m(3), m(4), m(5), m(6), m(7), m(8), m(9), m(10), m(11), m(12), m(13), m(14), m(15), m(16), m(17),
				m(18), m(19), m(20), m(21), m(22), m(23), m(24), m(25), m(26), m(27), m(28), m(29), m(30), m(31),
				m(32), m(33), m(34), m(35), m(36), m(37), m(38), m(39), m(40), m(41), m(42), m(43), m(44), m(45),
				m(46), m(47), m(48), m(49), m(50), m(51), m(52), m(53), m(54), m(55), m(56), m(57),
				// p95
				m(58), m(59), m(60),
			},
			intervals: 2,
			//
			want: []CostInterval{
				{From: m(3), To: m(30) - 1},
				{From: m(30), To: m(57)},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			got := prepareIntervals(tt.costs, tt.intervals)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetPercentileValue(t *testing.T) {
	t.Parallel()

	type test struct {
		p    int
		want money.Money
	}
	tests := []struct {
		data   []money.Money
		checks []test
	}{
		// Tests from https://en.wikipedia.org/wiki/Percentile#Worked_examples_of_the_nearest-rank_method
		{
			data: []money.Money{15, 20, 35, 40, 50},
			checks: []test{
				{p: 5, want: 15},
				{p: 30, want: 20},
				{p: 40, want: 20},
				{p: 50, want: 35},
				{p: 100, want: 50},
			},
		},
		{
			data: []money.Money{3, 6, 7, 8, 8, 10, 13, 15, 16, 20},
			checks: []test{
				{p: 25, want: 7},
				{p: 50, want: 8},
				{p: 75, want: 15},
				{p: 100, want: 20},
			},
		},
		{
			data: []money.Money{3, 6, 7, 8, 8, 9, 10, 13, 15, 16, 20},
			checks: []test{
				{p: 25, want: 7},
				{p: 50, want: 9},
				{p: 75, want: 15},
				{p: 100, want: 20},
			},
		},
		// Edge cases
		{
			data: []money.Money{1, 2, 3, 4, 5},
			checks: []test{
				{p: -2, want: 1},
				{p: 0, want: 1},
				{p: 200, want: 5},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			for _, check := range tt.checks {
				got := getPercentileValue(tt.data, check.p)
				require.Equal(t, check.want, got)
			}
		})
	}
}
