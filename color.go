package main

import (
	"fmt"
	gc "image/color"
	"os"
	"strings"

	"github.com/wayneashleyberry/truecolor/pkg/color"
	"golang.org/x/sys/unix"
)

type ColorName string

// These are strings because they will be used to map into a configuration file
// someday.
const (
	ColorNormal      ColorName = "normal"
	ColorDateTime    ColorName = "date/time"
	ColorLevelDebug  ColorName = "level-debug"
	ColorLevelInfo   ColorName = "level-info"
	ColorLevelWarn   ColorName = "level-warn"
	ColorLevelError  ColorName = "level-error"
	ColorLevelDPanic ColorName = "level-dpanic"
	ColorLevelFatal  ColorName = "level-fatal"
	ColorMessage     ColorName = "message"
	ColorStackTrace  ColorName = "stacktrace"
	ColorData        ColorName = "data"
	ColorDataLiteral ColorName = "data-literal"
)

func RGB(r, g, b uint8) gc.Color {
	return &gc.NRGBA{R: r, G: g, B: b, A: 255}
}

type Palette map[ColorName]gc.Color

var DefaultPalette = Palette{
	ColorNormal:      RGB(0xdd, 0xdd, 0xdd),
	ColorDateTime:    RGB(0xdd, 0xdd, 0xdd),
	ColorLevelDebug:  RGB(0x66, 0x66, 0xff),
	ColorLevelInfo:   RGB(0x14, 0xff, 0xff),
	ColorLevelWarn:   RGB(0xff, 0xff, 0x00),
	ColorLevelError:  RGB(0xff, 0xd7, 0x00),
	ColorLevelDPanic: RGB(0xff, 0x5f, 0x00),
	ColorLevelFatal:  RGB(0xff, 0x00, 0x00),
	ColorMessage:     RGB(0xff, 0xff, 0xff),
	ColorStackTrace:  RGB(0x76, 0x76, 0x76),
	ColorData:        RGB(0xaa, 0xaa, 0xaa),
	ColorDataLiteral: RGB(0x88, 0x88, 0x99),
}

type PlainColorizer interface {
	C(c ColorName, value ...any) string
}

type SugaredColorizer struct {
	PlainColorizer
}

func NewSugaredColorizer(pc PlainColorizer) *SugaredColorizer {
	return &SugaredColorizer{pc}
}

func (fc *SugaredColorizer) Cf(color ColorName, f string, args ...any) string {
	return fc.C(color, fmt.Sprintf(f, args...))
}

type RawColorizer struct{}

func (rc *RawColorizer) C(c gc.Color, v ...any) string {
	r, g, b, a := c.RGBA()
	return color.RGBA(gc.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}).Sprint(v...)
}

type ColorAuto struct {
	on  *ColorOn
	off *ColorOff
	tty bool
}

func NewColorAuto() *ColorAuto {
	ca := ColorAuto{
		on:  NewColorOn(DefaultPalette),
		off: &ColorOff{},
	}

	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	ca.tty = err != nil

	return &ca
}

func (ca *ColorAuto) C(c ColorName, v ...any) string {
	if ca.tty {
		return ca.on.C(c, v...)
	}
	return ca.off.C(c, v...)
}

type ColorOff struct{}

func (co *ColorOff) C(_ ColorName, v ...any) string {
	return fmt.Sprint(v...)
}

type ColorOn struct {
	raw     RawColorizer
	palette Palette
}

func NewColorOn(p Palette) *ColorOn {
	return &ColorOn{palette: p}
}

func (co *ColorOn) C(c ColorName, v ...any) string {
	gcc, ok := co.palette[c]
	if !ok {
		gcc, ok = co.palette[ColorNormal]
		if !ok {
			gcc = RGB(0xdd, 0xdd, 0xdd)
		}
	}
	return co.raw.C(gcc, v...)
}

var l2cn = map[string]ColorName{
	"debug":  ColorLevelDebug,
	"info":   ColorLevelInfo,
	"warn":   ColorLevelWarn,
	"error":  ColorLevelError,
	"dpanic": ColorLevelDPanic,
	"fatal":  ColorLevelFatal,
}

func LevelToColorName(level string) ColorName {
	cn, ok := l2cn[strings.ToLower(level)]
	if ok {
		return cn
	}
	return "level-info"
}
