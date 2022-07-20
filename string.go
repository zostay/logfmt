package main

import (
	"errors"
	"strings"
)

var (
	errType = errors.New("value is of wrong type")
	errSet  = errors.New("value is not set")
)

// getFloat64 retrieves a float64 value from the given generic map.
func getFloat64(d map[string]any, k string) (float64, error) {
	if fi, ok := d[k]; ok {
		if f, tok := fi.(float64); tok {
			return f, nil
		} else {
			return 0, errType
		}
	} else {
		return 0, errSet
	}
}

// getString retrieves a string value from the given generic map.
func getString(d map[string]any, k string) (string, error) {
	if si, ok := d[k]; ok {
		if s, tok := si.(string); tok {
			return s, nil
		} else {
			return "", errType
		}
	} else {
		return "", errSet
	}
}

// insertIndent is a function that will insert i spaces before each start of
// line in the given string.
func insertIndent(st string, i int) string {
	indent := strings.Repeat(" ", i)
	return indent + strings.ReplaceAll(st, "\n", "\n"+indent)
}
