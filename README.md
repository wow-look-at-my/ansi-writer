# ansi-writer

A small Go library for ANSI escape codes. Specify colors by name or RGB — the library detects terminal capabilities and automatically downgrades to the best available depth (truecolor → 256 → 16 → none).

Zero dependencies. Single file.

## Install

```
go get github.com/wow-look-at-my/ansi-writer
```

## Quick start

```go
package main

import (
	"fmt"

	ansi "github.com/wow-look-at-my/ansi-writer"
)

func main() {
	// Named colors
	fmt.Println(ansi.Red.FG() + "error!" + ansi.Reset)

	// RGB / Hex colors — downgraded automatically
	fmt.Println(ansi.RGB(255, 165, 0).FG() + "orange" + ansi.Reset)
	fmt.Println(ansi.Hex(0x1E90FF).FG() + "dodger blue" + ansi.Reset)

	// Style helper applies codes and appends Reset for you
	fmt.Println(ansi.Style("warning", ansi.Bold, ansi.Yellow.FG()))

	// Background colors
	fmt.Println(ansi.Style("highlighted", ansi.White.FG(), ansi.Blue.BG()))
}
```

## Colors

### Named colors

16 standard ANSI colors are available as package-level variables:

`Black` `Red` `Green` `Yellow` `Blue` `Magenta` `Cyan` `White`

`BrightBlack` `BrightRed` `BrightGreen` `BrightYellow` `BrightBlue` `BrightMagenta` `BrightCyan` `BrightWhite`

Each has `.FG()` and `.BG()` methods that return the appropriate escape sequence string.

### Custom colors

```go
ansi.RGB(255, 128, 0)  // from RGB components
ansi.Hex(0xFF8000)     // from a hex value
```

Custom colors are automatically downgraded to the nearest match when the terminal doesn't support truecolor:

| Terminal capability | What happens |
|---|---|
| Truecolor (`COLORTERM=truecolor`) | Exact 24-bit RGB |
| 256 colors (`TERM=*256color`) | Nearest from 6x6x6 cube, grayscale ramp, or base 16 |
| 16 colors (default) | Nearest standard ANSI color |
| No color (`NO_COLOR` set) | Empty string — output is clean |

### Forcing a color mode

```go
ansi.SetMode(ansi.ModeTrueColor) // force 24-bit
ansi.SetMode(ansi.Mode256)       // force 256
ansi.SetMode(ansi.Mode16)        // force 16
ansi.SetMode(ansi.ModeNone)      // disable all color
ansi.SetMode(ansi.ModeAuto)      // back to auto-detection
```

## Text styles

All styles are plain strings — concatenate them directly:

```go
fmt.Print(ansi.Bold + ansi.Underline + "important" + ansi.Reset)
```

| Constant | Effect |
|---|---|
| `Bold` | Bold |
| `Dim` | Dim / faint |
| `Italic` | Italic |
| `Underline` | Underline |
| `Blink` | Blink |
| `RapidBlink` | Rapid blink |
| `Reverse` | Swap FG/BG |
| `Hidden` | Hidden |
| `Strikethrough` | Strikethrough |
| `Reset` | Reset all attributes |

Each has a corresponding `Reset*` variant (e.g. `ResetBold`, `ResetItalic`).

## Cursor movement

```go
// Absolute — move to row 5, column 10 (1-based)
fmt.Print(ansi.Cursor(ansi.Pos{X: 10, Y: 5}, ansi.Abs))

// Relative — move 3 down and 2 right from current position
fmt.Print(ansi.Cursor(ansi.Pos{X: 2, Y: 3}, ansi.Rel))

// Relative — move 1 up and 4 left
fmt.Print(ansi.Cursor(ansi.Pos{X: -4, Y: -1}, ansi.Rel))
```

Save and restore cursor position:

```go
fmt.Print(ansi.CursorSave)
// ... move around and draw ...
fmt.Print(ansi.CursorRestore)
```

Show/hide cursor: `ansi.CursorShow`, `ansi.CursorHide`

## Screen control

```go
fmt.Print(ansi.EraseScreen)     // clear entire screen
fmt.Print(ansi.EraseLine)       // clear current line
fmt.Print(ansi.ScrollUp(3))     // scroll up 3 lines
fmt.Print(ansi.ScrollDown(1))   // scroll down 1 line
```

Erase variants: `EraseScreenToEnd`, `EraseScreenToStart`, `EraseScreen`, `EraseScreenAll`, `EraseLineToEnd`, `EraseLineToStart`, `EraseLine`

## Hyperlinks and titles

```go
// Clickable hyperlink (terminals with OSC 8 support)
fmt.Println(ansi.Link("https://example.com", "click here"))

// Set terminal window title
fmt.Print(ansi.SetTitle("my app"))
```

## The `Style` helper

`Style` wraps text with any number of escape sequences and appends `Reset` automatically:

```go
ansi.Style("text", ansi.Bold, ansi.Red.FG())
// equivalent to: ansi.Bold + ansi.Red.FG() + "text" + ansi.Reset
```

## License

See [LICENSE](LICENSE) for details.
