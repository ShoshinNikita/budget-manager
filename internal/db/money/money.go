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

// Add functions

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

// Sub functions

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
