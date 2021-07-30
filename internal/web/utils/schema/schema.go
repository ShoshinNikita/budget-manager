//nolint:gochecknoglobals,gochecknoinits
package schema

import (
	"net/url"
	"reflect"
	"time"

	"github.com/gorilla/schema"
)

var (
	encoder = schema.NewEncoder()
	decoder = schema.NewDecoder()
)

var invalidDecodeValue = reflect.Value{}

func init() {
	encoder.SetAliasTag("json")
	encoder.RegisterEncoder(time.Time{}, func(v reflect.Value) string {
		t := v.Interface().(time.Time) //nolint:forcetypeassert
		return t.Format(time.RFC3339)
	})

	decoder.SetAliasTag("json")
	decoder.IgnoreUnknownKeys(true)
	decoder.RegisterConverter(time.Time{}, func(s string) reflect.Value {
		if time, err := time.Parse(time.RFC3339, s); err == nil {
			return reflect.ValueOf(time)
		}
		return invalidDecodeValue
	})
}

func Encode(src interface{}, dst url.Values) error {
	return encoder.Encode(src, dst)
}

func Decode(dst interface{}, src url.Values) error {
	return decoder.Decode(dst, src)
}
