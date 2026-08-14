## Unreleased  2026-08-13

 * Fixed tty detection, which was inverted: `--color=auto` (the default) treated a terminal as a non-terminal and vice versa. As a workaround, `auto` had been forced to `on`, so ANSI escapes were written even when output was piped or redirected to a file.
 * `--color=auto` now detects the actual output destination, so `-o file` is no longer colorized when standard output happens to be a terminal. Use `--color=on` to force color when piping into a pager such as `less -R`.
 * Fixed `-o`/`--output`, which opened the output file read-only. Every write failed and the error was discarded, so the file was left empty and logfmt still exited 0. Appending with `-a` did not create the file if it was missing.
 * Fixed the error message for an unopenable output file, which formatted the nil file handle instead of the file name.
 * `--color=auto` now honors the custom palette from `.logfmt.yaml` instead of always using the default palette.

## Unreleased  2026-08-12

 * Fixed the CI coverage step, which filtered the package list on the wrong module path (`github.com/zostay/today`).
 * Merged the CI test and coverage steps into a single `go test -v ./... -coverprofile=coverage.out` run, so the suite is no longer run twice.

## Unreleased  2026-07-20

 * Merged Dependabot PR #51: chore(deps): bump actions/setup-go from 6 to 7

## Unreleased  2026-07-14

 * Merged Dependabot PR #49: chore(deps): bump golang.org/x/sys from 0.46.0 to 0.47.0

## Unreleased  2026-06-30

 * Merged Dependabot PR #45: chore(deps): bump golang.org/x/sys from 0.45.0 to 0.46.0
 * Merged Dependabot PR #46: chore(deps): bump actions/checkout from 6 to 7

## Unreleased  2026-05-30

 * Merged Dependabot PR #43: chore(deps): bump golang.org/x/sys from 0.44.0 to 0.45.0
 * Merged Dependabot PR #39: chore(deps): bump golang.org/x/sys from 0.43.0 to 0.44.0.

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
