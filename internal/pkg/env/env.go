package env

import (
	"encoding"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

//nolint:gochecknoglobals
var durationType = reflect.TypeOf((*time.Duration)(nil)).Elem()

//nolint:funlen
func Load(key string, target interface{}) error {
	envValue, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}

	if u, ok := target.(encoding.TextUnmarshaler); ok {
		return u.UnmarshalText([]byte(envValue))
	}

	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}
	value = value.Elem()
	if !value.CanSet() {
		return errors.New("target can't be set")
	}

	wrapParseErr := func(err error) error {
		return errors.Wrapf(err, "value for %q is invalid", key)
	}

	// Special cases
	if value.Type() == durationType {
		duration, err := time.ParseDuration(envValue)
		if err != nil {
			return wrapParseErr(err)
		}
		value.SetInt(int64(duration))
		return nil
	}

	switch value.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(envValue)
		if err != nil {
			return wrapParseErr(err)
		}
		value.SetBool(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(envValue, 10, 64)
		if err != nil {
			return wrapParseErr(err)
		}
		value.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(envValue, 10, 64)
		if err != nil {
			return wrapParseErr(err)
		}
		value.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(envValue, 64)
		if err != nil {
			return wrapParseErr(err)
		}
		value.SetFloat(v)

	case reflect.String:
		value.SetString(envValue)

	default:
		return errors.Errorf("target of kind %q is not supported", value.Kind())
	}

	return nil
}
