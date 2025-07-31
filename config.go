package main

import (
	"fmt"
	gc "image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration structure for logfmt
type Config struct {
	OutputFile             string              `yaml:"output_file" mapstructure:"output_file"`
	AppendToFile           bool                `yaml:"append_to_file" mapstructure:"append_to_file"`
	Colorize               string              `yaml:"colorize" mapstructure:"colorize"`
	HighlightWorryWords    bool                `yaml:"highlight_worry_words" mapstructure:"highlight_worry_words"`
	ExperimentalAccessLogs bool                `yaml:"experimental_access_logs" mapstructure:"experimental_access_logs"`
	TimestampField         string              `yaml:"timestampe_field" mapstructure:"timestamp_field"`
	MessageField           string              `yaml:"message_field" mapstructure:"message_field"`
	LevelField             string              `yaml:"level_field" mapstructure:"level_field"`
	CallerField            string              `yaml:"caller_field" mapstructure:"caller_field"`
	TrimFields             []string            `yaml:"trim_fields" mapstructure:"trim_fields"`
	ShowNull               bool                `yaml:"show_null" mapstructure:"show_null"`
	ExtractFields          []string            `yaml:"extract_fields" mapstructure:"extract_fields"`
	Colors                 map[string]string   `yaml:"colors" mapstructure:"colors"`
	WorryWords             map[string][]string `yaml:"worries" mapstructure:"worries"`
}

var worryWordConfigSeverity = map[string]WorrySeverity{
	"none": WorryNone,
	"info": WorryInfo,
	"warn": WorryWarn,
	"err":  WorryErr,
	"crit": WorryCrit,
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	c := &Config{
		OutputFile:             "-",
		AppendToFile:           false,
		Colorize:               "auto",
		HighlightWorryWords:    true,
		ExperimentalAccessLogs: false,
		TimestampField:         "ts",
		MessageField:           "msg",
		LevelField:             "level",
		CallerField:            "caller",
		TrimFields:             []string{"level", "msg", "stacktrace", "error"},
		ShowNull:               false,
		ExtractFields:          []string{"error", "stacktrace"},
		Colors:                 make(map[string]string),
	}

	c.WorryWords = make(map[string][]string, len(worryWordConfigSeverity))
	for csev, wsev := range worryWordConfigSeverity {
		c.WorryWords[csev] = invertWorryWords[wsev]
	}

	return c
}

// LoadConfig loads configuration from files and environment variables
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set config file name and type
	v.SetConfigName(".logfmt")
	v.SetConfigType("yaml")

	// Add config search paths
	if err := addConfigPaths(v); err != nil {
		return nil, fmt.Errorf("failed to add config paths: %w", err)
	}

	// Set environment variable prefix
	v.SetEnvPrefix("LOGFMT")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	config := DefaultConfig()
	setViperDefaults(v, config)

	// Read config file (if it exists)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// addConfigPaths adds configuration file search paths:
// 1. Current directory and upward search
// 2. Home directory
func addConfigPaths(v *viper.Viper) error {
	// Add current directory and walk upward
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk upward from current directory
	dir := currentDir
	for {
		v.AddConfigPath(dir)
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	// Add home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	v.AddConfigPath(homeDir)

	return nil
}

// setViperDefaults sets default values in viper from config struct
func setViperDefaults(v *viper.Viper, config *Config) {
	v.SetDefault("output_file", config.OutputFile)
	v.SetDefault("append_to_file", config.AppendToFile)
	v.SetDefault("colorize", config.Colorize)
	v.SetDefault("highlight_worry_words", config.HighlightWorryWords)
	v.SetDefault("experimental_access_logs", config.ExperimentalAccessLogs)
	v.SetDefault("timestamp_field", config.TimestampField)
	v.SetDefault("message_field", config.MessageField)
	v.SetDefault("level_field", config.LevelField)
	v.SetDefault("caller_field", config.CallerField)
	v.SetDefault("trim_fields", config.TrimFields)
	v.SetDefault("show_null", config.ShowNull)
	v.SetDefault("extract_fields", config.ExtractFields)
	v.SetDefault("colors", config.Colors)
	v.SetDefault("worry_words", config.WorryWords)
}

// GetCustomPalette creates a color palette from the configuration
func (c *Config) GetCustomPalette() (Palette, error) {
	palette := make(Palette)

	// Start with default palette
	for name, color := range DefaultPalette {
		palette[name] = color
	}

	// Override with custom colors from config
	for colorName, colorValue := range c.Colors {
		color, err := parseColor(colorValue)
		if err != nil {
			return nil, fmt.Errorf("invalid color %q for %q: %w", colorValue, colorName, err)
		}
		palette[ColorName(colorName)] = color
	}

	return palette, nil
}

// parseColor parses a color string in various formats:
// - hex: "#ff0000" or "ff0000"
// - rgb: "rgb(255,0,0)" or "255,0,0"
func parseColor(colorStr string) (gc.Color, error) {
	colorStr = strings.TrimSpace(colorStr)

	// Handle hex colors
	if strings.HasPrefix(colorStr, "#") {
		colorStr = colorStr[1:]
	}

	// Try hex format (6 characters)
	if len(colorStr) == 6 {
		var r, g, b uint8
		if _, err := fmt.Sscanf(colorStr, "%02x%02x%02x", &r, &g, &b); err == nil {
			return RGB(r, g, b), nil
		}
	}

	// Try RGB format
	if strings.HasPrefix(colorStr, "rgb(") && strings.HasSuffix(colorStr, ")") {
		colorStr = strings.TrimPrefix(colorStr, "rgb(")
		colorStr = strings.TrimSuffix(colorStr, ")")
	}

	// Parse comma-separated RGB values
	if strings.Contains(colorStr, ",") {
		var r, g, b uint8
		if _, err := fmt.Sscanf(colorStr, "%d,%d,%d", &r, &g, &b); err == nil {
			return RGB(r, g, b), nil
		}
	}

	return nil, fmt.Errorf("unsupported color format: %s", colorStr)
}

// PrintConfigHelp prints comprehensive configuration help
func PrintConfigHelp() {
	fmt.Print(`Configuration Help for logfmt
=============================

logfmt supports configuration via .logfmt.yaml files and environment variables.

CONFIGURATION FILE LOCATIONS:
  1. Current directory and upward search for .logfmt.yaml
  2. Home directory ~/.logfmt.yaml

ENVIRONMENT VARIABLES:
  All options can be set with LOGFMT_ prefix:
    LOGFMT_COLORIZE=on
    LOGFMT_TIMESTAMP_FIELD=ts

CONFIGURATION OPTIONS:

Output Options:
  output_file: "-"              # Output file ("-" for stdout)  
  append_to_file: false         # Append to output file
  colorize: "auto"              # Color mode: "auto", "on", "off"

Processing Options:
  highlight_worry_words: true   # Highlight error/warning keywords
  experimental_access_logs: false  # Enable access log parsing
  show_null: false              # Show null values in output

Field Configuration:
  timestamp_field: "ts"         # Timestamp field name
  message_field: "msg"          # Message field name  
  level_field: "level"          # Log level field name
  caller_field: "caller"        # Caller info field name

Field Arrays:
  trim_fields:                  # Fields to remove from JSON output
    - "level"
    - "msg"
    - "stacktrace" 
    - "error"
  
  extract_fields:               # Fields to extract and display separately
    - "error"
    - "stacktrace"

Custom Worry Words:             # Words that get highlighted
  worries:
    info:                       # highlighted via worry-info
      - invalid
    warn:                       # highlighted via worry-warn
      - warning
    err:                        # highlighted via worry-err
      - error
      - failure
      - failed
      - fail
    crit:                       # highlighted via worry-crit
      - fatal

Custom Colors:
  colors:                       # Override default color palette
    normal: "#dddddd"           # Normal text
    "date/time": "#dddddd"      # Timestamp color
    level-debug: "#6666ff"      # DEBUG level
    level-info: "#14ffff"       # INFO level  
    level-warn: "#ffff00"       # WARN level
    level-error: "#ffd700"      # ERROR level
    level-dpanic: "#ff5f00"     # DPANIC level
    level-fatal: "#ff0000"      # FATAL level
    message: "#ffffff"          # Log message text
    stacktrace: "#ff6666"       # Stack traces
    data: "#aaaaaa"             # JSON data
    data-literal: "#88aa88"     # JSON literals
    worry-info: "#14ffff"       # Worry words (info)
    worry-err: "#ff0000"        # Worry words (error)
    worry-warn: "#ffff00"       # Worry words (warning)
    worry-crit: "#ff5f00"       # Worry words (critical)
    extracted: "#ff9999"        # Extracted field content

COLOR FORMATS:
  - Hex: "#ff0000" or "ff0000"
  - RGB: "rgb(255,0,0)" or "255,0,0"

EXAMPLE CONFIGURATION FILE (.logfmt.yaml):
  colorize: "on"
  timestamp_field: "timestamp"
  highlight_worry_words: true
  trim_fields:
    - "level"
    - "msg"
  colors:
    level-error: "#ff0000"
    level-warn: "255,255,0"

INITIALIZATION:
  Create a default config file:
    logfmt --init-config                    # Creates ~/.logfmt.yaml
    logfmt --init-config myproject.yaml     # Creates myproject.yaml

`)
}

// InitializeConfigFile creates a configuration file with default values
func InitializeConfigFile(filename string) error {
	// Default to home directory if no filename provided
	if filename == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		filename = filepath.Join(homeDir, ".logfmt.yaml")
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("configuration file already exists: %s", filename)
	}

	fh, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}
	defer fh.Close()

	enc := yaml.NewEncoder(fh)
	enc.SetIndent(2)

	// Create the configuration content, using current defaults
	err = enc.Encode(DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to encode default config: %w", err)
	}

	fmt.Printf("Configuration file created: %s\n", filename)
	fmt.Printf(`
This file is completely configured for current defaults. It is recommended that 
you edit this file and remove any configuration that you don't care about. That 
way, you'll benefit from improvements to the defaults in future releases.
`)
	return nil
}
