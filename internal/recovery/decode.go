package recovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

func DecodeDeclaration(data []byte) (Declaration, error) {
	var d Declaration
	if err := decode(data, MaxDeclarationBytes, &d); err != nil {
		return d, err
	}
	return d, d.Validate()
}
func DecodeResult(data []byte) (Result, error) {
	var r Result
	if err := decode(data, MaxResultBytes, &r); err != nil {
		return r, err
	}
	return r, r.Validate()
}

// The normal Go decoder accepts duplicate/case-insensitive keys and nulls.
// Inspect the complete bounded document against its exact JSON field names first.
func decode(data []byte, limit int, target any) error {
	bad := fmt.Errorf("invalid recovery JSON: exact fields, non-null values and one UTF-8 object required")
	if len(data) > limit || !utf8.Valid(data) || !validSurrogates(data) {
		return bad
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	value, err := readValue(d, 0)
	if err != nil {
		return bad
	}
	if _, err = d.Token(); err != io.EOF {
		return bad
	}
	if !shape(value, reflect.TypeOf(target).Elem()) {
		return bad
	}
	// Re-encode inspected values so integer-valued version numbers match Schema.
	if m, ok := value.(map[string]any); ok {
		if version, exists := m["agent_version"]; exists && !text(version.(string), 1, 128) {
			return bad
		}
		if n, ok := m["schema_version"].(json.Number); ok {
			if !versionOne(n.String()) {
				return bad
			}
			m["schema_version"] = 1
		}
	}
	b, err := json.Marshal(value)
	if err != nil || json.Unmarshal(b, target) != nil {
		return bad
	}
	return nil
}

func readValue(d *json.Decoder, depth int) (any, error) {
	if depth > 6 {
		return nil, fmt.Errorf("too deep")
	}
	t, err := d.Token()
	if err != nil || t == nil {
		return nil, fmt.Errorf("missing value")
	}
	switch t {
	case json.Delim('{'):
		m := map[string]any{}
		for d.More() {
			key, e := d.Token()
			if e != nil {
				return nil, e
			}
			k, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("key")
			}
			if _, ok = m[k]; ok {
				return nil, fmt.Errorf("duplicate")
			}
			v, e := readValue(d, depth+1)
			if e != nil {
				return nil, e
			}
			m[k] = v
		}
		_, err = d.Token()
		return m, err
	case json.Delim('['):
		a := []any{}
		for d.More() {
			v, e := readValue(d, depth+1)
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

func shape(v any, t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}
		fields := map[string]reflect.Type{}
		required := map[string]bool{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("json")
			name, optional, _ := bytes.Cut([]byte(tag), []byte(","))
			fields[string(name)] = f.Type
			required[string(name)] = string(optional) != "omitempty"
		}
		for k, value := range m {
			typ, ok := fields[k]
			if !ok || !shape(value, typ) {
				return false
			}
			delete(required, k)
		}
		for _, needed := range required {
			if needed {
				return false
			}
		}
		return true
	case reflect.Map:
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}
		for _, item := range m {
			if !shape(item, t.Elem()) {
				return false
			}
		}
		return true
	case reflect.Slice:
		a, ok := v.([]any)
		if !ok {
			return false
		}
		for _, item := range a {
			if !shape(item, t.Elem()) {
				return false
			}
		}
		return true
	case reflect.String:
		_, ok := v.(string)
		return ok
	case reflect.Int:
		_, ok := v.(json.Number)
		return ok
	}
	return false
}

// encoding/json replaces unpaired escaped UTF-16 surrogates. IDs must not be
// silently rewritten, so reject them before Token decoding.
func validSurrogates(data []byte) bool {
	inside := false
	for i := 0; i < len(data); i++ {
		if data[i] == '"' {
			inside = !inside
			continue
		}
		if !inside || data[i] != 92 {
			continue
		}
		i++
		if i >= len(data) {
			return false
		}
		if data[i] != 'u' {
			continue
		}
		if i+4 >= len(data) {
			return false
		}
		n, err := strconv.ParseUint(string(data[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if n >= 0xdc00 && n <= 0xdfff {
			return false
		}
		if n < 0xd800 || n > 0xdbff {
			continue
		}
		if i+6 >= len(data) || data[i+1] != 92 || data[i+2] != 'u' {
			return false
		}
		low, err := strconv.ParseUint(string(data[i+3:i+7]), 16, 16)
		if err != nil || low < 0xdc00 || low > 0xdfff {
			return false
		}
		i += 6
	}
	return true
}

// Check numeric equality to 1 in linear space/time. Untrusted exponent text must
// not allocate a huge integer even though the complete document is bounded.
func versionOne(n string) bool {
	exponent := int64(0)
	if i := strings.IndexAny(n, "eE"); i >= 0 {
		var err error
		exponent, err = strconv.ParseInt(n[i+1:], 10, 32)
		if err != nil {
			return false
		}
		n = n[:i]
	}
	fraction := 0
	if i := strings.IndexByte(n, '.'); i >= 0 {
		fraction = len(n) - i - 1
		n = n[:i] + n[i+1:]
	}
	n = strings.TrimLeft(n, "0")
	significant := strings.TrimRight(n, "0")
	return significant == "1" && exponent-int64(fraction)+int64(len(n)-len(significant)) == 0
}
