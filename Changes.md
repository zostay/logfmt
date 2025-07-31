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
