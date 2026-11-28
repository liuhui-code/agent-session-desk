package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"unicode/utf8"
)

// JSON is exact and non-null. Go's decoder alone accepts duplicate and
// case-insensitive aliases, which can make review and execution disagree.
func inspect(data []byte) (any, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("invalid UTF-8 update document")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	v, err := jsonValue(d, 0)
	if err != nil {
		return nil, err
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, fmt.Errorf("trailing update JSON")
	}
	return v, nil
}
func jsonValue(d *json.Decoder, depth int) (any, error) {
	if depth > 16 {
		return nil, fmt.Errorf("update JSON too deep")
	}
	t, err := d.Token()
	if err != nil {
		return nil, fmt.Errorf("null or missing update value")
	}
	switch t {
	case json.Delim('{'):
		m := map[string]any{}
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return nil, err
			}
			name, ok := key.(string)
			if !ok || seen[strings.ToLower(name)] {
				return nil, fmt.Errorf("duplicate update key")
			}
			seen[strings.ToLower(name)] = true
			v, err := jsonValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			m[name] = v
		}
		_, err = d.Token()
		return m, err
	case json.Delim('['):
		var a []any
		for d.More() {
			v, e := jsonValue(d, depth+1)
			if e != nil {
				return nil, e
			}
			a = append(a, v)
		}
		_, err = d.Token()
		return a, err
	}
	return t, nil
}
func exactFields(v any, t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			fields[name] = f.Type
		}
		for n, value := range m {
			field, ok := fields[n]
			if !ok || !exactFields(value, field) {
				return false
			}
		}
	case reflect.Map:
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}
		for _, value := range m {
			if !exactFields(value, t.Elem()) {
				return false
			}
		}
	case reflect.String:
		_, ok := v.(string)
		return ok
	case reflect.Bool:
		_, ok := v.(bool)
		return ok
	case reflect.Int, reflect.Int64:
		_, ok := v.(json.Number)
		return ok
	}
	return true
}
