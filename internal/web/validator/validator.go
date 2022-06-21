package validator

import (
	"encoding/json"
	"reflect"

	"github.com/ShoshinNikita/budget-manager/v2/internal/pkg/errors"
)

type Validatable interface {
	IsValid() error
}

// Valid type indicates that the field must be checked by Validator
type Valid[T Validatable] struct {
	value T
}

var (
	_ json.Unmarshaler = (*Valid[Validatable])(nil)
	_ toValidate       = (*Valid[Validatable])(nil)
)

func (v *Valid[T]) Get() T {
	return v.value
}

func (v *Valid[T]) UnmarshalJSON(data []byte) error {
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	v.value = value
	return nil
}

func (v Valid[T]) isValid() error {
	return v.value.IsValid()
}

type toValidate interface {
	isValid() error
}

// Validator validates structure fields with type Valid[T].
//
// By default it uses 'json' tag for composing error messages. Tag name can be changed
// with SetTagName method. If no tag found, field name will be used
type Validator struct {
	tagName string
}

func NewValidator() *Validator {
	return &Validator{
		tagName: "json",
	}
}

func (v *Validator) SetTagName(name string) *Validator {
	v.tagName = name
	return v
}

func (v *Validator) Validate(s any) error {
	rv := reflect.ValueOf(s)
	rv = deref(rv)

	if rv.Kind() != reflect.Struct {
		return errors.Errorf("can validate only structs, got %s", rv.Kind())
	}

	rt := rv.Type()

	n := rv.NumField()
	for i := 0; i < n; i++ {
		ft := rt.Field(i)
		if !ft.IsExported() {
			continue
		}

		fv := rv.Field(i)

		// Type assertion has better performance than (reflect.Type).Implements
		if f, ok := fv.Interface().(toValidate); ok {
			if err := f.isValid(); err != nil {
				return v.buildError(ft, err)
			}
		}

		fv = deref(fv)
		if fv.Kind() == reflect.Struct {
			if err := v.Validate(fv.Interface()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v Validator) buildError(field reflect.StructField, err error) error {
	fieldName, ok := field.Tag.Lookup(v.tagName)
	if !ok {
		fieldName = field.Name
	}

	return errors.Wrapf(err, "invalid field %q", fieldName)
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	return v
}

var defaultValidator = NewValidator()

func Validate(s any) error {
	return defaultValidator.Validate(s)
}
