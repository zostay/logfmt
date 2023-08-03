package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:   "logfmt [ <input-file> ]",
		Short: "Format standard input for the named input-file",
		Args:  cobra.MaximumNArgs(1),
		Run:   FormatLogLines,
	}
	outputFile             string
	appendToFile           bool
	colorize               string
	highlightWorryWords    bool
	experimentalAccessLogs bool
	tsField                string
	msgField               string
	msgFormat              string
	lvlField               string
	callerField            string
	trimFields             []string
)

func init() {
	cmd.Flags().StringVarP(&outputFile, "output", "o", "-", "output file write to or - for standard output")
	cmd.Flags().BoolVarP(&appendToFile, "append", "a", false, "set to append to existing output")
	cmd.Flags().StringVarP(&colorize, "color", "c", "auto", "set the colorize mode (auto, on, off)")
	cmd.Flags().BoolVar(&highlightWorryWords, "highlight-worry-words", true, "enable highlighting of worry-words")
	cmd.Flags().BoolVar(&experimentalAccessLogs, "experimental-access-logs", false, "enable access log parsing")
	cmd.Flags().StringVarP(&tsField, "timestamp-field", "t", "ts", "set the timestamp field name")
	cmd.Flags().StringVar(&msgField, "message-field", "msg", "set the message field name")
	cmd.Flags().StringVar(&lvlField, "level-field", "level", "set the level field name")
	cmd.Flags().StringVar(&callerField, "caller-field", "caller", "set the caller field name")
	cmd.Flags().StringVarP(&msgFormat, "message-format", "m", "", "set the message format using gotemplate")
	cmd.Flags().StringArrayVarP(&trimFields, "trim-field", "T", []string{"level", "msg", "stacktrace", "error"}, "set fields to trim from the output")
}

// setupInput sets up the input file handle based on command-line input or returns err.
func setupInput(args []string) (io.Reader, error) {
	input := os.Stdin
	if len(args) > 0 {
		var err error
		input, err = os.Open(args[0])
		if err != nil {
			return nil, fmt.Errorf("failed to open %q: %v", args[0], err)
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
			return nil, fmt.Errorf("failed to %s %q: %v", mode, output, err)
		}
	}

	return output, nil
}

// onErrReportAndQuit handles an error by writing it to standard error and exiting.
func onErrReportAndQuit(err error) {
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func setupColorizer() *SugaredColorizer {
	// FIXME tty detection is broken, so...
	if colorize == "auto" {
		colorize = "on"
	}

	var colorizer *SugaredColorizer
	if colorize == "off" {
		colorizer = NewSugaredColorizer(&ColorOff{})
	} else if colorize == "on" {
		colorizer = NewSugaredColorizer(NewColorOn(DefaultPalette))
	} else {
		colorizer = NewSugaredColorizer(NewColorAuto())
	}
	return colorizer
}

// FormatLogLines ingests a file or standard input, breaks the input into lines,
// and attempts to parse each line. If parsing is successful, if formats that
// log line prettily. If parsing fails, the line is output as-is. It keeps going
// until the input handle closes.
func FormatLogLines(_ *cobra.Command, args []string) {
	input, err := setupInput(args)
	onErrReportAndQuit(err)

	output, err := setupOutput()
	onErrReportAndQuit(err)

	colorizer := setupColorizer()

	buffed := bufio.NewScanner(input)
	for buffed.Scan() {
		line := buffed.Text()
		lineData, err := parseLogLine(buffed.Bytes(), tsField)
		if err != nil {
			outputRawLogLine(output, colorizer, line)
		} else {
			trimFields = append(trimFields, tsField)
			if msgFormat == "" {
				msgFormat = fmt.Sprintf("{{index . %q}}", msgField)
			}
			outputFormattedLogLine(output, colorizer, lineData, tsField, msgFormat, trimFields)
		}
	}
}
