package ansi

import (
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// Style constants
// ---------------------------------------------------------------------------

func TestStyleConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Reset", Reset, "\x1b[0m"},
		{"Bold", Bold, "\x1b[1m"},
		{"Dim", Dim, "\x1b[2m"},
		{"Italic", Italic, "\x1b[3m"},
		{"Underline", Underline, "\x1b[4m"},
		{"Blink", Blink, "\x1b[5m"},
		{"RapidBlink", RapidBlink, "\x1b[6m"},
		{"Reverse", Reverse, "\x1b[7m"},
		{"Hidden", Hidden, "\x1b[8m"},
		{"Strikethrough", Strikethrough, "\x1b[9m"},
		{"ResetBold", ResetBold, "\x1b[22m"},
		{"ResetItalic", ResetItalic, "\x1b[23m"},
		{"ResetUnderline", ResetUnderline, "\x1b[24m"},
		{"ResetBlink", ResetBlink, "\x1b[25m"},
		{"ResetReverse", ResetReverse, "\x1b[27m"},
		{"ResetHidden", ResetHidden, "\x1b[28m"},
		{"ResetStrikethrough", ResetStrikethrough, "\x1b[29m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Cursor constants
// ---------------------------------------------------------------------------

func TestCursorConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"CursorSave", CursorSave, "\x1b7"},
		{"CursorRestore", CursorRestore, "\x1b8"},
		{"CursorShow", CursorShow, "\x1b[?25h"},
		{"CursorHide", CursorHide, "\x1b[?25l"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Erase constants
// ---------------------------------------------------------------------------

func TestEraseConstants(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"EraseScreenToEnd", EraseScreenToEnd, "\x1b[0J"},
		{"EraseScreenToStart", EraseScreenToStart, "\x1b[1J"},
		{"EraseScreen", EraseScreen, "\x1b[2J"},
		{"EraseScreenAll", EraseScreenAll, "\x1b[3J"},
		{"EraseLineToEnd", EraseLineToEnd, "\x1b[0K"},
		{"EraseLineToStart", EraseLineToStart, "\x1b[1K"},
		{"EraseLine", EraseLine, "\x1b[2K"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Color mode detection
// ---------------------------------------------------------------------------

func TestDetectModeNoColor(t *testing.T) {
	// Reset global state for this test
	old := mode
	defer func() { mode = old }()

	t.Setenv("NO_COLOR", "1")
	got := detectMode()
	if got != ModeNone {
		t.Errorf("NO_COLOR set: got %d, want ModeNone (%d)", got, ModeNone)
	}
}

func TestDetectModeTrueColor(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	os.Unsetenv("NO_COLOR")
	got := detectMode()
	if got != ModeTrueColor {
		t.Errorf("COLORTERM=truecolor: got %d, want ModeTrueColor (%d)", got, ModeTrueColor)
	}
}

func TestDetectMode24Bit(t *testing.T) {
	t.Setenv("COLORTERM", "24bit")
	os.Unsetenv("NO_COLOR")
	got := detectMode()
	if got != ModeTrueColor {
		t.Errorf("COLORTERM=24bit: got %d, want ModeTrueColor (%d)", got, ModeTrueColor)
	}
}

func TestDetectMode256(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("COLORTERM")
	got := detectMode()
	if got != Mode256 {
		t.Errorf("TERM=xterm-256color: got %d, want Mode256 (%d)", got, Mode256)
	}
}

func TestDetectMode16Fallback(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("COLORTERM")
	t.Setenv("TERM", "xterm")
	got := detectMode()
	if got != Mode16 {
		t.Errorf("plain TERM: got %d, want Mode16 (%d)", got, Mode16)
	}
}

func TestSetModeOverride(t *testing.T) {
	old := mode
	defer func() { mode = old }()

	SetMode(Mode256)
	if GetMode() != Mode256 {
		t.Errorf("SetMode(Mode256): GetMode() = %d, want %d", GetMode(), Mode256)
	}
}

// ---------------------------------------------------------------------------
// Color FG/BG in TrueColor mode
// ---------------------------------------------------------------------------

func withMode(m ColorMode, fn func()) {
	old := mode
	SetMode(m)
	defer func() { mode = old }()
	fn()
}

func TestColorFGTrueColor(t *testing.T) {
	withMode(ModeTrueColor, func() {
		tests := []struct {
			name  string
			color Color
			want  string
		}{
			{"Red", Red, "\x1b[38;2;205;0;0m"},
			{"Green", Green, "\x1b[38;2;0;205;0m"},
			{"RGB orange", RGB(255, 165, 0), "\x1b[38;2;255;165;0m"},
			{"Hex", Hex(0xFF8800), "\x1b[38;2;255;136;0m"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.color.FG()
				if got != tt.want {
					t.Errorf("FG() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestColorBGTrueColor(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := RGB(0, 0, 255).BG()
		want := "\x1b[48;2;0;0;255m"
		if got != want {
			t.Errorf("BG() = %q, want %q", got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// Color in Mode16
// ---------------------------------------------------------------------------

func TestColorFG16Named(t *testing.T) {
	withMode(Mode16, func() {
		tests := []struct {
			name  string
			color Color
			want  string
		}{
			{"Black", Black, "\x1b[30m"},
			{"Red", Red, "\x1b[31m"},
			{"Green", Green, "\x1b[32m"},
			{"Yellow", Yellow, "\x1b[33m"},
			{"Blue", Blue, "\x1b[34m"},
			{"Magenta", Magenta, "\x1b[35m"},
			{"Cyan", Cyan, "\x1b[36m"},
			{"White", White, "\x1b[37m"},
			{"BrightBlack", BrightBlack, "\x1b[90m"},
			{"BrightRed", BrightRed, "\x1b[91m"},
			{"BrightGreen", BrightGreen, "\x1b[92m"},
			{"BrightYellow", BrightYellow, "\x1b[93m"},
			{"BrightBlue", BrightBlue, "\x1b[94m"},
			{"BrightMagenta", BrightMagenta, "\x1b[95m"},
			{"BrightCyan", BrightCyan, "\x1b[96m"},
			{"BrightWhite", BrightWhite, "\x1b[97m"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.color.FG()
				if got != tt.want {
					t.Errorf("FG() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestColorBG16Named(t *testing.T) {
	withMode(Mode16, func() {
		tests := []struct {
			name  string
			color Color
			want  string
		}{
			{"Black", Black, "\x1b[40m"},
			{"Red", Red, "\x1b[41m"},
			{"White", White, "\x1b[47m"},
			{"BrightBlack", BrightBlack, "\x1b[100m"},
			{"BrightWhite", BrightWhite, "\x1b[107m"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.color.BG()
				if got != tt.want {
					t.Errorf("BG() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

// ---------------------------------------------------------------------------
// Color downgrading (RGB → 16)
// ---------------------------------------------------------------------------

func TestRGBDowngradeTo16(t *testing.T) {
	withMode(Mode16, func() {
		tests := []struct {
			name string
			r, g, b uint8
			wantIdx int8
		}{
			{"pure red", 255, 0, 0, 9},       // bright red
			{"pure green", 0, 255, 0, 10},     // bright green
			{"pure blue", 0, 0, 255, 4},       // standard blue (0,0,238 is closer than bright blue 92,92,255)
			{"white", 255, 255, 255, 15},       // bright white
			{"black", 0, 0, 0, 0},             // black
			{"near red", 200, 10, 10, 1},      // standard red
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := RGB(tt.r, tt.g, tt.b)
				got := closestANSI16(c.r, c.g, c.b)
				if got != tt.wantIdx {
					t.Errorf("closestANSI16(%d,%d,%d) = %d, want %d", tt.r, tt.g, tt.b, got, tt.wantIdx)
				}
			})
		}
	})
}

// ---------------------------------------------------------------------------
// Color downgrading (RGB → 256)
// ---------------------------------------------------------------------------

func TestRGBDowngradeTo256(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b uint8
		wantIdx uint8
	}{
		{"pure white", 255, 255, 255, 231}, // cube (5,5,5) = exact match
		{"pure black", 0, 0, 0, 16},        // cube (0,0,0) = exact match
		{"mid gray", 128, 128, 128, 244},   // grayscale ramp
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := closestANSI256(tt.r, tt.g, tt.b)
			if got != tt.wantIdx {
				t.Errorf("closestANSI256(%d,%d,%d) = %d, want %d", tt.r, tt.g, tt.b, got, tt.wantIdx)
			}
		})
	}
}

func TestColorFG256(t *testing.T) {
	withMode(Mode256, func() {
		// Named colors in 256 mode should use their index
		got := Red.FG()
		want := "\x1b[38;5;1m"
		if got != want {
			t.Errorf("Red.FG() in Mode256 = %q, want %q", got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// ModeNone
// ---------------------------------------------------------------------------

func TestModeNone(t *testing.T) {
	withMode(ModeNone, func() {
		if got := Red.FG(); got != "" {
			t.Errorf("ModeNone: Red.FG() = %q, want empty", got)
		}
		if got := RGB(255, 0, 0).BG(); got != "" {
			t.Errorf("ModeNone: RGB.BG() = %q, want empty", got)
		}
	})
}

// ---------------------------------------------------------------------------
// Hex constructor
// ---------------------------------------------------------------------------

func TestHex(t *testing.T) {
	c := Hex(0xAABBCC)
	if c.r != 0xAA || c.g != 0xBB || c.b != 0xCC {
		t.Errorf("Hex(0xAABBCC) = {%d,%d,%d}, want {170,187,204}", c.r, c.g, c.b)
	}
	if c.idx16 != -1 {
		t.Errorf("Hex color idx16 = %d, want -1", c.idx16)
	}
}

// ---------------------------------------------------------------------------
// Cursor
// ---------------------------------------------------------------------------

func TestCursorAbsolute(t *testing.T) {
	tests := []struct {
		name string
		pos  Pos
		want string
	}{
		{"origin", Pos{1, 1}, "\x1b[1;1H"},
		{"row5 col10", Pos{10, 5}, "\x1b[5;10H"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Cursor(tt.pos, Abs)
			if got != tt.want {
				t.Errorf("Cursor(%v, Abs) = %q, want %q", tt.pos, got, tt.want)
			}
		})
	}
}

func TestCursorRelative(t *testing.T) {
	tests := []struct {
		name string
		pos  Pos
		want string
	}{
		{"no movement", Pos{0, 0}, ""},
		{"right 3", Pos{3, 0}, "\x1b[3C"},
		{"left 2", Pos{-2, 0}, "\x1b[2D"},
		{"up 1", Pos{0, -1}, "\x1b[1A"},
		{"down 4", Pos{0, 4}, "\x1b[4B"},
		{"up 2 right 3", Pos{3, -2}, "\x1b[2A\x1b[3C"},
		{"down 1 left 5", Pos{-5, 1}, "\x1b[1B\x1b[5D"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Cursor(tt.pos, Rel)
			if got != tt.want {
				t.Errorf("Cursor(%v, Rel) = %q, want %q", tt.pos, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Scroll
// ---------------------------------------------------------------------------

func TestScroll(t *testing.T) {
	if got, want := ScrollUp(3), "\x1b[3S"; got != want {
		t.Errorf("ScrollUp(3) = %q, want %q", got, want)
	}
	if got, want := ScrollDown(1), "\x1b[1T"; got != want {
		t.Errorf("ScrollDown(1) = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Link
// ---------------------------------------------------------------------------

func TestLink(t *testing.T) {
	got := Link("https://example.com", "click")
	want := "\x1b]8;;https://example.com\x1b\\click\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("Link = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// SetTitle
// ---------------------------------------------------------------------------

func TestSetTitle(t *testing.T) {
	got := SetTitle("My App")
	want := "\x1b]0;My App\x1b\\"
	if got != want {
		t.Errorf("SetTitle = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Style
// ---------------------------------------------------------------------------

func TestStyle(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Style("hello", Bold, Red.FG())
		want := "\x1b[1m\x1b[38;2;205;0;0mhello\x1b[0m"
		if got != want {
			t.Errorf("Style = %q, want %q", got, want)
		}
	})
}

func TestStyleNoArgs(t *testing.T) {
	got := Style("plain")
	want := "plain\x1b[0m"
	if got != want {
		t.Errorf("Style with no codes = %q, want %q", got, want)
	}
}
