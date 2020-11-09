package statistics

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

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
