// Package ansi provides simple ANSI escape code utilities for terminal output.
//
// Colors are represented by the [Color] type, which handles automatic
// downgrading based on detected terminal capabilities. Specify colors by
// name or RGB value — the library emits the best possible escape sequence.
//
//	fmt.Println(ansi.Red.FG("error!"))
//	fmt.Println(ansi.RGB(255, 165, 0).FG("orange"))
//	fmt.Println(ansi.Bold().FG(ansi.Yellow).Text("warning"))
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

// SGR reset codes. Reset performs a full style reset; the individual
// Reset* constants turn off a single attribute.
const (
	Reset              = CSI + "0m"
	ResetBold          = CSI + "22m"
	ResetDim           = CSI + "22m" // same as ResetBold per ANSI spec
	ResetItalic        = CSI + "23m"
	ResetUnderline     = CSI + "24m"
	ResetBlink         = CSI + "25m"
	ResetReverse       = CSI + "27m"
	ResetHidden        = CSI + "28m"
	ResetStrikethrough = CSI + "29m"
)

// Internal SGR codes used by style functions and StyledText methods.
const (
	boldCode          = CSI + "1m"
	dimCode           = CSI + "2m"
	italicCode        = CSI + "3m"
	underlineCode     = CSI + "4m"
	blinkCode         = CSI + "5m"
	rapidBlinkCode    = CSI + "6m"
	reverseCode       = CSI + "7m"
	hiddenCode        = CSI + "8m"
	strikethroughCode = CSI + "9m"
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
type Color struct {
	r, g, b uint8
	idx16   int8 // 0-15 for named colors, -1 for pure RGB
}

// RGB creates a 24-bit color. It will be downgraded automatically if the
// terminal doesn't support true color.
func RGB(r, g, b uint8) Color {
	return Color{r: r, g: g, b: b, idx16: -1}
}

// Hex creates a color from a 24-bit hex value (e.g. 0xFF8800).
func Hex(hex uint32) Color {
	return Color{
		r:     uint8((hex >> 16) & 0xFF),
		g:     uint8((hex >> 8) & 0xFF),
		b:     uint8(hex & 0xFF),
		idx16: -1,
	}
}

func named(idx int8, r, g, b uint8) Color {
	return Color{r: r, g: g, b: b, idx16: idx}
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

// FG returns a [StyledText] with this color set as the foreground. If text
// is provided, it is set on the StyledText for immediate rendering.
//
//	ansi.Red.FG()              // chainable: ansi.Red.FG().Bold().Text("err")
//	ansi.Red.FG("error")       // shortcut:  fmt.Println(ansi.Red.FG("error"))
func (c Color) FG(text ...string) StyledText {
	s := StyledText{codes: []string{c.fgCode()}}
	if len(text) > 0 {
		s.text = text[0]
	}
	return s
}

// BG returns a [StyledText] with this color set as the background. If text
// is provided, it is set on the StyledText for immediate rendering.
func (c Color) BG(text ...string) StyledText {
	s := StyledText{codes: []string{c.bgCode()}}
	if len(text) > 0 {
		s.text = text[0]
	}
	return s
}

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
// StyledText
// ---------------------------------------------------------------------------

// StyledText accumulates ANSI escape codes for fluent text styling.
// It is immutable — every chaining method returns a new StyledText.
//
//	ansi.Bold().FG(ansi.Red).Text("error")
//	ansi.Red.FG().Bold().Underline().Text("critical")
type StyledText struct {
	codes []string
	text  string
}

// with returns a new StyledText with the given code appended.
func (s StyledText) with(code string) StyledText {
	newCodes := make([]string, len(s.codes), len(s.codes)+1)
	copy(newCodes, s.codes)
	newCodes = append(newCodes, code)
	return StyledText{codes: newCodes, text: s.text}
}

// FG adds a foreground color to the style chain.
func (s StyledText) FG(c Color) StyledText { return s.with(c.fgCode()) }

// BG adds a background color to the style chain.
func (s StyledText) BG(c Color) StyledText { return s.with(c.bgCode()) }

// Bold adds bold to the style chain.
func (s StyledText) Bold() StyledText { return s.with(boldCode) }

// Dim adds dim/faint to the style chain.
func (s StyledText) Dim() StyledText { return s.with(dimCode) }

// Italic adds italic to the style chain.
func (s StyledText) Italic() StyledText { return s.with(italicCode) }

// Underline adds underline to the style chain.
func (s StyledText) Underline() StyledText { return s.with(underlineCode) }

// Blink adds blink to the style chain.
func (s StyledText) Blink() StyledText { return s.with(blinkCode) }

// RapidBlink adds rapid blink to the style chain.
func (s StyledText) RapidBlink() StyledText { return s.with(rapidBlinkCode) }

// Reverse adds reverse video to the style chain.
func (s StyledText) Reverse() StyledText { return s.with(reverseCode) }

// Hidden adds hidden to the style chain.
func (s StyledText) Hidden() StyledText { return s.with(hiddenCode) }

// Strikethrough adds strikethrough to the style chain.
func (s StyledText) Strikethrough() StyledText { return s.with(strikethroughCode) }

// Text renders the styled text: accumulated codes + text + [Reset].
// This is the terminal method of the fluent chain.
//
//	ansi.Bold().FG(ansi.Red).Text("error")
func (s StyledText) Text(text string) string {
	return s.render(text)
}

// String implements [fmt.Stringer]. If text has been set (via a constructor
// shortcut like Bold("hello") or Red.FG("text")), it renders the full
// styled string. Otherwise it returns just the accumulated escape codes,
// which is useful for manual string building.
func (s StyledText) String() string {
	if s.text != "" {
		return s.render(s.text)
	}
	// No text — return raw codes only.
	var b strings.Builder
	for _, c := range s.codes {
		b.WriteString(c)
	}
	return b.String()
}

// render builds the final output: codes + text + Reset.
func (s StyledText) render(text string) string {
	// Check if all codes are empty (ModeNone produces empty strings).
	hasCode := false
	for _, c := range s.codes {
		if c != "" {
			hasCode = true
			break
		}
	}
	if !hasCode {
		return text
	}
	var b strings.Builder
	for _, c := range s.codes {
		b.WriteString(c)
	}
	b.WriteString(text)
	b.WriteString(Reset)
	return b.String()
}

// ---------------------------------------------------------------------------
// SGR style functions
// ---------------------------------------------------------------------------

// Bold returns a [StyledText] with bold applied. If text is provided, it
// is set for immediate rendering.
//
//	ansi.Bold("important")                // shortcut
//	ansi.Bold().FG(ansi.Red).Text("err")  // chained
func Bold(text ...string) StyledText {
	return sgrFunc(boldCode, text)
}

// Dim returns a [StyledText] with dim/faint applied.
func Dim(text ...string) StyledText {
	return sgrFunc(dimCode, text)
}

// Italic returns a [StyledText] with italic applied.
func Italic(text ...string) StyledText {
	return sgrFunc(italicCode, text)
}

// Underline returns a [StyledText] with underline applied.
func Underline(text ...string) StyledText {
	return sgrFunc(underlineCode, text)
}

// Blink returns a [StyledText] with blink applied.
func Blink(text ...string) StyledText {
	return sgrFunc(blinkCode, text)
}

// RapidBlink returns a [StyledText] with rapid blink applied.
func RapidBlink(text ...string) StyledText {
	return sgrFunc(rapidBlinkCode, text)
}

// Reverse returns a [StyledText] with reverse video applied.
func Reverse(text ...string) StyledText {
	return sgrFunc(reverseCode, text)
}

// Hidden returns a [StyledText] with hidden applied.
func Hidden(text ...string) StyledText {
	return sgrFunc(hiddenCode, text)
}

// Strikethrough returns a [StyledText] with strikethrough applied.
func Strikethrough(text ...string) StyledText {
	return sgrFunc(strikethroughCode, text)
}

func sgrFunc(code string, text []string) StyledText {
	s := StyledText{codes: []string{code}}
	if len(text) > 0 {
		s.text = text[0]
	}
	return s
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
