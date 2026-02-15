package ansi

import (
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// SGR codes (via style functions)
// ---------------------------------------------------------------------------

func TestSGRCodes(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Reset", Reset, "\x1b[0m"},
		{"Bold", Bold().String(), "\x1b[1m"},
		{"Dim", Dim().String(), "\x1b[2m"},
		{"Italic", Italic().String(), "\x1b[3m"},
		{"Underline", Underline().String(), "\x1b[4m"},
		{"Blink", Blink().String(), "\x1b[5m"},
		{"RapidBlink", RapidBlink().String(), "\x1b[6m"},
		{"Reverse", Reverse().String(), "\x1b[7m"},
		{"Hidden", Hidden().String(), "\x1b[8m"},
		{"Strikethrough", Strikethrough().String(), "\x1b[9m"},
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
// SGR style functions with text shortcut
// ---------------------------------------------------------------------------

func TestSGRFunctionWithText(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Bold", Bold("hello").String(), "\x1b[1mhello\x1b[0m"},
		{"Dim", Dim("hello").String(), "\x1b[2mhello\x1b[0m"},
		{"Italic", Italic("hello").String(), "\x1b[3mhello\x1b[0m"},
		{"Underline", Underline("hello").String(), "\x1b[4mhello\x1b[0m"},
		{"Strikethrough", Strikethrough("hello").String(), "\x1b[9mhello\x1b[0m"},
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
// Color FG/BG raw codes in TrueColor mode
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
				got := tt.color.fgCode()
				if got != tt.want {
					t.Errorf("fgCode() = %q, want %q", got, tt.want)
				}
			})
		}
	})
}

func TestColorBGTrueColor(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := RGB(0, 0, 255).bgCode()
		want := "\x1b[48;2;0;0;255m"
		if got != want {
			t.Errorf("bgCode() = %q, want %q", got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// Color FG/BG with text in TrueColor mode
// ---------------------------------------------------------------------------

func TestColorFGWithText(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Red.FG("error").String()
		want := "\x1b[38;2;205;0;0merror\x1b[0m"
		if got != want {
			t.Errorf("Red.FG(\"error\").String() = %q, want %q", got, want)
		}
	})
}

func TestColorBGWithText(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Blue.BG("warning").String()
		want := "\x1b[48;2;0;0;238mwarning\x1b[0m"
		if got != want {
			t.Errorf("Blue.BG(\"warning\").String() = %q, want %q", got, want)
		}
	})
}

func TestColorFGChained(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Red.FG().Bold().Text("err")
		want := "\x1b[38;2;205;0;0m\x1b[1merr\x1b[0m"
		if got != want {
			t.Errorf("Red.FG().Bold().Text() = %q, want %q", got, want)
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
				got := tt.color.fgCode()
				if got != tt.want {
					t.Errorf("fgCode() = %q, want %q", got, tt.want)
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
				got := tt.color.bgCode()
				if got != tt.want {
					t.Errorf("bgCode() = %q, want %q", got, tt.want)
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
			name    string
			r, g, b uint8
			wantIdx int8
		}{
			{"pure red", 255, 0, 0, 9},
			{"pure green", 0, 255, 0, 10},
			{"pure blue", 0, 0, 255, 4},
			{"white", 255, 255, 255, 15},
			{"black", 0, 0, 0, 0},
			{"near red", 200, 10, 10, 1},
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
		{"pure white", 255, 255, 255, 231},
		{"pure black", 0, 0, 0, 16},
		{"mid gray", 128, 128, 128, 244},
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
		got := Red.fgCode()
		want := "\x1b[38;5;1m"
		if got != want {
			t.Errorf("Red.fgCode() in Mode256 = %q, want %q", got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// ModeNone
// ---------------------------------------------------------------------------

func TestModeNone(t *testing.T) {
	withMode(ModeNone, func() {
		if got := Red.fgCode(); got != "" {
			t.Errorf("ModeNone: Red.fgCode() = %q, want empty", got)
		}
		if got := RGB(255, 0, 0).bgCode(); got != "" {
			t.Errorf("ModeNone: RGB.bgCode() = %q, want empty", got)
		}
	})
}

func TestModeNoneFGWithText(t *testing.T) {
	withMode(ModeNone, func() {
		got := Red.FG("hello").String()
		if got != "hello" {
			t.Errorf("ModeNone: Red.FG(\"hello\").String() = %q, want %q", got, "hello")
		}
	})
}

func TestModeNoneStyledText(t *testing.T) {
	withMode(ModeNone, func() {
		got := Bold().FG(Red).Text("hello")
		// In ModeNone, FG code is empty but Bold code is still present.
		// Bold code is a raw SGR string, not mode-dependent.
		want := "\x1b[1mhello\x1b[0m"
		if got != want {
			t.Errorf("ModeNone: Bold().FG(Red).Text() = %q, want %q", got, want)
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
// StyledText chaining
// ---------------------------------------------------------------------------

func TestStyledTextChaining(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Bold().FG(Red).Text("hello")
		want := "\x1b[1m\x1b[38;2;205;0;0mhello\x1b[0m"
		if got != want {
			t.Errorf("Bold().FG(Red).Text() = %q, want %q", got, want)
		}
	})
}

func TestStyledTextFGAndBG(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Red.FG().BG(Blue).Bold().Text("fancy")
		want := "\x1b[38;2;205;0;0m\x1b[48;2;0;0;238m\x1b[1mfancy\x1b[0m"
		if got != want {
			t.Errorf("FG+BG+Bold = %q, want %q", got, want)
		}
	})
}

func TestStyledTextStringNoText(t *testing.T) {
	got := Bold().String()
	want := "\x1b[1m"
	if got != want {
		t.Errorf("Bold().String() = %q, want %q", got, want)
	}
}

func TestStyledTextStringWithText(t *testing.T) {
	got := Bold("hello").String()
	want := "\x1b[1mhello\x1b[0m"
	if got != want {
		t.Errorf("Bold(\"hello\").String() = %q, want %q", got, want)
	}
}

func TestStyledTextReuse(t *testing.T) {
	withMode(ModeTrueColor, func() {
		base := Red.FG().Bold()
		got1 := base.Text("first")
		got2 := base.Text("second")
		want1 := "\x1b[38;2;205;0;0m\x1b[1mfirst\x1b[0m"
		want2 := "\x1b[38;2;205;0;0m\x1b[1msecond\x1b[0m"
		if got1 != want1 {
			t.Errorf("reuse first = %q, want %q", got1, want1)
		}
		if got2 != want2 {
			t.Errorf("reuse second = %q, want %q", got2, want2)
		}
	})
}

func TestStyledTextImmutability(t *testing.T) {
	withMode(ModeTrueColor, func() {
		base := Bold()
		withRed := base.FG(Red)
		withBlue := base.FG(Blue)

		gotRed := withRed.Text("r")
		gotBlue := withBlue.Text("b")

		wantRed := "\x1b[1m\x1b[38;2;205;0;0mr\x1b[0m"
		wantBlue := "\x1b[1m\x1b[48;2;0;0;238mb\x1b[0m"

		if gotRed != wantRed {
			t.Errorf("immutability red = %q, want %q", gotRed, wantRed)
		}
		// Blue.fgCode() is "\x1b[38;2;0;0;238m", not bgCode
		wantBlue = "\x1b[1m\x1b[38;2;0;0;238mb\x1b[0m"
		if gotBlue != wantBlue {
			t.Errorf("immutability blue = %q, want %q", gotBlue, wantBlue)
		}
	})
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
