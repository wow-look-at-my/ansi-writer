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
	// Simple — one attribute
	fmt.Println(ansi.Red.FG("error!"))
	fmt.Println(ansi.Bold("important"))

	// RGB / Hex colors — downgraded automatically
	fmt.Println(ansi.RGB(255, 165, 0).FG("orange"))
	fmt.Println(ansi.Hex(0x1E90FF).FG("dodger blue"))

	// Chained — multiple attributes
	fmt.Println(ansi.Bold().FG(ansi.Yellow).Text("warning"))
	fmt.Println(ansi.Red.FG().BG(ansi.White).Italic().Text("fancy"))
}
```

## Colors

### Named colors

16 standard ANSI colors are available as package-level variables:

`Black` `Red` `Green` `Yellow` `Blue` `Magenta` `Cyan` `White`

`BrightBlack` `BrightRed` `BrightGreen` `BrightYellow` `BrightBlue` `BrightMagenta` `BrightCyan` `BrightWhite`

Each has `.FG()` and `.BG()` methods that return a `StyledText` for fluent chaining or immediate rendering:

```go
ansi.Red.FG("error")                   // quick: red text with auto-reset
ansi.Red.FG().Bold().Text("critical")  // chained: red + bold
```

### Custom colors

```go
ansi.RGB(255, 128, 0).FG("orange")
ansi.Hex(0xFF8000).FG("amber")
```

Custom colors are automatically downgraded to the nearest match when the terminal doesn't support truecolor:

| Terminal capability | What happens |
|---|---|
| Truecolor (`COLORTERM=truecolor`) | Exact 24-bit RGB |
| 256 colors (`TERM=*256color`) | Nearest from 6x6x6 cube, grayscale ramp, or base 16 |
| 16 colors (default) | Nearest standard ANSI color |
| No color (`NO_COLOR` set) | Plain text — no escape codes emitted |

### Forcing a color mode

```go
ansi.SetMode(ansi.ModeTrueColor) // force 24-bit
ansi.SetMode(ansi.Mode256)       // force 256
ansi.SetMode(ansi.Mode16)        // force 16
ansi.SetMode(ansi.ModeNone)      // disable all color
ansi.SetMode(ansi.ModeAuto)      // back to auto-detection
```

## Text styles

Style functions return `StyledText` — use them standalone or chain them:

```go
ansi.Bold("important")                          // quick
ansi.Bold().Underline().FG(ansi.Red).Text("!!") // chained
```

| Function | Effect |
|---|---|
| `Bold()` | Bold |
| `Dim()` | Dim / faint |
| `Italic()` | Italic |
| `Underline()` | Underline |
| `Blink()` | Blink |
| `RapidBlink()` | Rapid blink |
| `Reverse()` | Swap FG/BG |
| `Hidden()` | Hidden |
| `Strikethrough()` | Strikethrough |

`Reset` is an exported constant for manual string building. Each style also has a `Reset*` constant (e.g. `ResetBold`, `ResetItalic`).

## Reusable styles

`StyledText` is immutable — branching from a base style is safe:

```go
errStyle := ansi.Red.FG().Bold()
warnStyle := ansi.Yellow.FG().Bold()

fmt.Println(errStyle.Text("error: something broke"))
fmt.Println(warnStyle.Text("warning: check this"))
```

## `fmt.Stringer` support

`StyledText` implements `fmt.Stringer`, so it works directly with `fmt`:

```go
fmt.Println(ansi.Bold("hello"))
fmt.Printf("status: %s\n", ansi.Red.FG("FAIL"))
```

## Raw escape codes

When `StyledText` has no text set, `.String()` returns just the escape codes:

```go
codes := ansi.Red.FG().Bold().String()    // "\x1b[38;2;205;0;0m\x1b[1m"
fmt.Print(codes + "manual building" + ansi.Reset)
```

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

## License

See [LICENSE](LICENSE) for details.
