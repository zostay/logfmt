package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

var (
	ErrUnparseable = errors.New("unable to parse log line")
)

type LineParser func([]byte, string) (map[string]any, error)

var lineParsers = []LineParser{
	parseJsonLogLine,
	parseZapConsoleLikeLogLine,
}

var lineParsersWithAccessLogs = []LineParser{
	parseJsonLogLine,
	parseAccessLogLine,
	parseZapConsoleLikeLogLine,
}

const (
	RFC3339NanoAlt = "2006-01-02T15:04:05.999999999Z0700"
)

func convertGenericTimestampToTime(lineData map[string]any, tsField string) {
	if _, ok := lineData[tsField]; !ok {
		return
	}

	tsf, err := getFloat64(lineData, tsField)
	if err == nil {
		s, n := math.Modf(tsf)
		lineData[tsField] = time.Unix(int64(s), int64(n))
		return
	}

	tss, err := getString(lineData, tsField)
	if err == nil {
		tryFormats := []string{time.RFC3339Nano, RFC3339NanoAlt}
		for _, tfmt := range tryFormats {
			t, err := time.Parse(tfmt, tss)
			if err != nil {
				continue
			}
			lineData[tsField] = t
			return
		}
	}

	lineData[tsField] = time.Time{}
}

// parseJsonLogLine tries to parse the log line as JSON and returns a generic map
// containing the result.
func parseJsonLogLine(line []byte, tsField string) (map[string]any, error) {
	lineData := map[string]any{}
	err := json.Unmarshal(line, &lineData)
	if err != nil {
		return nil, err
	}

	convertGenericTimestampToTime(lineData, tsField)

	return lineData, nil
}

var (
	MillisDateMatch = regexp.MustCompile(`^\d\.\d+[eE][+]\d+`)

	// why doesn't golang's PCRE support the /x option... STOP THE INSANITY YOU RE
	// IMPLEMENTERS! IT CAN BE READABLE IF YOU JUST IMPLEMENT THE READABILITY
	// IMPROVING FEATURES!!!!
	IstioDefaultLogLineMatch = regexp.MustCompile(`^\[(?P<startTime>[^\\]+)] "(?P<requestLine>[^"]+)" (?P<responseCode>\S+) (?P<responseFlags>\S+) (?P<responseCodeDetails>\S+) (?P<connectionTerminationDetails>\S+) "(?P<upstreamTransportFailureReason>[^"]+)" (?P<bytesRecieved>\S+) (?P<bytesSent>\S+) (?P<duration>\S+) (?P<responseUpstreamTime>\S+) "(?P<forwardedFor>[^"]+)" "(?P<userAgent>[^"]+)" "(?P<requestId>[^"]+)" "(?P<authority>[^"]+)" "(?P<upstreamHost>[^"]+)" (?P<upstreamCluster>\S+) (?P<upstreamLocalAddress>\S+) (?P<downstreamLocalAddress>\S+) (?P<downstreamRemoteAddress>\S+) (?P<requestedServerNames>\S+) (?P<routeName>\S+)$`)

	Word = regexp.MustCompile(`^\S+`)

	WS = " \t\n\r"
)

// TODO Support other access log formats

// parseAccessLogLine attempts to parse line as an Envoy Proxy-style access log.
func parseAccessLogLine(line []byte, tsField string) (map[string]any, error) {
	res := make(map[string]any, 30)
	if sm := IstioDefaultLogLineMatch.FindSubmatch(line); sm != nil {
		for _, name := range IstioDefaultLogLineMatch.SubexpNames()[1:] {
			i := IstioDefaultLogLineMatch.SubexpIndex(name)
			if name == "startTime" {
				ts, err := time.Parse("2006-01-02T15:04:05.000Z", string(sm[i]))
				if err == nil {
					res[tsField] = ts
				}
			} else if name == "requestLine" {
				res["msg"] = string(sm[i])
			} else {
				res[name] = string(sm[i])
			}
		}
	} else {
		return nil, fmt.Errorf("not an Envoy Proxy access log line")
	}

	return res, nil
}

// parseTimestamp parses a timestamp prefix from the line and returns it or
// returns an error.
func parseTimestamp(line []byte) (time.Time, []byte, error) {
	// look for the timestamp
	var tsbs []byte
	if tsis := MillisDateMatch.FindIndex(line); tsis != nil && tsis[0] == 0 {
		tsbs = line[tsis[0]:tsis[1]]
		line = bytes.TrimLeft(line[tsis[1]:], WS)
	}
	if tsbs == nil {
		return time.Time{}, nil, ErrUnparseable
	}

	tsf, err := strconv.ParseFloat(string(tsbs), 64)
	if err != nil {
		return time.Time{}, nil, err
	}

	s, n := math.Modf(tsf)
	return time.Unix(int64(s), int64(n)), line, nil
}

// parseWord parses the first word out of the input line or returns error if end
// of string has been reached.
func parseWord(line []byte) (string, []byte, error) {
	var lvbs []byte
	if lvis := Word.FindIndex(line); lvis != nil && lvis[0] == 0 {
		lvbs = line[lvis[0]:lvis[1]]
		line = bytes.TrimLeft(line[lvis[1]:], WS)
	}
	if lvbs == nil {
		return "", nil, ErrUnparseable
	}

	return string(lvbs), line, nil
}

// parseStructure parses the console logger structured bit from the end of line,
// which we find by looking for a } at the end and then working our way forward
// until we find the matching {. If we don't end up with a match, we return an
// error. If we do, we parse it as JSON. If that fails, we return an error. If it
// all succeeds, we return the structured data parsed out.
func parseStructure(line []byte) (map[string]any, []byte, error) {
	line = bytes.TrimRight(line, WS)
	if line[len(line)-1] != '}' {
		return nil, nil, ErrUnparseable
	}

	var i int
	c := 0
	finished := false
	for i = len(line) - 1; i >= 0; i-- {
		switch line[i] {
		case '}':
			c++
		case '{':
			c--
		}

		if c == 0 {
			finished = true
			break
		}
	}

	if finished {
		var structure map[string]any
		err := json.Unmarshal(line[i:], &structure)
		if err != nil {
			return nil, nil, err
		}

		return structure, line[:i], nil
	}

	return nil, nil, ErrUnparseable
}

// parseZapConsoleLikeLogLine parses the console encoder logger, which is used by
// the development logger configurations in Uber's zap logger.
func parseZapConsoleLikeLogLine(line []byte, tsField string) (map[string]any, error) {
	var (
		ts                        time.Time
		level, loggerName, caller string
		err                       error
		lineData                  map[string]any
	)

	ts, line, err = parseTimestamp(line)
	if err != nil {
		return nil, err
	}

	level, line, err = parseWord(line)
	if err != nil {
		return nil, err
	}

	loggerName, line, err = parseWord(line)
	if err != nil {
		return nil, err
	}

	caller, line, err = parseWord(line)
	if err != nil {
		return nil, err
	}

	// time level logger caller message structure

	if structure, remainingLine, err := parseStructure(line); err == nil {
		lineData = structure
		line = remainingLine
	} else {
		lineData = make(map[string]any, 5)
	}

	lineData[tsField] = ts
	lineData["level"] = level
	lineData["logger"] = loggerName
	lineData["caller"] = caller
	lineData["msg"] = string(line)

	return lineData, nil
}

// parseLogLine tries to parse the log line from whatever format it appears to be
// in, trying one parser after another until it hits the fallback parser.
func parseLogLine(line []byte, tsField string) (map[string]any, error) {
	lp := lineParsers
	if experimentalAccessLogs {
		lp = lineParsersWithAccessLogs
	}

	for _, lineParser := range lp {
		if lineData, err := lineParser(line, tsField); err == nil {
			return lineData, nil
		}
	}

	return nil, ErrUnparseable
}
