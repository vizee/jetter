package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func nextKeyField(key string) (string, bool, string) {
	if len(key) > 0 && key[0] == '[' {
		i := 1
		for i < len(key) {
			if key[i] == ']' {
				return key[1:i], true, key[i+1:]
			}
			i++
		}
	}

	p := 0
	if len(key) > 0 && key[0] == '.' {
		p = 1
	}
	i := p
	for i < len(key) {
		if key[i] == '.' || key[i] == '[' {
			return key[p:i], false, key[i:]
		}
		i++
	}
	return key[p:], false, ""
}

type keyField struct {
	name  string
	index bool
}

func allKeyFields(key string) []keyField {
	var fields []keyField
	for key != "" {
		field, idx, rem := nextKeyField(key)
		fields = append(fields, keyField{
			name:  field,
			index: idx,
		})
		key = rem
	}
	return fields
}

func applyAssign(values any, fields []keyField, value string) (any, error) {
	if len(fields) == 0 {
		return value, nil
	}

	if values == nil {
		if fields[0].index {
			values = []any{}
		} else {
			values = map[string]any{}
		}
	}

	switch vs := values.(type) {
	case map[string]any:
		if fields[0].index {
			return nil, fmt.Errorf("Invalid field: [%s]", fields[0].name)
		}

		val, err := applyAssign(vs[fields[0].name], fields[1:], value)
		if err != nil {
			return nil, err
		}
		vs[fields[0].name] = val

		return vs, nil
	case []any:
		if !fields[0].index {
			return nil, fmt.Errorf("Invalid field: .%s", fields[0].name)
		}

		idx, err := strconv.ParseUint(fields[0].name, 10, 64)
		if err != nil {
			return nil, err
		}

		if idx >= uint64(len(vs)) {
			a := make([]any, idx+1)
			copy(a, vs)
			vs = a
		}

		val, err := applyAssign(vs[idx], fields[1:], value)
		if err != nil {
			return nil, err
		}
		vs[idx] = val

		return vs, nil
	default:
		return nil, fmt.Errorf("Cannot assign to values of %T", values)
	}
}

func loadValues(baseFile string, assigns []string) (any, error) {
	var values any
	if baseFile != "" {
		fileData, err := os.ReadFile(baseFile)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(fileData, &values)
		if err != nil {
			return nil, err
		}
	}

	for _, assign := range assigns {
		eq := strings.IndexByte(assign, '=')
		if eq < 0 {
			return nil, fmt.Errorf("Invalid assign expression: %s", assign)
		}
		fields := allKeyFields(assign[:eq])
		badField := false
		for i := range fields {
			if fields[i].name == "" {
				badField = true
			}
		}
		if len(fields) == 0 || badField || fields[0].index {
			return nil, fmt.Errorf("Invalid key field: %s", assign[:eq])
		}
		applied, err := applyAssign(values, fields, assign[eq+1:])
		if err != nil {
			return nil, err
		}
		values = applied
	}

	return values, nil
}
