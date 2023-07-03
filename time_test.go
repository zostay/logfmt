package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertGenericTimestampToTime(t *testing.T) {
	testDates := map[string]time.Time{
		"2022-07-21T11:12:35.540Z":     time.Date(2022, 7, 21, 11, 12, 35, 540_000_000, time.UTC),
		"2022-07-27T13:09:39.381-0500": time.Date(2022, 7, 27, 13, 9, 39, 381_000_000, time.FixedZone("UTC-5", -5*60*60)),
	}

	for ds, dt := range testDates {
		lineData := map[string]any{
			"ts": ds,
		}

		convertGenericTimestampToTime(lineData, "ts")

		require.IsType(t, time.Time{}, lineData["ts"], "date %s parsed type", ds)

		dt = dt.UTC()
		lineData["ts"] = lineData["ts"].(time.Time).UTC()

		assert.Equal(t, dt, lineData["ts"], "date %s parsed", ds)
	}
}
