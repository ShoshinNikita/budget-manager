package money_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ShoshinNikita/budget_manager/internal/db/money"
)

func TestConvertMoney(t *testing.T) {
	t.Run("int to int", func(t *testing.T) {
		require := require.New(t)

		tests := []struct {
			in        int64
			res       Money
			converted int64
		}{
			{in: -20, res: Money(-2000), converted: -20},
			{in: 0, res: Money(0), converted: 0},
			{in: 15, res: Money(1500), converted: 15},
			{in: 1000000, res: Money(100000000), converted: 1000000},
		}

		for _, tt := range tests {
			res := FromInt(tt.in)
			require.Equal(tt.res, res)
			require.Equal(tt.converted, res.ToInt())
		}
	})

	t.Run("float to float", func(t *testing.T) {
		require := require.New(t)

		tests := []struct {
			in        float64
			res       Money
			converted float64
		}{
			{in: -20.50, res: Money(-2050), converted: -20.50},
			{in: 0, res: Money(0), converted: 0},
			{in: 0.75, res: Money(75), converted: 0.75},
			{in: 15.30, res: Money(1530), converted: 15.30},
			{in: 1000000.87, res: Money(100000087), converted: 1000000.87},
		}

		for _, tt := range tests {
			res := FromFloat(tt.in)
			require.Equal(tt.res, res)
			require.Equal(tt.converted, res.ToFloat())
		}
	})

	t.Run("int to float", func(t *testing.T) {
		require := require.New(t)

		tests := []struct {
			in        int64
			res       Money
			converted float64
		}{
			{in: -20, res: Money(-2000), converted: -20},
			{in: 0, res: Money(0), converted: 0},
			{in: 15, res: Money(1500), converted: 15},
			{in: 1000000, res: Money(100000000), converted: 1000000},
		}

		for _, tt := range tests {
			res := FromInt(tt.in)
			require.Equal(tt.res, res)
			require.Equal(tt.converted, res.ToFloat())
		}
	})

	t.Run("float to int", func(t *testing.T) {
		require := require.New(t)

		tests := []struct {
			in        float64
			res       Money
			converted int64
		}{
			{in: -20.5, res: Money(-2050), converted: -20},
			{in: 0.30, res: Money(30), converted: 0},
			{in: 15, res: Money(1500), converted: 15},
			{in: 1000000.87, res: Money(100000087), converted: 1000000},
		}

		for _, tt := range tests {
			res := FromFloat(tt.in)
			require.Equal(tt.res, res)
			require.Equal(tt.converted, res.ToInt())
		}
	})
}
