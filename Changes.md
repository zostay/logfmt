## 0.3.0  2026-08-13

 * Fixed tty detection for `--color=auto`, the default mode. The check was inverted, so a terminal was treated as a non-terminal and a pipe as a terminal. Because of that, `auto` had been forced to always colorize, and ANSI escape sequences were written even when output was piped or redirected into a file. Color is now enabled only when the output really is a terminal. To force color when piping into a pager such as `less -R`, use `--color=on`.
 * `--color=auto` now inspects the actual output destination, so `-o file` is no longer colorized just because standard output happens to be a terminal.
 * `--color=auto` now honors the custom palette from `.logfmt.yaml` instead of always using the built-in palette.
 * Fixed `-o`/`--output`, which opened the output file read-only. Every write failed, the error was discarded, and logfmt left an empty file behind while still exiting successfully. Appending with `-a` also failed to create the file when it did not already exist.
 * An unrecognized `--color` value, or `colorize` setting in `.logfmt.yaml`, is now an error listing the valid modes instead of being silently treated as `auto`. The value is checked before the output file is opened, so a typo no longer truncates an existing `-o` file.
 * Fixed the error message shown when the output file cannot be opened, which formatted the nil file handle rather than the file name.
 * Error messages written to standard error now end with a newline.
 * Updated golang.org/x/sys (v0.43.0 -> v0.47.0).
 * Fixed the CI coverage step, which filtered the package list on the wrong module path, and merged the duplicate test runs into one.

## 0.2.1  2026-03-09

 * Upgraded Go from v1.24 to v1.25.
 * Upgraded golangci-lint to v2.4.0 to support Go 1.25.
 * Updated software dependencies:
   * github.com/spf13/cobra (v1.9.1 -> v1.10.2)
   * github.com/spf13/viper (v1.20.1 -> v1.21.0)
   * golang.org/x/sys (v0.35.0 -> v0.42.0)

## 0.2.0  2025-07-31

 * Added the `.logfmt.yaml` configuration file. All command-line features are configurable through this file. In addition, the colors and worry words are configurable via this file as well.
 * Added extra information about configuration to `--help`
 * Added a new `--help-config` command describing the configuration in more detail.
 * Added a new `--init-config` command to create a basic configuration file anywhere.
 * Added a new `--init-config-home` command to create a basic configuration file in the current user's home directory.
 * Removed the `--message-format/-m` argument, which was never implemented.

## 0.1.0  2025-07-30

 * Add the `--extract-field/-X` option to name which fields are extracted from the JSON output and displayed inline (default: `error,stacktrace`).
 * Hide null values from extra fields section.
 * Add the `--show-null` option to show null values in extra fields section.
 * Updated software dependencies:
   * Go (v1.18 -> v1.24)
   * github.com/spf13/cobra (v1.8.1 -> v1.9.1)
   * github.com/spf13/pflag (v1.0.5 -> v1.0.6)
   * golang.org/x/sys (v0.30.0 -> v0.35.0)

## 0.0.1  2025-02-13

 * Initial release.
