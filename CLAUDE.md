# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is `logfmt`, a Go CLI tool that formats mixed text/JSON log output to be more human-readable. It ingests log streams (typically from kubectl logs or similar) and colorizes/prettifies JSON logs while passing through plain text lines unchanged.

## Architecture

The application uses a modular architecture with clear separation of concerns:

- **main.go**: Entry point using cobra command framework
- **cmd.go**: Command-line interface setup and main processing loop
- **config.go**: Configuration management using Viper (file + environment variables)
- **parse.go**: Multiple log line parsers (JSON, Zap console-like, access logs)
- **output.go**: Formatted output rendering with colorization
- **color.go**: Color palette and terminal color management
- **convert.go**: Data type conversion utilities (time parsing, etc.)
- **worry-words.go**: Keyword highlighting for error/warning terms

The core flow: config load → input → parse (try multiple parsers) → format → colorize → output

## Key Dependencies

- **github.com/spf13/cobra**: CLI framework for commands and flags
- **github.com/spf13/viper**: Configuration management with file and environment variable support
- **github.com/wayneashleyberry/truecolor**: Terminal color support
- **github.com/stretchr/testify**: Testing framework

## Development Commands

### Building and Installing
```bash
go install ./           # Install to $GOPATH/bin
make install            # Same as above
```

### Testing
```bash
go test ./...           # Run all tests
make test              # Same as above
go test -v ./...       # Verbose test output
```

### Linting
```bash
golangci-lint run      # Run linter (requires golangci-lint v2.3.0+)
```

The project uses golangci-lint in CI with specific version v2.3.0. Local development should use compatible versions.

### Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out  # View in browser
```

## Running the Tool

```bash
# From source
go run . [flags] [input-file]

# Typical usage with kubectl
kubectl logs deploy/app -f | logfmt

# Common flags
logfmt --color=on --highlight-worry-words=true --timestamp-field=ts
```

## Code Patterns

- All source files are in the main package (single-binary CLI tool)
- Uses embed directive for version.txt
- Error handling follows Go conventions with explicit error returns
- Parsing uses a chain-of-responsibility pattern (multiple parsers tried in sequence)
- Colorization uses a strategy pattern with different color backends

## Testing Notes

- Tests use testify for assertions
- Coverage target is currently 0% (REQUIRED_COVERAGE=0 in CI)
- Time parsing tests are in time_test.go
- CI runs tests with coverage reporting but no enforcement yet

## Configuration

The tool supports configuration via `.logfmt.yaml` files using Viper:

### Configuration File Locations
1. **Upward search**: Starting from current directory, searches upward for `.logfmt.yaml`
2. **Home directory**: `~/.logfmt.yaml`

### Environment Variables
All config options can be set via environment variables with `LOGFMT_` prefix:
```bash
export LOGFMT_COLORIZE=on
export LOGFMT_TIMESTAMP_FIELD=timestamp
```

### Configuration Options
- **Output**: `output_file`, `append_to_file`, `colorize`
- **Processing**: `highlight_worry_words`, `experimental_access_logs`, `show_null`
- **Fields**: `timestamp_field`, `message_field`, `level_field`, `caller_field`
- **Field Arrays**: `trim_fields`, `extract_fields`
- **Colors**: Custom color palette via `colors` map (hex, RGB formats supported)

### Example Configuration
```yaml
colorize: "on"
timestamp_field: "ts"
colors:
  level-error: "#ff0000"  # hex format
  level-warn: "255,255,0"  # RGB format
```