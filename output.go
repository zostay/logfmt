package main

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// outputRawLogLine outputs a line that failed to be parsed.
func outputRawLogLine(out io.Writer, c *SugaredColorizer, line string) {
	if highlightWorryWords {
		line = HighlightWorries(c, line)
	}
	_, _ = fmt.Fprintln(out, c.C(ColorNormal, line))
}

// outputFormattedLogLine will take a parsed log line and pretty print it. The
// format is "TS LEVEL MSG {EXTRA}". The TS is an RFC3339Nano formatted time
// stamp. The LEVEL is the log level. The MSG is the text of the message. The
// EXTRA is omitted if no additional fields are present. If additional fields
// are present, those are converted back to JSON and rendered. If a stacktrace
// is present, it will be output indented below the log line.
func outputFormattedLogLine(
	out io.Writer,
	c *SugaredColorizer,
	lineData map[string]any,
	tsField, msgFormat string,
	trimFields []string,
) {
	tsTimeStr := "0000-00-00T00:00:00.000000-00:00"
	if tsTime, err := getTime(lineData, tsField); err == nil {
		tsTimeStr = tsTime.Format(time.RFC3339Nano)
	}

	level, _ := getString(lineData, lvlField)
	level = strings.ToUpper(level)

	sw := &strings.Builder{}
	msgT := template.Must(template.New(msgField).Parse(msgFormat))
	_ = msgT.Execute(sw, lineData)
	msg := sw.String()
	// msg, _ := getString(lineData, "msg")
	// if err == nil {
	// 	delete(lineData, "msg")
	// }

	errorStr, _ := getString(lineData, "error")
	st, _ := getString(lineData, "stacktrace")

	for _, field := range trimFields {
		delete(lineData, field)
	}

	if highlightWorryWords {
		msg = HighlightWorries(c, msg)
	}

	f := "%s %-6s %s"
	levelColorName := LevelToColorName(level)
	args := []any{
		c.C(ColorDateTime, tsTimeStr),
		c.C(levelColorName, level),
		c.C(ColorMessage, msg),
	}

	if len(lineData) > 0 {
		if !showNull {
			keepData := make(map[string]any, len(lineData))
			for k, v := range lineData {
				if v != nil {
					keepData[k] = v
				}
			}
			lineData = keepData
		}

		// If this has an error, we have something in the logs that can be
		// parsed from JSON but not put back into JSON? Seems unlikely.
		lineDataBytes, _ := json.Marshal(lineData)
		coloredDataBytes := colorizeDataBytes(c, lineDataBytes)

		f += " %s"
		args = append(args, c.C(ColorData, string(coloredDataBytes)))
	}

	f += "\n"
	_, _ = fmt.Fprintf(out, f, args...)

	if errorStr != "" {
		_, _ = fmt.Fprintln(out, c.C(ColorLevelError, strings.TrimSpace(errorStr)))
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
