package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// outputRawLogLine outputs a line that failed to be parsed.
func outputRawLogLine(out io.Writer, c *SugaredColorizer, line string) {
	_, _ = fmt.Fprintln(out, c.C(ColorNormal, line))
}

// outputFormattedLogLine will take a parsed log line and pretty print it. The
// format is "TS LEVEL MSG {EXTRA}". The TS is an RFC3339Nano formatted time
// stamp. The LEVEL is the log level. The MSG is the text of the message. The
// EXTRA is omitted if no additional fields are present. If additional fields
// are present, those are converted back to JSON and rendered. If a stacktrace
// is present, it will be output indented below the log line.
func outputFormattedLogLine(out io.Writer, c *SugaredColorizer, lineData map[string]any) {
	tsTimeStr := "0000-00-00T00:00:00.000000-00:00"
	if tsTime, err := getTime(lineData, "ts"); err == nil {
		tsTimeStr = tsTime.Format(time.RFC3339Nano)
		delete(lineData, "ts")
	}

	level, err := getString(lineData, "level")
	if err == nil {
		delete(lineData, "level")
	}
	level = strings.ToUpper(level)

	msg, err := getString(lineData, "msg")
	if err == nil {
		delete(lineData, "msg")
	}

	error, err := getString(lineData, "error")
	if err == nil {
		delete(lineData, "error")
	}

	st, err := getString(lineData, "stacktrace")
	if err == nil {
		delete(lineData, "stacktrace")
	}

	f := "%s %-6s %s"
	levelColorName := LevelToColorName(level)
	args := []any{
		c.C(ColorDateTime, tsTimeStr),
		c.C(levelColorName, level),
		c.C(ColorMessage, msg),
	}

	if len(lineData) > 0 {
		// If this has an error, we have something in the logs that can be
		// parsed from JSON but not put back into JSON? Seems unlikely.
		lineDataBytes, _ := json.Marshal(lineData)
		coloredDataBytes := colorizeDataBytes(c, lineDataBytes)

		f += " %s"
		args = append(args, c.C(ColorData, string(coloredDataBytes)))
	}

	f += "\n"
	_, _ = fmt.Fprintf(out, f, args...)

	if error != "" {
		fmt.Fprintln(out, c.C(ColorLevelError, strings.TrimSpace(error)))
	}

	if st != "" {
		_, _ = fmt.Fprintln(out, c.C(ColorStackTrace, insertIndent(st, 4)))
	}
}

var dataLiteral = regexp.MustCompile(`"(?:[^\\"]+|\\.)*"|'(?:[^\\']+|\\.)*'|\d+`)

// colorizeDataBytes finds strings and numbers and colorizes them using
// "data-literal."
func colorizeDataBytes(c *SugaredColorizer, bs []byte) []byte {
	return dataLiteral.ReplaceAllFunc(bs, func(mbs []byte) []byte {
		return []byte(c.C(ColorDataLiteral, string(mbs)))
	})
}

// insertIndent is a function that will insert i spaces before each start of
// line in the given string.
func insertIndent(st string, i int) string {
	indent := strings.Repeat(" ", i)
	return indent + strings.ReplaceAll(st, "\n", "\n"+indent)
}
