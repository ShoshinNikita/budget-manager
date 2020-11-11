package money

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const precision = 100

// Money is an amount with precision 2
//
//   - FromInt(15) -> 15
//   - FromFloat(15.07) -> 15.07
//   - FromFloat(-15.073) -> -15.07
//   - FromFloat(15.078) -> 15.07
//
type Money int64

// FromInt converts int64 to Money
func FromInt(m int64) Money {
	return Money(m * precision)
}

// FromFloat converts float64 to Money
func FromFloat(m float64) Money {
	// We can't convert float64 to int64 with precision 2 by multiplying it by 100 because we can get something
	// like this: 17.83 * 100 = 1782.9999999999998
	//
	// We can use package 'github.com/shopspring/decimal', but it requires major refactoring.
	// So, as a temporary solution, we use this algorithm:
	//
	// 1. Convert float64 to string with fixed precision
	// 2. Remove decimal separator '.'
	// 3. Parse this strings as int64
	//

	// Use precision 3 instead of 2 because 'AppendFloat' rounds float64
	s := strconv.AppendFloat(nil, m, 'f', 3, 64)
	// Replace decimal separator with the first digit and first digit with the second one: 17.830 -> 178830 -> 178330
	s[len(s)-4], s[len(s)-3] = s[len(s)-3], s[len(s)-2]
	// Trim last 2 digits
	s = s[:len(s)-2]

	res, err := strconv.ParseInt(string(s), 10, 64)
	if err != nil {
		// Just in case
		panic(err)
	}

	return Money(res)
}

// Int converts Money to int64
func (m Money) Int() int64 {
	return int64(m) / precision
}

// Float converts Money to float64
func (m Money) Float() float64 {
	return float64(m) / precision
}

// String converts Money to string. Money is always formatted as a number with 2 digits
// after decimal point (123.45, 123.00 and etc.)
func (m Money) String() string {
	return fmt.Sprintf("%.2f", m.Float())
}

// Arithmetic operations

// Add returns sum of original and passed Moneys
func (m Money) Add(add Money) Money {
	return m + add
}

// Sub returns remainder after subtraction
func (m Money) Sub(sub Money) Money {
	return m - sub
}

// Div divides Money by n (if n <= 0, it panics)
func (m Money) Div(n int64) Money {
	if n <= 0 {
		panic("n must be greater than zero")
	}

	// Don't use Money.ToInt for better precision
	money := int64(m)
	return Money(money / n)
}

// Other

// Round is like 'math.Round'
func (m Money) Round() Money {
	if m == 0 {
		return m
	}

	mod := m % precision
	if mod == 0 {
		return m
	}

	m -= mod
	switch {
	case mod >= 50:
		m += precision
	case mod <= -50:
		m -= precision
	}

	return m
}

// Ceil is like 'math.Ceil'
func (m Money) Ceil() Money {
	if m == 0 {
		return m
	}

	mod := m % precision
	if mod == 0 {
		return m
	}

	m -= mod
	if m > 0 {
		m += precision
	}
	return m
}

// Floor is like 'math.Floor'
func (m Money) Floor() Money {
	if m == 0 {
		return m
	}

	mod := m % precision
	if mod == 0 {
		return m
	}

	m -= mod
	if m < 0 {
		m -= precision
	}
	return m
}

// Encoding and Decoding

var (
	_ json.Marshaler   = (*Money)(nil)
	_ json.Unmarshaler = (*Money)(nil)
)

func (m Money) MarshalJSON() ([]byte, error) {
	// Always format with 2 digits after decimal point (123.45, 123.00 and etc.)
	return []byte(m.String()), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	f, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return err
	}
	*m = FromFloat(f)
	return nil
}

var _ fmt.Formatter = (*Money)(nil)

// Format implements 'fmt.Formatter' interface. It divides a number in groups of three didits
// separated by thin space
func (m Money) Format(f fmt.State, c rune) {
	const thinSpace = " "

	str := m.String()

	switch c {
	case 'd':
		str = str[:len(str)-3]
		if str == "-0" {
			str = "0"
		}
	case 'f':
		// Do nothing
	default:
		var negative bool
		// There's a case when minus is separated by thin space (- 100 000.00).
		// So, trim it for a while.
		if str[0] == '-' {
			negative = true
			str = str[1:]
		}

		// This algorithm can be buggy because the string is changing in process, but
		// it works for 1000000000000.00 (one trillion must be enough for all cases) and
		// it is very simple. So, leave it as is.

		for i := len(str) - 6; i > 0; i -= 3 {
			// We don't use comma as a separator because:
			//
			//   The 22nd General Conference on Weights and Measures declared in 2003 that
			//   "the symbol for the decimal marker shall be either the point on the line or
			//   the comma on the line". It further reaffirmed that "numbers may be divided in
			//   groups of three in order to facilitate reading; neither dots nor commas are ever
			//   inserted in the spaces between groups"
			//
			// Source: https://en.wikipedia.org/wiki/Decimal_separator#Current_standards
			//
			// Use thin space ' ' instead (https://en.wikipedia.org/wiki/Thin_space)

			str = str[:i] + thinSpace + str[i:]
		}

		if negative {
			str = "-" + str
		}
	}

	f.Write([]byte(str))
}
