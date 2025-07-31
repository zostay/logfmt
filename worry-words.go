package main

import (
	"bufio"
	"strings"
	"unicode"
)

type WorrySeverity int

const (
	WorryNone = 0 + iota
	WorryInfo
	WorryWarn
	WorryErr
	WorryCrit
)

// WorryWords is the list of words that trigger worry-word highlighting in messages.
var WorryWords = map[string]WorrySeverity{
	"500":                       WorryInfo,
	"503":                       WorryInfo,
	"404":                       WorryInfo,
	"400":                       WorryInfo,
	"401":                       WorryInfo,
	"warning":                   WorryWarn,
	"warn":                      WorryWarn,
	"error":                     WorryErr,
	"failure":                   WorryErr,
	"failed":                    WorryErr,
	"fail":                      WorryErr,
	"incorrect":                 WorryInfo,
	"invalid":                   WorryInfo,
	"certificate_verify_failed": WorryErr,
}

var worryLevelColors = map[WorrySeverity]ColorName{
	WorryNone: ColorNormal,
	WorryInfo: ColorWorryInfo,
	WorryWarn: ColorWorryWarn,
	WorryErr:  ColorWorryError,
	WorryCrit: ColorWorryCritical,
}

var invertWorryWords map[WorrySeverity][]string

func init() {
	invertWorryWords = make(map[WorrySeverity][]string)
	for word, sev := range WorryWords {
		if _, hasSev := invertWorryWords[sev]; !hasSev {
			invertWorryWords[sev] = []string{word}
			continue
		}
		invertWorryWords[sev] = append(invertWorryWords[sev], word)
	}
}

// setupWorries reconfigures the WorryWords, if the Worries are configured.
func setupWorries(c *Config) {
	if len(c.WorryWords) == 0 {
		return
	}

	ww := make(map[string]WorrySeverity)
	for sev, words := range c.WorryWords {
		for _, word := range words {
			ww[word] = worryWordConfigSeverity[sev]
		}
	}

	if len(ww) == 0 {
		return
	}

	WorryWords = ww
}

// IsWordCharacter matches letters, digits, marks, and combining punctuation.
// That is, given a rune, this returns true if and only if the rune is included
// in one or more of these Unicode range categories: L, N, M, and Pc.
func IsWordCharacter(c rune) bool {
	return unicode.Is(unicode.L, c) || unicode.Is(unicode.N, c) || unicode.Is(unicode.M, c) || unicode.Is(unicode.Pc, c)
}

// ScanWordsTheRightWay breaks up on word boundaries. A word boundary is defined
// as it is for \b in Perl regular expressions, i.e., the strings is split at any
// boundary between a word character (as defined by \w) and a non-word character
// (as defined as the negation of \w, a.k.a., \W). The character class of \w
// matches letters, digits, Unicode marks, and connector punctuation (like
// underscore). See IsWordCharacter.
func ScanWordsTheRightWay(
	data []byte,
	_ bool,
) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	dataR := []rune(string(data))
	isWord := IsWordCharacter(dataR[0])
	for i := range dataR[1:] {
		if IsWordCharacter(dataR[i]) != isWord {
			token = []byte(string(dataR[0:i]))
			advance = len(token)
			return
		}
	}

	advance = len(data)
	token = data
	return
}

func HighlightWorries(
	c *SugaredColorizer,
	msg string,
) string {
	msgOut := &strings.Builder{}
	msgReader := strings.NewReader(msg)
	scanner := bufio.NewScanner(msgReader)
	scanner.Split(ScanWordsTheRightWay)
	for scanner.Scan() {
		word := scanner.Text()
		worryLevel, ok := WorryWords[strings.ToLower(word)]
		if ok {
			color := worryLevelColors[worryLevel]
			msgOut.WriteString(c.C(color, word))
		} else {
			msgOut.WriteString(word)
		}
	}
	return msgOut.String()
}
