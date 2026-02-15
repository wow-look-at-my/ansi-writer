# CLAUDE.md

## Project Overview

`ansi-writer` is a zero-dependency Go library for ANSI escape codes and terminal output manipulation. It provides color output with automatic downgrading based on terminal capabilities (truecolor → 256 → 16 → none), text styling, cursor control, and OSC sequences.

**Module:** `github.com/wow-look-at-my/ansi-writer`
**Go version:** 1.24.7

## Repository Structure

```
ansi-writer/
├── ansi.go          # Entire library implementation (~445 lines)
├── ansi_test.go     # Test suite (~465 lines)
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

2. **SGR style constants** — `Bold`, `Italic`, `Underline`, `Reset`, etc. Pre-built escape strings.

3. **Color mode system** — `ColorMode` enum (`ModeAuto`, `ModeTrueColor`, `Mode256`, `Mode16`, `ModeNone`) with thread-safe auto-detection from environment variables (`NO_COLOR`, `COLORTERM`, `TERM`). Uses `sync.Once` + `sync.Mutex`.

4. **Color type** — Core `Color` struct with `r, g, b uint8` and `idx16 int8` (index for named colors, -1 for pure RGB). Constructors: `RGB()`, `Hex()`, and internal `named()`. Methods: `FG()`, `BG()` which emit mode-appropriate escape sequences.

5. **Color downgrading** — Euclidean RGB distance calculations to find the closest match in 16-color and 256-color palettes. The 256-color matcher checks the 6x6x6 cube, grayscale ramp, and base 16 colors.

6. **Cursor control** — `Pos{X, Y}` type and `Cursor()` function supporting both absolute (`Abs`) and relative (`Rel`) positioning.

7. **Screen/scroll control** — `EraseScreen*`, `EraseLine*`, `ScrollUp()`, `ScrollDown()`, cursor save/restore/show/hide.

8. **OSC functions** — `Link()` for OSC 8 hyperlinks, `SetTitle()` for terminal window titles.

9. **Style helper** — `Style(text, codes...)` wraps text with escape codes and appends `Reset`.

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
- **Section comments** group related tests (style constants, cursor, erase, color modes, downgrading, etc.).
- Tests cover: all SGR constants, cursor positioning (absolute and relative), erase sequences, scroll, color mode detection, FG/BG output in all modes, color downgrading accuracy, OSC links/titles, and the `Style()` helper.

## Public API Surface

**Types:** `Color`, `ColorMode`, `Pos`

**Color constructors:** `RGB(r, g, b)`, `Hex(0xRRGGBB)`

**Named colors:** `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White` and bright variants (`BrightBlack`, `BrightRed`, etc.)

**Color methods:** `FG() string`, `BG() string`

**Mode control:** `SetMode(ColorMode)`, `GetMode() ColorMode`

**Cursor/screen:** `Cursor(Pos, bool)`, `ScrollUp(n)`, `ScrollDown(n)`, plus erase/save/restore/show/hide constants

**OSC:** `Link(url, text)`, `SetTitle(title)`

**Styling:** `Style(text, codes...)`
