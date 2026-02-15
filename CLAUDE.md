# CLAUDE.md

## Project Overview

`ansi-writer` is a zero-dependency Go library for ANSI escape codes and terminal output manipulation. It provides color output with automatic downgrading based on terminal capabilities (truecolor → 256 → 16 → none), text styling, cursor control, and OSC sequences.

**Module:** `github.com/wow-look-at-my/ansi-writer`
**Go version:** 1.24.7

## Repository Structure

```
ansi-writer/
├── ansi.go          # Entire library implementation (~470 lines)
├── ansi_test.go     # Test suite (~560 lines)
└── go.mod           # Module definition (zero external dependencies)
```

This is a single-package library. All code lives in `ansi.go` with tests in `ansi_test.go`.

## Commands

### Run tests
```
go test ./...
```

### Run tests with verbose output
```
go test -v ./...
```

### Run a specific test
```
go test -run TestName ./...
```

### Format code
```
gofmt -w .
```

### Vet code
```
go vet ./...
```

There is no Makefile, CI configuration, or linter config. Standard Go tooling applies.

## Architecture

The library is organized into these logical sections within `ansi.go`:

1. **Base escape sequences** — Constants `ESC`, `CSI`, `OSC`, `ST` that form the foundation for all codes.

2. **Style type** — `Style` is a struct with a `String()` method that returns an empty string in `ModeNone`. Color styles store RGB + idx16 data and resolve lazily against the current color mode. Because `Style` is a struct, `fmt.Sprint` inserts spaces between adjacent `Style` values; use `fmt.Sprintf` with `%s` verbs to avoid this.

3. **SGR style and reset codes** — Exported `Style` vars: `Bold`, `Dim`, `Italic`, `Underline`, `Blink`, `RapidBlink`, `Reverse`, `Hidden`, `Strikethrough`, plus `Reset` and individual `Reset*` codes. Used directly with fmt: `fmt.Print(ansi.Bold, "text", ansi.Reset)`.

4. **Color mode system** — `ColorMode` enum (`ModeAuto`, `ModeTrueColor`, `Mode256`, `Mode16`, `ModeNone`) with thread-safe auto-detection from environment variables (`NO_COLOR`, `COLORTERM`, `TERM`). Uses `sync.Once` + `sync.Mutex`.

5. **Color type** — Core `Color` struct with `r, g, b uint8`, `idx16 int8` (index for named colors, -1 for pure RGB), and `FG`/`BG` fields of type `Style`. Constructors: `RGB()`, `Hex()`, and internal `named()` (all via `newColor()`). Color styles are lazily resolved — `Style.String()` checks the current color mode and computes the appropriate escape code.

6. **Color downgrading** — Euclidean RGB distance calculations to find the closest match in 16-color and 256-color palettes. The 256-color matcher checks the 6x6x6 cube, grayscale ramp, and base 16 colors.

7. **Cursor control** — `Pos{X, Y}` type and `Cursor()` function supporting both absolute (`Abs`) and relative (`Rel`) positioning.

8. **Screen/scroll control** — `EraseScreen*`, `EraseLine*`, `ScrollUp()`, `ScrollDown()`, cursor save/restore/show/hide.

9. **OSC functions** — `Link()` for OSC 8 hyperlinks, `SetTitle()` for terminal window titles.

## Code Conventions

- **Naming:** Exported names use PascalCase; unexported use camelCase. Escape sequence constants use SCREAMING_SNAKE for base sequences (`ESC`, `CSI`).
- **Section organization:** Logical sections in `ansi.go` are separated by comment blocks with dashes.
- **No error returns:** The library assumes valid input and does not return errors.
- **Thread safety:** Global color mode state is protected by `sync.Mutex`.
- **String building:** Uses `strings.Builder` for multi-part escape sequences, direct concatenation for simple ones.
- **No interfaces:** Simple function/method API only.

## Testing Conventions

- **Table-driven tests** using anonymous structs with descriptive field names.
- **Environment manipulation** via `t.Setenv()` for color mode detection tests.
- **Helper function** `withMode(m ColorMode, fn func())` temporarily sets the color mode for a test block, then restores it.
- **Section comments** group related tests (SGR codes, fmt integration, cursor, erase, color modes, downgrading, ModeNone, etc.).
- Tests cover: all SGR codes, Style with fmt.Sprint, Style in ModeNone, cursor positioning (absolute and relative), erase sequences, scroll, color mode detection, FG/BG raw codes in all modes, FG/BG as Style, color downgrading accuracy, OSC links/titles.

## Public API Surface

**Types:** `Color`, `ColorMode`, `Pos`, `Style`

**Color constructors:** `RGB(r, g, b)`, `Hex(0xRRGGBB)`

**Named colors:** `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White` and bright variants (`BrightBlack`, `BrightRed`, etc.)

**Color fields:** `FG Style`, `BG Style` (lazily resolved in `String()`)

**Style vars:** `Bold`, `Dim`, `Italic`, `Underline`, `Blink`, `RapidBlink`, `Reverse`, `Hidden`, `Strikethrough`

**Reset vars:** `Reset`, `ResetBold`, `ResetDim`, `ResetItalic`, `ResetUnderline`, `ResetBlink`, `ResetReverse`, `ResetHidden`, `ResetStrikethrough`

**Mode control:** `SetMode(ColorMode)`, `GetMode() ColorMode`

**Cursor/screen:** `Cursor(Pos, bool)`, `ScrollUp(n)`, `ScrollDown(n)`, plus erase/save/restore/show/hide constants

**OSC:** `Link(url, text)`, `SetTitle(title)`
