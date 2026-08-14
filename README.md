# logfmt

A simple mixed text/JSON log formatter. It makes logs a little prettier and
colorizes them. I don't like reading JSON, and this makes it so I don't have to,
mostly.

Point a stream of logs at it — from `kubectl`, a file, or a pipe — and it
reformats each JSON line into something readable, colorizes it, and passes
anything it can't parse through untouched.

## What it does

Given this:

```
{"ts":"2026-08-13T14:22:03.117Z","level":"info","msg":"listening for connections","caller":"server/main.go:84","addr":":8080","tls":true}
{"ts":"2026-08-13T14:22:09.482Z","level":"warn","msg":"upstream returned 503, retrying","caller":"proxy/client.go:212","upstream":"cart-svc","attempt":2}
{"ts":"2026-08-13T14:22:10.004Z","level":"error","msg":"request failed","caller":"proxy/client.go:240","error":"dial tcp 10.0.0.7:443: connection refused","stacktrace":"proxy.(*Client).Do\n\tproxy/client.go:240\nserver.handle\n\tserver/main.go:118"}
Starting up: no JSON here, just a plain line
```

you get this:

```
2026-08-13T14:22:03.117Z INFO   listening for connections {"addr":":8080","caller":"server/main.go:84","tls":true}
2026-08-13T14:22:09.482Z WARN   upstream returned 503, retrying {"attempt":2,"caller":"proxy/client.go:212","upstream":"cart-svc"}
2026-08-13T14:22:10.004Z ERROR  request failed {"caller":"proxy/client.go:240"}
    dial tcp 10.0.0.7:443: connection refused
    proxy.(*Client).Do
    	proxy/client.go:240
    server.handle
    	server/main.go:118
Starting up: no JSON here, just a plain line
```

In a terminal it is colorized: the level is colored by severity, JSON literals
are dimmed, and worry words like `503` and `failed` are highlighted inside the
message.

Note what happened to each line:

- The timestamp, level, and message were pulled out to the front, and the
  remaining fields were re-rendered as compact JSON at the end.
- `error` and `stacktrace` were extracted and printed indented underneath.
- The last line was not JSON, so it was passed through unchanged.

## Installation

```bash
go install github.com/zostay/logfmt@latest
```

Or download a prebuilt binary for Linux or macOS from the
[releases page](https://github.com/zostay/logfmt/releases).

## Usage

```bash
# The usual case
kubectl logs deploy/app -f | logfmt

# From a file
logfmt app.log

# From a pipe
cat app.log | logfmt

# To a file instead of the terminal
logfmt -o pretty.log app.log
```

Run `logfmt -h` for the flag list and `logfmt --help-config` for the full
configuration reference.

## Input formats

Each line is tried against the parsers in order, and the first one that succeeds
wins. If none do, the line is printed as-is — so a stream that mixes JSON logs
with plain startup banners and stack traces stays readable.

1. **JSON** — one JSON object per line, the common structured-logging format.
2. **zap console** — the console encoder used by Uber's
   [zap](https://github.com/uber-go/zap) development logger, of the form
   `<epoch> <level> <logger> <caller> <message> {<fields>}`.
3. **Envoy/Istio access logs** — off by default, enable with
   `--experimental-access-logs`. Only the default Envoy Proxy access log format
   is recognized.
4. **Anything else** — passed through unchanged, still worry-word highlighted.

Timestamps are recognized as RFC 3339, RFC 3339 with a numeric zone offset,
Python `logging` format (`2006-01-02 15:04:05,999`), or a numeric epoch.

## Color

`--color` takes `auto` (the default), `on`, or `off`.

In `auto` mode, color is emitted only when the output is a terminal, so piping
into a file or another program produces clean text with no escape sequences. If
you are piping into a pager and want to keep the color, force it:

```bash
kubectl logs deploy/app -f | logfmt --color=on | less -R
```

The palette is configurable — see [Configuration](#configuration).

## Worry words

Words that suggest something is wrong are highlighted inside the message, at
four severities: `info`, `warn`, `err`, and `crit`. Out of the box this covers
things like `404`, `500`, `503`, `warning`, `error`, `failed`, and `invalid`.
Matching is case-insensitive and on whole words only.

Turn it off with `--highlight-worry-words=false`, or replace the word list in
the config file:

```yaml
worries:
  info: ["404", "429", deprecated]
  warn: [warning, retrying]
  err: [error, failed, panic]
  crit: [fatal, corrupted]
```

Defining any `worries` replaces the built-in list rather than adding to it.

## Fields

logfmt needs to know which keys hold the timestamp, level, message, and caller.
The defaults suit zap-style logs:

| Flag | Default | Purpose |
| --- | --- | --- |
| `-t`, `--timestamp-field` | `ts` | Timestamp, moved to the front of the line |
| `--level-field` | `level` | Log level, upper-cased and colored by severity |
| `--message-field` | `msg` | The message text |
| `--caller-field` | `caller` | Caller / source location |

Two more control what happens to the rest of the fields:

- `-T`, `--trim-field` — drop these from the trailing JSON. Defaults to `level`,
  `msg`, `stacktrace`, and `error`, since those are already shown elsewhere.
- `-X`, `--extract-field` — print these indented on their own lines below the
  entry, instead of inline. Defaults to `error` and `stacktrace`.

For logs that use different names:

```bash
logfmt -t timestamp --level-field severity --message-field message \
       -T severity -T message app.log
```

Two things to know about these flags:

- **`-T` and `-X` replace the defaults, they do not add to them.** Passing
  `-T noisy` alone means `level` and `msg` are no longer trimmed and will
  reappear in the trailing JSON. Repeat the flag to list everything you want:
  `-T level -T msg -T noisy`.
- **Renaming a field does not update the trim list.** That is why the example
  above repeats `severity` and `message` as `-T` values — otherwise they show up
  both at the front of the line and again in the JSON tail.

By default, fields whose value is `null` are omitted from the trailing JSON.
Pass `--show-null` to keep them.

## Configuration

Every flag can be set in a `.logfmt.yaml` file, so you don't have to retype them.
logfmt searches upward from the current directory for `.logfmt.yaml`, then falls
back to `~/.logfmt.yaml`.

**Only the first file found is used — configuration files do not merge.** A
`.logfmt.yaml` in your project directory replaces `~/.logfmt.yaml` outright
rather than overriding individual keys, so anything you want to keep has to be
repeated in the project file.

Write a starter config with the current defaults:

```bash
logfmt --init-config-home          # writes ~/.logfmt.yaml
logfmt --init-config ./.logfmt.yaml
```

A small example:

```yaml
colorize: auto
timestamp_field: ts
highlight_worry_words: true

trim_fields: [level, msg, stacktrace, error]
extract_fields: [error, stacktrace]

colors:
  level-error: "#ff0000"
  level-warn: "255,255,0"
```

Colors accept hex (`#ff0000` or `ff0000`) or RGB (`rgb(255,0,0)` or `255,0,0`).
The palette keys are `normal`, `date/time`, `level-debug`, `level-info`,
`level-warn`, `level-error`, `level-dpanic`, `level-fatal`, `message`,
`stacktrace`, `data`, `data-literal`, `worry-info`, `worry-warn`, `worry-err`,
`worry-crit`, and `extracted`.

Any option can also be set through the environment with a `LOGFMT_` prefix:

```bash
export LOGFMT_COLORIZE=on
export LOGFMT_TIMESTAMP_FIELD=timestamp
```

Run `logfmt --help-config` for the complete list with descriptions.

## Flags

```
  -a, --append                      set to append to existing output
      --caller-field string         set the caller field name (default "caller")
  -c, --color string                set the colorize mode (auto, on, off) (default "auto")
      --experimental-access-logs    enable access log parsing
  -X, --extract-field stringArray   set fields to extract from the output for display (default [error,stacktrace])
  -h, --help                        help for logfmt
      --help-config                 show comprehensive configuration help
      --highlight-worry-words       enable highlighting of worry-words (default true)
      --init-config string          initialize configuration file with specified filename
      --init-config-home            initialize configuration file in home directory (~/.logfmt.yaml)
      --level-field string          set the level field name (default "level")
      --message-field string        set the message field name (default "msg")
  -o, --output string               output file write to or - for standard output (default "-")
      --show-null                   show null values in output
  -t, --timestamp-field string      set the timestamp field name (default "ts")
  -T, --trim-field stringArray      set fields to trim from the output (default [level,msg,stacktrace,error])
      --version                     print the version and exit
```

## Contributing

Issues and pull requests are welcome at
[github.com/zostay/logfmt](https://github.com/zostay/logfmt).

```bash
make test              # or: go test ./...
golangci-lint run      # requires golangci-lint v2.4.0+
make install           # install to $GOPATH/bin
```

## License

Copyright (c) 2022 Sterling Hanenkamp

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
