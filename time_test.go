package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertGenericTimestampToTime(t *testing.T) {
	testDates := map[string]time.Time{
		"2022-07-21T11:12:35.540Z": time.Date(2022, 7, 21, 11, 12, 35, 540_000_000, time.UTC),
	}

	for ds, dt := range testDates {
		lineData := map[string]any{
			"ts": ds,
		}

		convertGenericTimestampToTime(lineData)

		assert.Equalf(t, dt, lineData["ts"], "date %s parses", ds)
	}
}
