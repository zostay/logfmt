package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// outputRawLogLine outputs a line that failed to be parsed.
func outputRawLogLine(out io.Writer, line string) {
	fmt.Fprintln(out, line)
}

// outputFormattedLogLine will take a parsed log line and pretty print it. The
// format is "TS LEVEL MSG {EXTRA}". The TS is an RFC3339Nano formatted time
// stamp. The LEVEL is the log level. The MSG is the text of the message. The
// EXTRA is omitted if no additional fields are present. If additional fields
// are present, those are converted back to JSON and rendered. If a stacktrace
// is present, it will be output indented below the log line.
func outputFormattedLogLine(out io.Writer, lineData map[string]any) {
	tsTimeStr := "0000-00-00T00:00:00.000000-00:00"
	ts, err := getFloat64(lineData, "ts")
	if errors.Is(err, errType) {
		tsStr, err := getString(lineData, "ts")
		if err == nil {
			tsTimeStr = tsStr
			delete(lineData, "ts")
		}
	} else if err == nil {
		delete(lineData, "ts")

		if ts > 0 {
			micros := int64(ts * 1_000_000)
			tsTime := time.UnixMicro(micros)
			tsTimeStr = tsTime.Format(time.RFC3339Nano)
		}
	}

	level, err := getString(lineData, "level")
	if err == nil {
		delete(lineData, "level")
	}
	level = strings.ToUpper(level)

	msg, _ := getString(lineData, "msg")
	if err == nil {
		delete(lineData, "msg")
	}

	st, _ := getString(lineData, "stacktrace")
	if err == nil {
		delete(lineData, "stacktrace")
	}

	if len(lineData) > 0 {
		// If this errors, we have something in the logs that can be parsed from JSON
		// but not put back into JSON? Seems unlikely.
		lineDataBytes, _ := json.Marshal(lineData)
		fmt.Fprintf(out, "%s %-6s %s %s\n", tsTimeStr, level, msg, lineDataBytes)
	} else {
		fmt.Fprintf(out, "%s %-6s %s\n", tsTimeStr, level, msg)
	}

	if st != "" {
		fmt.Fprintln(out, insertIndent(st, 4))
	}
}
