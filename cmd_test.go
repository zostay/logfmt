package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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

// TestSetupOutputWritesToFile guards against the output file being opened
// read-only: writes then fail with EBADF, and because the callers in output.go
// discard write errors, -o produced an empty file and exited 0.
func TestSetupOutputWritesToFile(t *testing.T) {
	origFile, origAppend := outputFile, appendToFile
	t.Cleanup(func() { outputFile, appendToFile = origFile, origAppend })

	path := filepath.Join(t.TempDir(), "out.log")
	outputFile, appendToFile = path, false

	output, err := setupOutput()
	require.NoError(t, err, "setup output")

	_, err = fmt.Fprintln(output, "written")
	require.NoError(t, err, "write to output file")
	require.NoError(t, output.(*os.File).Close(), "close output file")

	content, err := os.ReadFile(path)
	require.NoError(t, err, "read back output file")
	assert.Equal(t, "written\n", string(content), "output reaches the file")
}

// TestSetupOutputAppends checks that --append keeps existing content, which the
// missing O_CREATE in the append branch also affected.
func TestSetupOutputAppends(t *testing.T) {
	origFile, origAppend := outputFile, appendToFile
	t.Cleanup(func() { outputFile, appendToFile = origFile, origAppend })

	path := filepath.Join(t.TempDir(), "out.log")
	require.NoError(t, os.WriteFile(path, []byte("first\n"), 0644), "seed output file")

	outputFile, appendToFile = path, true

	output, err := setupOutput()
	require.NoError(t, err, "setup output")

	_, err = fmt.Fprintln(output, "second")
	require.NoError(t, err, "write to output file")
	require.NoError(t, output.(*os.File).Close(), "close output file")

	content, err := os.ReadFile(path)
	require.NoError(t, err, "read back output file")
	assert.Equal(t, "first\nsecond\n", string(content), "append preserves existing content")
}

// TestSetupOutputMissingDirectory checks the error names the file rather than
// the nil writer.
func TestSetupOutputMissingDirectory(t *testing.T) {
	origFile, origAppend := outputFile, appendToFile
	t.Cleanup(func() { outputFile, appendToFile = origFile, origAppend })

	outputFile, appendToFile = filepath.Join(t.TempDir(), "no-such-dir", "out.log"), false

	_, err := setupOutput()

	require.Error(t, err, "opening a file in a missing directory fails")
	assert.Contains(t, err.Error(), outputFile, "error names the output file")
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
