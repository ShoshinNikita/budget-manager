package money

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var (
	// For text/templates and html/templates
	_ fmt.Formatter = (*Money)(nil)
	// For json
	_ json.Marshaler   = (*Money)(nil)
	_ json.Unmarshaler = (*Money)(nil)
)

// Money is a sum of money multiplied by presicion (100). It can be negative
// Money is marshalled without multiplication:
//   - FromInt(15) -> 15
//   - FromFloat(15.07) -> 15.07
//   - FromFloat(15.073) -> 15.07
//   - FromFloat(15.078) -> 15.07
//
type Money int64

// -------------------------------------------------
// Convert functions
// -------------------------------------------------

const precision = 100

// FromInt converts int64 to Money
func FromInt(m int64) Money {
	return Money(m * precision)
}

// FromFloat converts float64 to Money
func FromFloat(m float64) Money {
	return Money(int64(m * precision))
}

// ToInt converts Money to int64
func (m Money) ToInt() int64 {
	return int64(m) / precision
}

// ToInt converts Money to float64
func (m Money) ToFloat() float64 {
	return float64(m) / precision
}

// -------------------------------------------------
// Add functions
// -------------------------------------------------

// Add returns sum of original and passed Moneys
func (m Money) Add(add Money) Money {
	return m + add
}

// AddInt returns sum of original money and passed int64
func (m Money) AddInt(add int64) Money {
	return m + FromInt(add)
}

// AddFloat returns sum of original money and passed float64
func (m Money) AddFloat(add float64) Money {
	return m + FromFloat(add)
}

// -------------------------------------------------
// Sub functions
// -------------------------------------------------

// Sub returns remainder after substraction
func (m Money) Sub(sub Money) Money {
	return m - sub
}

// SubInt returns remainder after substraction
func (m Money) SubInt(sub int64) Money {
	return m - FromInt(sub)
}

// SubFloat returns remainder after substraction
func (m Money) SubFloat(sub float64) Money {
	return m - FromFloat(sub)
}

// Divide divides Money by n (if n <= 0, it panics)
func (m Money) Divide(n int64) Money {
	if n <= 0 {
		panic("n must be greater than zero")
	}

	// Don't use Money.ToInt for better precision
	money := int64(m)
	return Money(money / n)
}

// -------------------------------------------------
// Marshalling and Unmarshalling
// -------------------------------------------------

func (m Money) MarshalJSON() ([]byte, error) {
	// Always format with 2 digits after decimal point (123.45, 123.00 and etc.)
	return []byte(fmt.Sprintf("%.2f", m.ToFloat())), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	f, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	*m = FromFloat(f)
	return nil
}

// Format implements 'fmt.Formatter' interface
func (m Money) Format(f fmt.State, c rune) {
	// Money.MarshalJSON always returns nil error
	data, _ := m.MarshalJSON()
	f.Write(data)
}
