package money

import (
	"bytes"
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Money struct {
	v decimal.Decimal
}

func FromInt(m int64) Money {
	return Money{decimal.NewFromInt(m)}
}

func FromFloat(m float64) Money {
	return Money{decimal.NewFromFloat(m)}
}

func FromString(s string) (Money, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Money{}, err
	}

	return Money{d}, nil
}

// Int converts Money to int64
func (m Money) Int() int64 {
	return m.v.IntPart()
}

// Float converts Money to float64
func (m Money) Float() float64 {
	f, _ := m.v.Float64()
	return f
}

// String converts Money to string. Money is always formatted as a number with 8 digits
// after decimal point (123.45000000, 123.00000000 and etc.)
func (m Money) String() string {
	return m.v.String()
}

// Add returns sum of original and passed Moneys
func (m Money) Add(add Money) Money {
	return Money{m.v.Add(add.v)}
}

// Sub returns remainder after subtraction
func (m Money) Sub(sub Money) Money {
	return Money{m.v.Sub(sub.v)}
}

// Div divides Money by n (if n <= 0, it panics)
func (m Money) Div(n int64) Money {
	if n == 0 {
		panic("n must be not equal to zero")
	}

	return Money{m.v.Div(decimal.NewFromInt(n))}
}

var (
	_ json.Marshaler   = (*Money)(nil)
	_ json.Unmarshaler = (*Money)(nil)
)

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	data = bytes.TrimPrefix(data, []byte(`"`))
	data = bytes.TrimSuffix(data, []byte(`"`))

	res, err := FromString(string(data))
	if err == nil {
		*m = res
	}
	return err
}
