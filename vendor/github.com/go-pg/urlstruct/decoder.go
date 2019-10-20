package urlstruct

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

type ScannerFunc func(v reflect.Value, values []string) error

var globalDecoder decoder

func DescribeStruct(typ reflect.Type) *StructInfo {
	return globalDecoder.DescribeStruct(typ)
}

func Unmarshal(values url.Values, strct interface{}) error {
	return globalDecoder.Unmarshal(values, strct)
}

type decoder struct {
	m sync.Map
}

func (f *decoder) DescribeStruct(typ reflect.Type) *StructInfo {
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("got %s, wanted %s", typ.Kind(), reflect.Struct))
	}

	if v, ok := f.m.Load(typ); ok {
		return v.(*StructInfo)
	}

	meta := newStructInfo(typ)
	if v, loaded := f.m.LoadOrStore(typ, meta); loaded {
		return v.(*StructInfo)
	}
	return meta
}

// Unmarshal unmarshals url values into the struct.
func (f *decoder) Unmarshal(values url.Values, strct interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(strct))
	meta := f.DescribeStruct(v.Type())

	var maps map[string][]string
	for name, values := range values {
		if name, key, ok := mapKey(name); ok {
			if maps == nil {
				maps = make(map[string][]string)
			}
			maps[name] = append(maps[name], key, values[0])
			continue
		}

		err := meta.decode(v, name, values)
		if err != nil {
			return err
		}
	}

	for name, values := range maps {
		err := meta.decode(v, name, values)
		if err != nil {
			return nil
		}
	}

	for _, f := range meta.unmarshalers {
		fv := v.FieldByIndex(f.Index)
		if fv.Kind() == reflect.Struct {
			fv = fv.Addr()
		} else if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}

		u := fv.Interface().(Unmarshaler)
		if err := u.UnmarshalValues(values); err != nil {
			return err
		}
	}

	if meta.isUnmarshaler {
		return strct.(Unmarshaler).UnmarshalValues(values)
	}
	return nil
}

func mapKey(s string) (name string, key string, ok bool) {
	ind := strings.IndexByte(s, '[')
	if ind == -1 || s[len(s)-1] != ']' {
		return "", "", false
	}
	key = s[ind+1 : len(s)-1]
	if key == "" {
		return "", "", false
	}
	name = s[:ind]
	return name, key, true
}
