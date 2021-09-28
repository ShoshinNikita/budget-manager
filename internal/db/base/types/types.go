package types

import (
	"database/sql"
	"database/sql/driver"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

// Uint represents optional uint. It treats NULL as 0 and vice versa
type Uint uint

var (
	_ sql.Scanner   = (*Uint)(nil)
	_ driver.Valuer = (*Uint)(nil)
)

func (v *Uint) Scan(src interface{}) error {
	if src == nil {
		*v = 0
		return nil
	}

	switch src := src.(type) {
	case int64:
		*v = Uint(src)
	default:
		return errors.Errorf("couldn't scan uint from %T", src)
	}
	return nil
}

func (v Uint) Value() (driver.Value, error) {
	if v == 0 {
		return nil, nil
	}
	return int64(v), nil
}

// String represents optional string. It treats NULL as empty string and vice versa
type String string

var _ sql.Scanner = (*String)(nil)

func (v *String) Scan(src interface{}) error {
	if src == nil {
		*v = ""
		return nil
	}

	switch src := src.(type) {
	case string:
		*v = String(src)
	case []byte:
		*v = String(src)
	default:
		return errors.Errorf("couldn't scan string from %T", src)
	}
	return nil
}

func (v String) Value() (driver.Value, error) {
	if v == "" {
		return nil, nil
	}
	return string(v), nil
}
