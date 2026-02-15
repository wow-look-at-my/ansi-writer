// Package ansi provides simple ANSI escape code utilities for terminal output.
//
// Colors are represented by the [Color] type, which handles automatic
// downgrading based on detected terminal capabilities. Styles and colors
// implement [fmt.Stringer] for direct use with fmt:
//
//	fmt.Print(ansi.Bold, ansi.Red.FG, "error!", ansi.Reset)
//	fmt.Print(ansi.Italic, "note", ansi.Reset)
//	fmt.Print(ansi.Cursor(ansi.Pos{10, 5}, ansi.Abs))
package ansi

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Base escape sequences.
const (
	ESC = "\x1b"     // Escape character
	CSI = ESC + "["  // Control Sequence Introducer
	OSC = ESC + "]"  // Operating System Command
	ST  = ESC + "\\" // String Terminator
)

// ---------------------------------------------------------------------------
// Style type
// ---------------------------------------------------------------------------

// Style represents an ANSI SGR escape sequence. It implements [fmt.Stringer]
// so it can be used directly with fmt functions. In [ModeNone], String
// returns an empty string.
//
// Because Style's underlying type is string, fmt.Sprint and fmt.Print
// will not insert spaces between adjacent Style values.
type Style string

// String returns the escape sequence, or an empty string in [ModeNone].
// Color styles are resolved lazily against the current [ColorMode].
func (s Style) String() string {
	if GetMode() == ModeNone {
		return ""
	}
	if len(s) == 6 && s[0] == 0 {
		c := Color{r: s[2], g: s[3], b: s[4], idx16: int8(s[5])}
		if s[1] == 'f' {
			return c.fgCode()
		}
		return c.bgCode()
	}
	return string(s)
}

func encodeColor(kind byte, c Color) Style {
	return Style([]byte{0, kind, c.r, c.g, c.b, byte(c.idx16)})
}

// ---------------------------------------------------------------------------
// SGR style and reset codes
// ---------------------------------------------------------------------------

// SGR style codes. Use directly with fmt:
//
//	fmt.Print(ansi.Bold, "important", ansi.Reset)
//	fmt.Print(ansi.Italic, ansi.Underline, "fancy", ansi.Reset)
var (
	Bold          = Style(CSI + "1m")
	Dim           = Style(CSI + "2m")
	Italic        = Style(CSI + "3m")
	Underline     = Style(CSI + "4m")
	Blink         = Style(CSI + "5m")
	RapidBlink    = Style(CSI + "6m")
	Reverse       = Style(CSI + "7m")
	Hidden        = Style(CSI + "8m")
	Strikethrough = Style(CSI + "9m")
)

// SGR reset codes. [Reset] performs a full style reset; the individual
// Reset* values turn off a single attribute.
var (
	Reset              = Style(CSI + "0m")
	ResetBold          = Style(CSI + "22m")
	ResetDim           = Style(CSI + "22m") // same as ResetBold per ANSI spec
	ResetItalic        = Style(CSI + "23m")
	ResetUnderline     = Style(CSI + "24m")
	ResetBlink         = Style(CSI + "25m")
	ResetReverse       = Style(CSI + "27m")
	ResetHidden        = Style(CSI + "28m")
	ResetStrikethrough = Style(CSI + "29m")
)

// Cursor constants (non-positional).
const (
	CursorSave    = ESC + "7"    // DECSC — save cursor position
	CursorRestore = ESC + "8"    // DECRC — restore cursor position
	CursorShow    = CSI + "?25h" // show cursor
	CursorHide    = CSI + "?25l" // hide cursor
)

// Erase constants.
const (
	EraseScreenToEnd   = CSI + "0J" // cursor to end of screen
	EraseScreenToStart = CSI + "1J" // start of screen to cursor
	EraseScreen        = CSI + "2J" // entire screen
	EraseScreenAll     = CSI + "3J" // entire screen + scrollback

	EraseLineToEnd   = CSI + "0K" // cursor to end of line
	EraseLineToStart = CSI + "1K" // start of line to cursor
	EraseLine        = CSI + "2K" // entire line
)

// Abs and Rel are convenience constants for the Cursor function.
const (
	Abs = true
	Rel = false
)

// ---------------------------------------------------------------------------
// Color mode
// ---------------------------------------------------------------------------

// ColorMode represents the terminal's color capability.
type ColorMode int

const (
	ModeAuto      ColorMode = iota // detect from environment (default)
	ModeTrueColor                  // 24-bit RGB
	Mode256                        // 256-color palette
	Mode16                         // standard 16 colors
	ModeNone                       // no color output
)

var (
	mode     ColorMode // ModeAuto by default (zero value)
	modeMu   sync.Mutex
	detected ColorMode
	once     sync.Once
)

// SetMode overrides auto-detection with a specific color mode.
func SetMode(m ColorMode) {
	modeMu.Lock()
	mode = m
	modeMu.Unlock()
}

// GetMode returns the effective color mode. If ModeAuto is set (default),
// it detects the mode from environment variables on the first call.
func GetMode() ColorMode {
	modeMu.Lock()
	m := mode
	modeMu.Unlock()
	if m != ModeAuto {
		return m
	}
	once.Do(func() {
		detected = detectMode()
	})
	return detected
}

func detectMode() ColorMode {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return ModeNone
	}
	ct := os.Getenv("COLORTERM")
	if ct == "truecolor" || ct == "24bit" {
		return ModeTrueColor
	}
	if strings.Contains(os.Getenv("TERM"), "256color") {
		return Mode256
	}
	return Mode16
}

// ---------------------------------------------------------------------------
// Color type
// ---------------------------------------------------------------------------

// Color represents a color that renders at the best available color depth.
// The FG and BG fields are [Style] values for use with fmt:
//
//	fmt.Print(ansi.Red.FG, "error", ansi.Reset)
//	fmt.Print(ansi.Blue.BG, "highlight", ansi.Reset)
type Color struct {
	r, g, b uint8
	idx16   int8 // 0-15 for named colors, -1 for pure RGB
	FG      Style
	BG      Style
}

func newColor(r, g, b uint8, idx16 int8) Color {
	c := Color{r: r, g: g, b: b, idx16: idx16}
	c.FG = encodeColor('f', c)
	c.BG = encodeColor('b', c)
	return c
}

// RGB creates a 24-bit color. It will be downgraded automatically if the
// terminal doesn't support true color.
func RGB(r, g, b uint8) Color {
	return newColor(r, g, b, -1)
}

// Hex creates a color from a 24-bit hex value (e.g. 0xFF8800).
func Hex(hex uint32) Color {
	return newColor(
		uint8((hex>>16)&0xFF),
		uint8((hex>>8)&0xFF),
		uint8(hex&0xFF),
		-1,
	)
}

func named(idx int8, r, g, b uint8) Color {
	return newColor(r, g, b, idx)
}

// Standard named colors.
var (
	Black   = named(0, 0, 0, 0)
	Red     = named(1, 205, 0, 0)
	Green   = named(2, 0, 205, 0)
	Yellow  = named(3, 205, 205, 0)
	Blue    = named(4, 0, 0, 238)
	Magenta = named(5, 205, 0, 205)
	Cyan    = named(6, 0, 205, 205)
	White   = named(7, 229, 229, 229)

	BrightBlack   = named(8, 127, 127, 127)
	BrightRed     = named(9, 255, 0, 0)
	BrightGreen   = named(10, 0, 255, 0)
	BrightYellow  = named(11, 255, 255, 0)
	BrightBlue    = named(12, 92, 92, 255)
	BrightMagenta = named(13, 255, 0, 255)
	BrightCyan    = named(14, 0, 255, 255)
	BrightWhite   = named(15, 255, 255, 255)
)

// fgCode returns the raw foreground escape sequence string.
func (c Color) fgCode() string {
	switch GetMode() {
	case ModeNone:
		return ""
	case Mode16:
		return c.fg16()
	case Mode256:
		return c.fg256()
	default: // ModeTrueColor
		return CSI + "38;2;" + utoa(c.r) + ";" + utoa(c.g) + ";" + utoa(c.b) + "m"
	}
}

// bgCode returns the raw background escape sequence string.
func (c Color) bgCode() string {
	switch GetMode() {
	case ModeNone:
		return ""
	case Mode16:
		return c.bg16()
	case Mode256:
		return c.bg256()
	default: // ModeTrueColor
		return CSI + "48;2;" + utoa(c.r) + ";" + utoa(c.g) + ";" + utoa(c.b) + "m"
	}
}

func (c Color) fg16() string {
	idx := c.nearest16()
	if idx < 8 {
		return CSI + strconv.Itoa(30+int(idx)) + "m"
	}
	return CSI + strconv.Itoa(90+int(idx)-8) + "m"
}

func (c Color) bg16() string {
	idx := c.nearest16()
	if idx < 8 {
		return CSI + strconv.Itoa(40+int(idx)) + "m"
	}
	return CSI + strconv.Itoa(100+int(idx)-8) + "m"
}

func (c Color) fg256() string {
	return CSI + "38;5;" + strconv.Itoa(int(c.nearest256())) + "m"
}

func (c Color) bg256() string {
	return CSI + "48;5;" + strconv.Itoa(int(c.nearest256())) + "m"
}

func (c Color) nearest16() int8 {
	if c.idx16 >= 0 {
		return c.idx16
	}
	return closestANSI16(c.r, c.g, c.b)
}

func (c Color) nearest256() uint8 {
	if c.idx16 >= 0 {
		return uint8(c.idx16) // first 16 of the 256 palette
	}
	return closestANSI256(c.r, c.g, c.b)
}

// ---------------------------------------------------------------------------
// Color matching
// ---------------------------------------------------------------------------

// Canonical RGB values for the 16 standard ANSI colors.
var ansi16RGB = [16][3]uint8{
	{0, 0, 0},       // black
	{205, 0, 0},     // red
	{0, 205, 0},     // green
	{205, 205, 0},   // yellow
	{0, 0, 238},     // blue
	{205, 0, 205},   // magenta
	{0, 205, 205},   // cyan
	{229, 229, 229}, // white
	{127, 127, 127}, // bright black
	{255, 0, 0},     // bright red
	{0, 255, 0},     // bright green
	{255, 255, 0},   // bright yellow
	{92, 92, 255},   // bright blue
	{255, 0, 255},   // bright magenta
	{0, 255, 255},   // bright cyan
	{255, 255, 255}, // bright white
}

func colorDist(r1, g1, b1, r2, g2, b2 uint8) int {
	dr := int(r1) - int(r2)
	dg := int(g1) - int(g2)
	db := int(b1) - int(b2)
	return dr*dr + dg*dg + db*db
}

func closestANSI16(r, g, b uint8) int8 {
	best := 0
	bestDist := colorDist(r, g, b, ansi16RGB[0][0], ansi16RGB[0][1], ansi16RGB[0][2])
	for i := 1; i < 16; i++ {
		d := colorDist(r, g, b, ansi16RGB[i][0], ansi16RGB[i][1], ansi16RGB[i][2])
		if d < bestDist {
			bestDist = d
			best = i
		}
	}
	return int8(best)
}

// The 6-level values used in the 216-color cube (indices 16-231).
var cubeValues = [6]uint8{0, 95, 135, 175, 215, 255}

func closestCubeIdx(v uint8) int {
	best := 0
	bestDist := abs8(v, cubeValues[0])
	for i := 1; i < 6; i++ {
		d := abs8(v, cubeValues[i])
		if d < bestDist {
			bestDist = d
			best = i
		}
	}
	return best
}

func abs8(a, b uint8) int {
	if a > b {
		return int(a) - int(b)
	}
	return int(b) - int(a)
}

func closestANSI256(r, g, b uint8) uint8 {
	// Check the 6x6x6 cube
	ri := closestCubeIdx(r)
	gi := closestCubeIdx(g)
	bi := closestCubeIdx(b)
	cubeIdx := uint8(16 + 36*ri + 6*gi + bi)
	cubeDist := colorDist(r, g, b, cubeValues[ri], cubeValues[gi], cubeValues[bi])

	// Check the 24-step grayscale ramp (indices 232-255): 8, 18, 28, ..., 238
	gray := (int(r) + int(g) + int(b)) / 3
	gi2 := (gray - 8 + 5) / 10 // nearest step
	if gi2 < 0 {
		gi2 = 0
	} else if gi2 > 23 {
		gi2 = 23
	}
	grayVal := uint8(8 + gi2*10)
	grayIdx := uint8(232 + gi2)
	grayDist := colorDist(r, g, b, grayVal, grayVal, grayVal)

	// Check the base 16 too
	base16 := closestANSI16(r, g, b)
	base16Dist := colorDist(r, g, b, ansi16RGB[base16][0], ansi16RGB[base16][1], ansi16RGB[base16][2])

	// Pick the best
	bestIdx := cubeIdx
	bestDist := cubeDist
	if grayDist < bestDist {
		bestIdx = grayIdx
		bestDist = grayDist
	}
	if base16Dist < bestDist {
		bestIdx = uint8(base16)
	}
	return bestIdx
}

// ---------------------------------------------------------------------------
// Pos & Cursor
// ---------------------------------------------------------------------------

// Pos represents a 2D terminal coordinate. X is the column, Y is the row.
type Pos struct{ X, Y int }

// Cursor emits escape sequences to move the cursor. If absolute is true
// (use [Abs]), it moves to the position (1-based row/col). If false (use
// [Rel]), it moves relative to the current position.
func Cursor(p Pos, absolute bool) string {
	if absolute {
		return CSI + strconv.Itoa(p.Y) + ";" + strconv.Itoa(p.X) + "H"
	}
	var b strings.Builder
	if p.Y < 0 {
		b.WriteString(CSI)
		b.WriteString(strconv.Itoa(-p.Y))
		b.WriteByte('A') // up
	} else if p.Y > 0 {
		b.WriteString(CSI)
		b.WriteString(strconv.Itoa(p.Y))
		b.WriteByte('B') // down
	}
	if p.X > 0 {
		b.WriteString(CSI)
		b.WriteString(strconv.Itoa(p.X))
		b.WriteByte('C') // forward
	} else if p.X < 0 {
		b.WriteString(CSI)
		b.WriteString(strconv.Itoa(-p.X))
		b.WriteByte('D') // back
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Scroll
// ---------------------------------------------------------------------------

// ScrollUp scrolls the terminal content up by n lines.
func ScrollUp(n int) string {
	return CSI + strconv.Itoa(n) + "S"
}

// ScrollDown scrolls the terminal content down by n lines.
func ScrollDown(n int) string {
	return CSI + strconv.Itoa(n) + "T"
}

// ---------------------------------------------------------------------------
// OSC functions
// ---------------------------------------------------------------------------

// Link wraps text in an OSC 8 hyperlink.
func Link(url, text string) string {
	return OSC + "8;;" + url + ST + text + OSC + "8;;" + ST
}

// SetTitle sets the terminal window title.
func SetTitle(title string) string {
	return OSC + "0;" + title + ST
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func utoa(n uint8) string {
	return strconv.FormatUint(uint64(n), 10)
}
