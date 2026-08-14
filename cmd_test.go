package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupColorizerMode pins down which colorizer each --color value selects.
// Without this, restoring the old "auto is broken, force it on" workaround
// leaves the suite green.
func TestSetupColorizerMode(t *testing.T) {
	orig := colorize
	t.Cleanup(func() { colorize = orig })

	tests := map[string]any{
		"off":  &ColorOff{},
		"on":   &ColorOn{},
		"auto": &ColorAuto{},
		"":     &ColorAuto{},
	}

	for mode, want := range tests {
		colorize = mode

		// A buffer is never a terminal, so auto resolves to a ColorAuto with
		// color disabled rather than being rewritten to "on".
		colorizer := setupColorizer(&bytes.Buffer{})

		require.NotNil(t, colorizer, "colorizer built for %q", mode)
		assert.IsType(t, want, colorizer.PlainColorizer, "colorizer type for --color=%q", mode)
	}
}

// TestSetupColorizerAutoIsOffForNonTerminal is the end-to-end guard for the
// regression in issue #1: in the default mode, output that is not a terminal
// must carry no escape sequences.
func TestSetupColorizerAutoIsOffForNonTerminal(t *testing.T) {
	orig := colorize
	t.Cleanup(func() { colorize = orig })

	colorize = "auto"

	colorizer := setupColorizer(&bytes.Buffer{})

	assert.Equal(t, "boom", colorizer.C(ColorLevelError, "boom"), "no color when output is not a terminal")
}
