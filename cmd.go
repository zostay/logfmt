package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

//go:embed version.txt
var Version string

var (
	cmd                    *cobra.Command
	config                 *Config
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
	version                bool
	showNull               bool
	extractFields          []string
	helpConfig             bool
	initConfig             string
	initConfigHome         bool
)

func init() {
	// Initialize command
	cmd = &cobra.Command{
		Use:   "logfmt [ <input-file> ]",
		Short: "Format standard input for the named input-file",
		Long:  `Format mixed text/JSON log output to be more human-readable.`,
		Args:  cobra.MaximumNArgs(1),
		Run:   formatLogLines,
	}

	origFunc := cmd.UsageFunc()
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		err := origFunc(cmd)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), `
Configuration:
  Use .logfmt.yaml files (searched upward from current dir, then home dir)
  or LOGFMT_* environment variables. Use --help-config for details.
`,
		)
		return err
	})

	// Load configuration first
	var err error
	config, err = LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		config = DefaultConfig()
	}

	// Define flags with defaults from config
	cmd.Flags().StringVarP(&outputFile, "output", "o", config.OutputFile, "output file write to or - for standard output")
	cmd.Flags().BoolVarP(&appendToFile, "append", "a", config.AppendToFile, "set to append to existing output")
	cmd.Flags().StringVarP(&colorize, "color", "c", config.Colorize, "set the colorize mode (auto, on, off)")
	cmd.Flags().BoolVar(&highlightWorryWords, "highlight-worry-words", config.HighlightWorryWords, "enable highlighting of worry-words")
	cmd.Flags().BoolVar(&experimentalAccessLogs, "experimental-access-logs", config.ExperimentalAccessLogs, "enable access log parsing")
	cmd.Flags().StringVarP(&tsField, "timestamp-field", "t", config.TimestampField, "set the timestamp field name")
	cmd.Flags().StringVar(&msgField, "message-field", config.MessageField, "set the message field name")
	cmd.Flags().StringVar(&lvlField, "level-field", config.LevelField, "set the level field name")
	cmd.Flags().StringVar(&callerField, "caller-field", config.CallerField, "set the caller field name")
	cmd.Flags().StringArrayVarP(&trimFields, "trim-field", "T", config.TrimFields, "set fields to trim from the output")
	cmd.Flags().BoolVar(&version, "version", false, "print the version and exit")
	cmd.Flags().BoolVar(&showNull, "show-null", config.ShowNull, "show null values in output")
	cmd.Flags().StringArrayVarP(&extractFields, "extract-field", "X", config.ExtractFields, "set fields to extract from the output for display")
	cmd.Flags().BoolVar(&helpConfig, "help-config", false, "show comprehensive configuration help")
	cmd.Flags().StringVar(&initConfig, "init-config", "", "initialize configuration file with specified filename")
	cmd.Flags().BoolVar(&initConfigHome, "init-config-home", false, "initialize configuration file in home directory (~/.logfmt.yaml)")
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

	// Get custom palette from config
	palette := DefaultPalette
	if config != nil {
		if customPalette, err := config.GetCustomPalette(); err == nil {
			palette = customPalette
		} else {
			fmt.Fprintf(os.Stderr, "Warning: failed to load custom colors: %v\n", err)
		}
	}

	var colorizer *SugaredColorizer
	switch colorize {
	case "off":
		colorizer = NewSugaredColorizer(&ColorOff{})
	case "on":
		colorizer = NewSugaredColorizer(NewColorOn(palette))
	default:
		colorizer = NewSugaredColorizer(NewColorAuto())
	}
	return colorizer
}

// formatLogLines ingests a file or standard input, breaks the input into lines,
// and attempts to parse each line. If parsing is successful, if formats that
// log line prettily. If parsing fails, the line is output as-is. It keeps going
// until the input handle closes.
func formatLogLines(cmd *cobra.Command, args []string) {
	if version {
		fmt.Printf("logfmt v%s\n", Version)
		os.Exit(0)
	}

	if helpConfig {
		PrintConfigHelp()
		os.Exit(0)
	}

	if cmd.Flags().Changed("init-config") {
		if err := InitializeConfigFile(initConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if initConfigHome {
		if err := InitializeConfigFile(""); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	setupWorries(config)

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
