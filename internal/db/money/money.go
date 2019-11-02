package money

// Money is a sum of money multiplied by presicion (100). It can be negative
type Money int64

// Convert functions

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
