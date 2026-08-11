package ansi

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// ---------------------------------------------------------------------------
// SGR codes (style and reset constants)
// ---------------------------------------------------------------------------

func TestSGRCodes(t *testing.T) {
	withMode(ModeTrueColor, func() {
		tests := []struct {
			name string
			got  string
			want string
		}{
			{"Bold", Bold.String(), "\x1b[1m"},
			{"Dim", Dim.String(), "\x1b[2m"},
			{"Italic", Italic.String(), "\x1b[3m"},
			{"Underline", Underline.String(), "\x1b[4m"},
			{"Blink", Blink.String(), "\x1b[5m"},
			{"RapidBlink", RapidBlink.String(), "\x1b[6m"},
			{"Reverse", Reverse.String(), "\x1b[7m"},
			{"Hidden", Hidden.String(), "\x1b[8m"},
			{"Strikethrough", Strikethrough.String(), "\x1b[9m"},
			{"Reset", Reset.String(), "\x1b[0m"},
			{"ResetBold", ResetBold.String(), "\x1b[22m"},
			{"ResetItalic", ResetItalic.String(), "\x1b[23m"},
			{"ResetUnderline", ResetUnderline.String(), "\x1b[24m"},
			{"ResetBlink", ResetBlink.String(), "\x1b[25m"},
			{"ResetReverse", ResetReverse.String(), "\x1b[27m"},
			{"ResetHidden", ResetHidden.String(), "\x1b[28m"},
			{"ResetStrikethrough", ResetStrikethrough.String(), "\x1b[29m"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				assert.Equal(t, tt.want, tt.got)

			})
		}
	})
}

// ---------------------------------------------------------------------------
// Style works with fmt
// ---------------------------------------------------------------------------

func TestStyleFmt(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := fmt.Sprint(Bold, "hello", Reset)
		want := "\x1b[1mhello\x1b[0m"
		assert.Equal(t, want, got)

	})
}

func TestStyleFmtWithColor(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := fmt.Sprintf("%s%s%s%s", Bold, Red.FG, "error", Reset)
		want := "\x1b[1m\x1b[38;2;205;0;0merror\x1b[0m"
		assert.Equal(t, want, got)

	})
}

// ---------------------------------------------------------------------------
// Style ModeNone
// ---------------------------------------------------------------------------

func TestStyleModeNone(t *testing.T) {
	withMode(ModeNone, func() {
		tests := []struct {
			name string
			s    Style
		}{
			{"Bold", Bold},
			{"Italic", Italic},
			{"Reset", Reset},
			{"ResetBold", ResetBold},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := tt.s.String()
				assert.Equal(t, "", got)

			})
		}
	})
}

func TestStyleModeNoneFmt(t *testing.T) {
	withMode(ModeNone, func() {
		got := fmt.Sprintf("%s%s%s%s", Bold, Red.FG, "hello", Reset)
		want := "hello"
		assert.Equal(t, want, got)

	})
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
			assert.Equal(t, tt.want, tt.got)

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
			assert.Equal(t, tt.want, tt.got)

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
	assert.Equal(t, ModeNone, got)

}

func TestDetectModeTrueColor(t *testing.T) {
	t.Setenv("COLORTERM", "truecolor")
	os.Unsetenv("NO_COLOR")
	got := detectMode()
	assert.Equal(t, ModeTrueColor, got)

}

func TestDetectMode24Bit(t *testing.T) {
	t.Setenv("COLORTERM", "24bit")
	os.Unsetenv("NO_COLOR")
	got := detectMode()
	assert.Equal(t, ModeTrueColor, got)

}

func TestDetectMode256(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("COLORTERM")
	got := detectMode()
	assert.Equal(t, Mode256, got)

}

func TestDetectMode16Fallback(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("COLORTERM")
	t.Setenv("TERM", "xterm")
	got := detectMode()
	assert.Equal(t, Mode16, got)

}

func TestSetModeOverride(t *testing.T) {
	old := mode
	defer func() { mode = old }()

	SetMode(Mode256)
	assert.Equal(t, Mode256, GetMode())

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
				assert.Equal(t, tt.want, got)

			})
		}
	})
}

func TestColorBGTrueColor(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := RGB(0, 0, 255).bgCode()
		want := "\x1b[48;2;0;0;255m"
		assert.Equal(t, want, got)

	})
}

// ---------------------------------------------------------------------------
// Color FG/BG as Style in TrueColor mode
// ---------------------------------------------------------------------------

func TestColorFGStyle(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Red.FG.String()
		want := "\x1b[38;2;205;0;0m"
		assert.Equal(t, want, got)

	})
}

func TestColorBGStyle(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := Blue.BG.String()
		want := "\x1b[48;2;0;0;238m"
		assert.Equal(t, want, got)

	})
}

func TestColorFGFmt(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := fmt.Sprint(Red.FG, "error", Reset)
		want := "\x1b[38;2;205;0;0merror\x1b[0m"
		assert.Equal(t, want, got)

	})
}

func TestColorBGFmt(t *testing.T) {
	withMode(ModeTrueColor, func() {
		got := fmt.Sprint(Blue.BG, "warning", Reset)
		want := "\x1b[48;2;0;0;238mwarning\x1b[0m"
		assert.Equal(t, want, got)

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
				assert.Equal(t, tt.want, got)

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
				assert.Equal(t, tt.want, got)

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
				assert.Equal(t, tt.wantIdx, got)

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
			assert.Equal(t, tt.wantIdx, got)

		})
	}
}

func TestColorFG256(t *testing.T) {
	withMode(Mode256, func() {
		got := Red.fgCode()
		want := "\x1b[38;5;1m"
		assert.Equal(t, want, got)

	})
}

// ---------------------------------------------------------------------------
// ModeNone
// ---------------------------------------------------------------------------

func TestModeNone(t *testing.T) {
	withMode(ModeNone, func() {
		got := Red.fgCode()
		assert.Equal(t, "", got)

		got = RGB(255, 0, 0).bgCode()
		assert.Equal(t, "", got)

	})
}

func TestColorFGLazyResolution(t *testing.T) {
	withMode(ModeTrueColor, func() {
		want16 := "\x1b[31m"
		wantTC := "\x1b[38;2;205;0;0m"
		fg := Red.FG
		got := fg.String()
		assert.Equal(t, wantTC, got)

		SetMode(Mode16)
		got = fg.String()
		assert.Equal(t, want16, got)

	})
}

func TestModeNoneColorFG(t *testing.T) {
	withMode(ModeNone, func() {
		got := Red.FG.String()
		assert.Equal(t, "", got)

	})
}

// ---------------------------------------------------------------------------
// Hex constructor
// ---------------------------------------------------------------------------

func TestHex(t *testing.T) {
	c := Hex(0xAABBCC)
	assert.False(t, c.r != 0xAA || c.g != 0xBB || c.b != 0xCC)
	assert.Equal(t, -1, c.idx16)

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
			assert.Equal(t, tt.want, got)

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
			assert.Equal(t, tt.want, got)

		})
	}
}

// ---------------------------------------------------------------------------
// Scroll
// ---------------------------------------------------------------------------

func TestScroll(t *testing.T) {
	got, want := ScrollUp(3), "\x1b[3S"
	assert.Equal(t, want, got)

	got, want = ScrollDown(1), "\x1b[1T"
	assert.Equal(t, want, got)

}

// ---------------------------------------------------------------------------
// Link
// ---------------------------------------------------------------------------

func TestLink(t *testing.T) {
	got := Link("https://example.com", "click")
	want := "\x1b]8;;https://example.com\x1b\\click\x1b]8;;\x1b\\"
	assert.Equal(t, want, got)

}

// ---------------------------------------------------------------------------
// SetTitle
// ---------------------------------------------------------------------------

func TestSetTitle(t *testing.T) {
	got := SetTitle("My App")
	want := "\x1b]0;My App\x1b\\"
	assert.Equal(t, want, got)

}

// ---------------------------------------------------------------------------
// Concat
// ---------------------------------------------------------------------------

func TestConcat(t *testing.T) {
	withMode(ModeTrueColor, func() {
		tests := []struct {
			name  string
			parts []any
			want  string
		}{
			{"strings only", []any{"hello", " ", "world"}, "hello world"},
			{"style and string", []any{Bold, "text", Reset}, "\x1b[1mtext\x1b[0m"},
			{"color fg", []any{Red.FG, "error", Reset}, "\x1b[38;2;205;0;0merror\x1b[0m"},
			{"mixed styles", []any{Bold, Red.FG, "error", Reset}, "\x1b[1m\x1b[38;2;205;0;0merror\x1b[0m"},
			{"empty", []any{}, ""},
			{"single string", []any{"alone"}, "alone"},
			{"single style", []any{Bold}, "\x1b[1m"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := Concat(tt.parts...)
				assert.Equal(t, tt.want, got)
			})
		}
	})
}

func TestConcatModeNone(t *testing.T) {
	withMode(ModeNone, func() {
		got := Concat(Bold, Red.FG, "hello", Reset)
		assert.Equal(t, "hello", got)
	})
}
