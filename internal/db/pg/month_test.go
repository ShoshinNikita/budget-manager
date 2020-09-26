package pg

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/money"
)

func TestRecomputeMonth(t *testing.T) {
	t.Parallel()

	toMoney := func(m int64) money.Money { //nolint:gocritic
		return money.FromInt(m)
	}

	tests := []struct {
		desc  string
		input Month
		want  Month
	}{
		{
			desc: "usual month",
			input: Month{
				Incomes: []Income{
					{Income: toMoney(700)},
					{Income: toMoney(150)},
					{Income: toMoney(150)},
				},
				MonthlyPayments: []MonthlyPayment{
					{Cost: toMoney(175)},
					{Cost: toMoney(25)},
				},
				//
				Days: []Day{
					{},
					{
						Spends: []Spend{{Cost: toMoney(99)}, {Cost: toMoney(1)}},
					},
					{
						Spends: []Spend{{Cost: toMoney(12)}},
					},
					{},
				},
			},
			want: Month{
				Incomes: []Income{
					{Income: toMoney(700)},
					{Income: toMoney(150)},
					{Income: toMoney(150)},
				},
				MonthlyPayments: []MonthlyPayment{
					{Cost: toMoney(175)},
					{Cost: toMoney(25)},
				},
				//
				DailyBudget: toMoney(200),
				Days: []Day{
					{
						Saldo: toMoney(200),
					},
					{
						Spends: []Spend{{Cost: toMoney(99)}, {Cost: toMoney(1)}},
						Saldo:  toMoney(300),
					},
					{
						Spends: []Spend{{Cost: toMoney(12)}},
						Saldo:  toMoney(488),
					},
					{
						Saldo: toMoney(688),
					},
				},
				//
				TotalIncome: toMoney(1000),
				TotalSpend:  toMoney(-312),
				Result:      toMoney(688),
			},
		},
		{
			desc: "negative spends (cashback)",
			input: Month{
				Incomes:         []Income{{Income: toMoney(1000)}},
				MonthlyPayments: nil,
				//
				Days: []Day{
					{},
					{
						Spends: []Spend{{Cost: toMoney(99)}, {Cost: toMoney(-99)}},
					},
					{
						Spends: []Spend{{Cost: toMoney(120)}},
					},
					{},
				},
			},
			want: Month{
				Incomes:         []Income{{Income: toMoney(1000)}},
				MonthlyPayments: nil,
				//
				DailyBudget: toMoney(250),
				Days: []Day{
					{
						Saldo: toMoney(250),
					},
					{
						Spends: []Spend{{Cost: toMoney(99)}, {Cost: toMoney(-99)}},
						Saldo:  toMoney(500),
					},
					{
						Spends: []Spend{{Cost: toMoney(120)}},
						Saldo:  toMoney(630),
					},
					{
						Saldo: toMoney(880),
					},
				},
				//
				TotalIncome: toMoney(1000),
				TotalSpend:  toMoney(-120),
				Result:      toMoney(880),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.desc, func(t *testing.T) {
			got := recomputeMonth(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}
