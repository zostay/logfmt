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
	outputFile   string
	appendToFile bool
	colorize     string
)

func init() {
	cmd.Flags().StringVarP(&outputFile, "output", "o", "-", "output file write to or - for standard output")
	cmd.Flags().BoolVarP(&appendToFile, "append", "a", false, "set to append to existing output")
	cmd.Flags().StringVarP(&colorize, "color", "c", "auto", "set the colorize mode (auto, on, off)")
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

func setupColorizer() *SugaredColorizer {
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
func FormatLogLines(command *cobra.Command, args []string) {
	input, err := setupInput(args)
	onErrReportAndQuit(err)

	output, err := setupOutput()
	onErrReportAndQuit(err)

	colorizer := setupColorizer()

	buffed := bufio.NewScanner(input)
	for buffed.Scan() {
		line := buffed.Text()
		lineData, err := parseLogLine(buffed.Bytes())
		if err != nil {
			outputRawLogLine(output, colorizer, line)
		} else {
			outputFormattedLogLine(output, colorizer, lineData)
		}
	}
}
