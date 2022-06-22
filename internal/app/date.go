package app

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

const dateStringLength = 10 // YYYY-MM-DD

type Date struct {
	Year  int
	Month time.Month
	Day   int
}

var (
	_ encoding.TextMarshaler   = (*Date)(nil)
	_ encoding.TextUnmarshaler = (*Date)(nil)
	_ json.Marshaler           = (*Date)(nil)
	_ json.Unmarshaler         = (*Date)(nil)
)

func (d Date) IsValid() error {
	t := time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
	if d.Year != t.Year() || d.Month != t.Month() || d.Day != t.Day() {
		return errors.Errorf("invalid date %s", d)
	}
	return nil
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Date) UnmarshalText(data []byte) error {
	wrapErr := func(err error) error {
		return errors.Wrapf(err, "invalid date %q", string(data))
	}

	if len(data) != dateStringLength {
		return wrapErr(errors.Errorf("expected %d characters", dateStringLength))
	}

	year, err := strconv.ParseInt(string(data[:4]), 10, 64)
	if err != nil {
		return wrapErr(errors.New("invalid year"))
	}
	month, err := strconv.ParseInt(string(data[5:7]), 10, 64)
	if err != nil {
		return wrapErr(errors.New("invalid month"))
	}
	day, err := strconv.ParseInt(string(data[8:]), 10, 64)
	if err != nil {
		return wrapErr(errors.New("invalid day"))
	}

	d.Year = int(year)
	d.Month = time.Month(month)
	d.Day = int(day)

	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	data = bytes.TrimPrefix(data, []byte(`"`))
	data = bytes.TrimSuffix(data, []byte(`"`))

	return d.UnmarshalText(data)
}
