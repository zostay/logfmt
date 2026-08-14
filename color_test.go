package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIsTerminalNotATerminal covers the destinations logfmt output actually
// gets redirected to. Each of these must report false, or logfmt writes ANSI
// escapes into a pipe or a file.
func TestIsTerminalNotATerminal(t *testing.T) {
	pr, pw, err := os.Pipe()
	require.NoError(t, err, "open pipe")
	defer func() { _ = pr.Close() }()
	defer func() { _ = pw.Close() }()

	regular, err := os.Create(filepath.Join(t.TempDir(), "out.log"))
	require.NoError(t, err, "create regular file")
	defer func() { _ = regular.Close() }()

	// /dev/null is a character device, so a mode-based check misidentifies it
	// as a terminal.
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	require.NoError(t, err, "open %s", os.DevNull)
	defer func() { _ = devNull.Close() }()

	notTerminals := map[string]io.Writer{
		"pipe":         pw,
		"regular file": regular,
		os.DevNull:     devNull,
		"buffer":       &bytes.Buffer{},
	}

	for name, w := range notTerminals {
		assert.False(t, isTerminal(w), "%s is not a terminal", name)
	}
}

// TestIsTerminalTerminal checks the positive case against the controlling
// terminal. There is no controlling terminal under CI or `go test` without a
// tty, so the test skips rather than fails there.
func TestIsTerminalTerminal(t *testing.T) {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		t.Skipf("no controlling terminal available: %v", err)
	}
	defer func() { _ = tty.Close() }()

	assert.True(t, isTerminal(tty), "/dev/tty is a terminal")
}

// TestColorAutoOffWhenNotATerminal is the regression guard for issue #1: with
// the detection inverted, auto mode colorized piped and redirected output.
func TestColorAutoOffWhenNotATerminal(t *testing.T) {
	regular, err := os.Create(filepath.Join(t.TempDir(), "out.log"))
	require.NoError(t, err, "create regular file")
	defer func() { _ = regular.Close() }()

	ca := NewColorAuto(DefaultPalette, regular)
	require.False(t, ca.tty, "auto mode detects non-terminal")

	got := ca.C(ColorLevelError, "boom")
	assert.Equal(t, "boom", got, "no escape sequences when output is not a terminal")
}

// TestColorAutoUsesSuppliedPalette guards against auto mode falling back to the
// default palette and ignoring configured colors.
func TestColorAutoUsesSuppliedPalette(t *testing.T) {
	custom := Palette{ColorLevelError: RGB(0x01, 0x02, 0x03)}

	ca := NewColorAuto(custom, &bytes.Buffer{})

	assert.Equal(t, custom, ca.on.palette, "auto mode keeps the supplied palette")
}
