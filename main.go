// Package main defines a log formatter for ingesting mixed text/jSON logs
// outputting something a little less ugly for humans to read.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "logfmt [ <input-file> ]",
		Short: "Format standard input for the named input-file",
		Args:  cobra.MaximumNArgs(1),
		Run:   FormatLogLines,
	}
	outputFile   string
	appendToFile bool
)

func init() {
	cmd.Flags().StringVarP(&outputFile, "output", "o", "-", "output file write to or - for standard output")
	cmd.Flags().BoolVarP(&appendToFile, "append", "a", false, "set to append to existing output")
}

// setupInput sets up the input file handle based on command-line input or returns err.
func setupInput(args []string) (io.Reader, error) {
	input := os.Stdin
	if len(args) > 0 {
		var err error
		input, err = os.Open(args[0])
		if err != nil {
			return nil, fmt.Errorf("Failed to open %q: %v", args[0], err)
		}
	}

	return input, nil
}

// setupOutput sets up the output file handle based on command-line input or returns err.
func setupOutput() (io.Writer, error) {
	var output io.Writer = os.Stdout
	if outputFile != "" && outputFile != "-" {
		var err error
		flags := os.O_CREATE
		mode := "create"
		if appendToFile {
			flags = os.O_APPEND
			mode = "append to"
		}
		output, err = os.OpenFile(outputFile, flags, 0644)
		if err != nil {
			return nil, fmt.Errorf("Failed to %s %q: %v", mode, output, err)
		}
	}

	return output, nil
}

// onErrReporteportAndQuit handles an error by writing it to standard error and exiting.
func onErrReportAndQuit(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

// FormatLogLines ingests a file or standard input, breaks the input into lines,
// and attempts to parse each line. If parsing is successful, if formats that
// log line prettily. If parsing fails, the line is output as-is. It keeps going
// until the input handle closes.
func FormatLogLines(command *cobra.Command, args []string) {
	input, err := setupInput(args)
	onErrReportAndQuit(err)

	output, err := setupOutput()
	onErrReportAndQuit(err)

	buffed := bufio.NewScanner(input)
	for buffed.Scan() {
		line := buffed.Text()
		lineData, err := parseLogLine(buffed.Bytes())
		if err != nil {
			outputRawLogLine(output, line)
		} else {
			outputFormattedLogLine(output, lineData)
		}
	}
}

// outputRawLogLine outputs a line that failed to be parsed.
func outputRawLogLine(out io.Writer, line string) {
	fmt.Fprintln(out, line)
}

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

// insertIndent is a function that will insert i spaces before each start of
// line in the given string.
func insertIndent(st string, i int) string {
	indent := strings.Repeat(" ", i)
	return indent + strings.ReplaceAll(st, "\n", "\n"+indent)
}

// parseLogLine tries to parse the log line as JSON and returns a generic map
// containing the result.
func parseLogLine(line []byte) (map[string]any, error) {
	lineData := map[string]any{}
	err := json.Unmarshal(line, &lineData)
	if err != nil {
		return nil, err
	}

	return lineData, nil
}

func main() {
	err := cmd.Execute()
	cobra.CheckErr(err)
}
