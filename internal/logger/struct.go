package logger

import (
	"fmt"
	"go/token"
	"reflect"
	"time"
)

//nolint:gochecknoglobals
var fmtStringer = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

// structToFields maps fields of a passed structure to Fields. It returns empty fields
// for a non-structure arguments.
func structToFields(i interface{}, namePrefix string) (fields Fields) {
	defer func() {
		if r := recover(); r != nil {
			fields = Fields{"struct_to_fields_panic": r, "type": fmt.Sprintf("%T", i)}
		}
	}()

	v := deref(reflect.ValueOf(i))
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()

	fieldCount := v.NumField()
	fields = make(Fields, fieldCount)
	for i := 0; i < fieldCount; i++ {
		name, skip := getName(t.Field(i), namePrefix)
		if skip {
			continue
		}

		var value interface{}

		f := v.Field(i)

		// Check whether the field implements 'fmt.Stringer' before 'deref' call
		// because 'String' method can be declared with a pointer receiver.
		if f.Type().Implements(fmtStringer) && (f.Kind() != reflect.Ptr || !f.IsNil()) {
			value = f.Interface().(fmt.Stringer).String()
		}

		f = deref(f)
		switch f := f.Interface().(type) { //nolint:gocritic
		case time.Time:
			value = f.Format(time.RFC3339)
		}

		if value != nil {
			fields[name] = value
			continue
		}

		// Fallback to basic types
		switch f.Kind() {
		case
			reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128,
			reflect.String,
			reflect.Array:

			fields[name] = f.Interface()

		case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface:
			var value interface{} = "<nil>"
			if !f.IsNil() {
				value = f.Interface()
			}
			fields[name] = value

		case reflect.Struct:
			for k, v := range structToFields(f.Interface(), "") {
				fields[name+"."+k] = v
			}
		}
	}
	return fields
}

func getName(f reflect.StructField, namePrefix string) (name string, skip bool) {
	if !token.IsExported(f.Name) {
		return "", true
	}
	name = f.Name
	if v := f.Tag.Get("json"); v != "" {
		name = v
	}
	if namePrefix != "" {
		name = namePrefix + "." + name
	}
	return name, false
}

// deref returns a pointer's value, traversing as many levels as needed
// until Value's kind is not reflect.Ptr or Value.IsNil returns true.
func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v
}
