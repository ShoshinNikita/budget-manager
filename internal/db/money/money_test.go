package money_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/ShoshinNikita/budget_manager/internal/db/money"
)

func TestConvertMoney(t *testing.T) {
	t.Run("int to int", testConvertMoney_IntToInt)
	t.Run("float to float", testConvertMoney_FloatToFloat)
	t.Run("int to float", testConvertMoney_IntToFloat)
	t.Run("float to int", testConvertMoney_FloatToInt)
}

func testConvertMoney_IntToInt(t *testing.T) { // nolint:stylecheck,golint
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
}

func testConvertMoney_FloatToFloat(t *testing.T) { // nolint:stylecheck,golint
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
}

func testConvertMoney_IntToFloat(t *testing.T) { // nolint:stylecheck,golint
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
}

func testConvertMoney_FloatToInt(t *testing.T) { // nolint:stylecheck,golint
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
}

func TestAdd(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		origin int64

		add Money
		res Money

		addInt int64
		resInt Money

		addFloat float64
		resFloat Money
	}{
		{
			origin: -150,
			add:    Money(2000), res: Money(-13000),
			addInt: 50, resInt: Money(-10000),
			addFloat: 15.53, resFloat: Money(-13447),
		},
		{
			origin: 0,
			add:    Money(2000), res: Money(2000),
			addInt: 53, resInt: Money(5300),
			addFloat: 15.53, resFloat: Money(1553),
		},
		{
			origin: 150,
			add:    Money(2000), res: Money(17000),
			addInt: 50, resInt: Money(20000),
			addFloat: 15.53, resFloat: Money(16553),
		},
		// Add negative
		{
			origin: 150,
			add:    Money(-2000), res: Money(13000),
			addInt: -50, resInt: Money(10000),
			addFloat: -15.53, resFloat: Money(13447),
		},
	}

	for _, tt := range tests {
		origin := FromInt(tt.origin)

		res := origin.Add(tt.add)
		require.Equal(tt.res, res)

		resInt := origin.AddInt(tt.addInt)
		require.Equal(tt.resInt, resInt)

		resFloat := origin.AddFloat(tt.addFloat)
		require.Equal(tt.resFloat, resFloat)
	}
}

func TestSub(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		origin int64

		sub Money
		res Money

		subInt int64
		resInt Money

		subFloat float64
		resFloat Money
	}{
		{
			origin: -150,
			sub:    Money(2000), res: Money(-17000),
			subInt: 50, resInt: Money(-20000),
			subFloat: 15.53, resFloat: Money(-16553),
		},
		{
			origin: 0,
			sub:    Money(2000), res: Money(-2000),
			subInt: 53, resInt: Money(-5300),
			subFloat: 15.53, resFloat: Money(-1553),
		},
		{
			origin: 150,
			sub:    Money(2000), res: Money(13000),
			subInt: 50, resInt: Money(10000),
			subFloat: 15.53, resFloat: Money(13447),
		},
		// Sub negative
		{
			origin: 150,
			sub:    Money(-2000), res: Money(17000),
			subInt: -50, resInt: Money(20000),
			subFloat: -15.53, resFloat: Money(16553),
		},
	}

	for _, tt := range tests {
		origin := FromInt(tt.origin)

		res := origin.Sub(tt.sub)
		require.Equal(tt.res, res)

		resInt := origin.SubInt(tt.subInt)
		require.Equal(tt.resInt, resInt)

		resFloat := origin.SubFloat(tt.subFloat)
		require.Equal(tt.resFloat, resFloat)
	}
}

func TestDivide(t *testing.T) {
	t.Run("common divides", func(t *testing.T) {
		require := require.New(t)

		tests := []struct {
			origin   Money
			n        int64
			res      Money
			resInt   int64
			resFloat float64
		}{
			{origin: FromInt(1500), n: 1, res: Money(150000), resInt: 1500, resFloat: 1500},
			{origin: FromInt(1500), n: 5, res: Money(30000), resInt: 300, resFloat: 300},
			{origin: FromInt(1500), n: 7, res: Money(21428), resInt: 214, resFloat: 214.28},
		}

		for _, tt := range tests {
			res := tt.origin.Divide(tt.n)
			require.Equal(tt.res, res)
			require.Equal(tt.resInt, res.ToInt())
			require.Equal(tt.resFloat, res.ToFloat())
		}
	})

	// Special case
	t.Run("divide by zero", func(t *testing.T) {
		require := require.New(t)

		defer func() {
			r := recover()
			require.NotNil(r)
		}()

		FromFloat(120.5).Divide(0)
	})
}

func TestJSON(t *testing.T) {
	t.Run("Marshal", testJSON_Marshal)
	t.Run("Unmarshal", testJSON_Unmarshal)
}

func testJSON_Marshal(t *testing.T) { // nolint:stylecheck,golint
	type testStruct struct {
		Money Money `json:"money"`
	}

	require := require.New(t)

	tests := []struct {
		input testStruct
		want  string
	}{
		{
			input: testStruct{Money: FromInt(357)},
			want:  `{"money":357.00}`,
		},
		{
			input: testStruct{Money: FromFloat(154.30)},
			want:  `{"money":154.30}`,
		},
		{
			input: testStruct{Money: FromFloat(0.07)},
			want:  `{"money":0.07}`,
		},
		{
			input: testStruct{Money: FromFloat(15.073)},
			want:  `{"money":15.07}`,
		},
		{
			input: testStruct{Money: FromFloat(15.078)},
			want:  `{"money":15.07}`,
		},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.input)
		require.Nil(err)
		require.Equal(tt.want, string(data))
	}
}

func testJSON_Unmarshal(t *testing.T) { // nolint:stylecheck,golint
	type testStruct struct {
		Money Money `json:"money"`
	}

	require := require.New(t)

	tests := []struct {
		input string
		want  testStruct
	}{
		{
			input: `{"money":357}`,
			want:  testStruct{Money: FromInt(357)},
		},
		{
			input: `{"money":154.30}`,
			want:  testStruct{Money: FromFloat(154.30)},
		},
		{
			input: `{"money":0.07}`,
			want:  testStruct{Money: FromFloat(0.07)},
		},
		{
			input: `{"money":"test"}`,
			want:  testStruct{Money: FromInt(0)},
		},
	}

	for _, tt := range tests {
		var res testStruct

		// Ignore errors
		_ = json.Unmarshal([]byte(tt.input), &res)
		require.Equal(tt.want, res)
	}
}

func TestFormat(t *testing.T) {
	require := require.New(t)

	tests := []struct {
		input  Money
		format string // default format is '%v'
		want   string
	}{
		{input: FromInt(357), want: "357.00"},
		{input: FromInt(-357), want: "-357.00"},
		{input: FromFloat(154.30), want: "154.30"},
		{input: FromFloat(-154.30), want: "-154.30"},
		{input: FromFloat(0.07), want: "0.07"},
		{input: FromFloat(15.073), want: "15.07"},
		{input: FromFloat(15.078), want: "15.07"},
		// Check grouping
		{input: FromInt(1_500), want: "1 500.00"},
		{input: FromInt(-1_500), want: "-1 500.00"},
		{input: FromInt(15_000), want: "15 000.00"},
		{input: FromInt(-15_000), want: "-15 000.00"},
		{input: FromInt(150_000), want: "150 000.00"},
		{input: FromInt(-150_000), want: "-150 000.00"},
		{input: FromFloat(1_500_000.05), want: "1 500 000.05"},
		{input: FromFloat(-1_500_000.05), want: "-1 500 000.05"},
		{input: FromFloat(15_000_000.05), want: "15 000 000.05"},
		{input: FromFloat(-15_000_000.05), want: "-15 000 000.05"},
		{input: FromFloat(150_000_000.05), want: "150 000 000.05"},
		{input: FromFloat(-150_000_000.05), want: "-150 000 000.05"},
		{input: FromFloat(1_500_000_000.05), want: "1 500 000 000.05"},
		{input: FromFloat(-1_500_000_000.05), want: "-1 500 000 000.05"},
		{input: FromFloat(15_000_000_000.05), want: "15 000 000 000.05"},
		{input: FromFloat(-15_000_000_000.05), want: "-15 000 000 000.05"},
		{input: FromFloat(150_000_000_000.00), want: "150 000 000 000.00"},
		{input: FromFloat(-150_000_000_000.00), want: "-150 000 000 000.00"},
		{input: FromInt(1_500_000_000_000), want: "1 500 000 000 000.00"},
		{input: FromInt(-1_500_000_000_000), want: "-1 500 000 000 000.00"},
		// Check formats
		{input: FromFloat(0.05), format: "%d", want: "0"},
		{input: FromFloat(-0.05), format: "%d", want: "0"},
		{input: FromFloat(0.05), format: "%f", want: "0.05"},
		{input: FromFloat(-0.05), format: "%f", want: "-0.05"},
		{input: FromFloat(1_500.25), format: "%d", want: "1500"},
		{input: FromFloat(-1_500.25), format: "%d", want: "-1500"},
		{input: FromFloat(1_500.25), format: "%f", want: "1500.25"},
		{input: FromFloat(-1_500.25), format: "%f", want: "-1500.25"},
	}

	for _, tt := range tests {
		var s string
		if tt.format == "" {
			tt.format = "%v"
		}

		s = fmt.Sprintf(tt.format, tt.input)
		require.Equal(tt.want, s)
	}
}
