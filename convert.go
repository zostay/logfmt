package main

import (
	"errors"
	"time"
)

var (
	errType = errors.New("value is of wrong type")
	errSet  = errors.New("value is not set")
)

// getTime retrieves a time.Time from the given generic map.
func getTime(d map[string]any, k string) (time.Time, error) {
	if ti, ok := d[k]; ok {
		if t, tok := ti.(time.Time); tok {
			return t, nil
		} else {
			return time.Time{}, errType
		}
	} else {
		return time.Time{}, errSet
	}
}

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

// getTime retrieves a time value from the given

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
