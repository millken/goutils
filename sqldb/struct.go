package sqldb

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

type field struct {
	name  string
	field reflect.StructField
}

var cachedFields atomic.Value // map[reflect.Type][]field

func appendFields(fields []field, t reflect.Type, index []int) []field {
	for i, n := 0, t.NumField(); i < n; i++ {
		if f := t.Field(i); f.IsExported() {
			if len(index) > 0 {
				f.Index = append(index, f.Index...)
			}
			if f.Anonymous {
				if f.Type.Kind() == reflect.Struct {
					fields = appendFields(fields, f.Type, f.Index)
				}
			} else if s, ok := f.Tag.Lookup("sql"); ok {
				fields = append(fields, field{s, f})
			} else if s, ok := f.Tag.Lookup("db"); ok {
				fields = append(fields, field{s, f})
			} else {
				fields = append(fields, field{f.Name, f})
			}
		}
	}
	return fields
}

func fields(t reflect.Type) []field {
	cache, _ := cachedFields.Load().(map[reflect.Type][]field)
	fields, ok := cache[t]
	if !ok {
		fields = appendFields(nil, t, nil)

		newCache := make(map[reflect.Type][]field, len(cache)+1)
		for k, v := range cache {
			newCache[k] = v
		}
		newCache[t] = fields
		cachedFields.Store(newCache)
	}
	return fields
}

func baseType(t reflect.Type, expected reflect.Kind) (reflect.Type, error) {
	t = deref(t)
	if t.Kind() != expected {
		return nil, fmt.Errorf("expected %s but got %s", expected, t.Kind())
	}
	return t, nil
}
