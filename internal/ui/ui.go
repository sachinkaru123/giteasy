// Package ui provides dependency-free terminal UI helpers: arrow-key select
// menus with a highlighted selection, styled text prompts, yes/no
// confirmations, boxed headers, and colored status lines.
//
// When stdin/stdout aren't a real terminal (pipes, CI, scripts) it falls
// back to plain numbered prompts, and colors honor the NO_COLOR variable.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	reader   = bufio.NewReader(os.Stdin)
	stdinTTY bool // stdin is an interactive terminal
	vtOK     bool // stdout is a terminal that understands ANSI sequences
	colorOn  bool // colors enabled
)

func init() {
	stdinTTY = isCharDevice(os.Stdin)
	if isCharDevice(os.Stdout) {
		vtOK = enableVT()
	}
	colorOn = vtOK && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb"
}

func isCharDevice(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// style wraps s in an ANSI SGR code when colors are enabled.
func style(code, s string) string {
	if !colorOn {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

// ---------------------------------------------------------------- headers

// Header prints a rounded box like:  ╭──────────────────╮
//
//	│  giteasy › Merge │
//	╰──────────────────╯
func Header(parts ...string) {
	text := strings.Join(append([]string{"giteasy"}, parts...), " › ")
	w := utf8.RuneCountInString(text) + 4
	bar := strings.Repeat("─", w)
	fmt.Println()
	fmt.Println(style("36", "╭"+bar+"╮"))
	fmt.Println(style("36", "│") + "  " + style("1;36", text) + "  " + style("36", "│"))
	fmt.Println(style("36", "╰"+bar+"╯"))
}

// ---------------------------------------------------------------- select

// Select shows a menu and returns the chosen index and value.
// Use ↑/↓ (or j/k), Enter to choose; number keys jump to an item.
func Select(title string, options []string) (int, string) {
	return SelectDefault(title, options, 0)
}

// SelectDefault is Select with an initially highlighted item.
func SelectDefault(title string, options []string, def int) (int, string) {
	if def < 0 || def >= len(options) {
		def = 0
	}
	if stdinTTY && vtOK {
		if idx, ok := selectArrows(title, options, def); ok {
			return idx, options[idx]
		}
	}
	return selectNumbered(title, options)
}

const maxVisible = 10

const (
	keyNone = iota
	keyUp
	keyDown
	keyEnter
	keyCancel
	keyHome
	keyEnd
	keyDigit = 100 // keyDigit+n for number keys 1-9
)

func parseKey(b []byte) int {
	if len(b) == 1 {
		switch b[0] {
		case 3:
			return keyCancel
		case 13, 10:
			return keyEnter
		case 'k':
			return keyUp
		case 'j':
			return keyDown
		}
		if b[0] >= '1' && b[0] <= '9' {
			return keyDigit + int(b[0]-'0')
		}
		return keyNone
	}
	if len(b) >= 3 && b[0] == 27 && (b[1] == '[' || b[1] == 'O') {
		switch b[2] {
		case 'A':
			return keyUp
		case 'B':
			return keyDown
		case 'H':
			return keyHome
		case 'F':
			return keyEnd
		}
	}
	return keyNone
}

// selectArrows renders the interactive menu. ok is false if the terminal
// couldn't be switched to raw mode (caller falls back to numbered input).
func selectArrows(title string, options []string, cur int) (int, bool) {
	restore, err := makeRaw()
	if err != nil {
		return 0, false
	}
	fmt.Print("\033[?25l") // hide cursor
	finish := func() {
		fmt.Print("\033[?25h")
		restore()
	}

	visible := len(options)
	if visible > maxVisible {
		visible = maxVisible
	}
	lines := visible + 2 // title + items + hint
	offset := 0
	first := true
	buf := make([]byte, 16)

	for {
		if cur < offset {
			offset = cur
		}
		if cur >= offset+visible {
			offset = cur - visible + 1
		}
		if !first {
			fmt.Printf("\033[%dA", lines)
		}
		first = false

		var sb strings.Builder
		sb.WriteString("\r\033[K" + style("1;36", "?") + " " + style("1", title) + "\r\n")
		for i := offset; i < offset+visible; i++ {
			sb.WriteString("\r\033[K")
			if i == cur {
				sb.WriteString(style("1;36", "❯ "+options[i]))
			} else {
				sb.WriteString("  " + options[i])
			}
			if i == offset && offset > 0 {
				sb.WriteString(style("2", "  ↑"))
			}
			if i == offset+visible-1 && offset+visible < len(options) {
				sb.WriteString(style("2", "  ↓"))
			}
			sb.WriteString("\r\n")
		}
		sb.WriteString("\r\033[K" + style("2", "  ↑/↓ move · enter select · ctrl+c cancel") + "\r\n")
		fmt.Print(sb.String())

		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			finish()
			os.Exit(1)
		}
		switch k := parseKey(buf[:n]); {
		case k == keyUp:
			cur = (cur - 1 + len(options)) % len(options)
		case k == keyDown:
			cur = (cur + 1) % len(options)
		case k == keyHome:
			cur = 0
		case k == keyEnd:
			cur = len(options) - 1
		case k > keyDigit:
			if d := k - keyDigit; d <= len(options) {
				cur = d - 1
			}
		case k == keyEnter:
			// Collapse the menu into a one-line summary of the answer.
			fmt.Printf("\033[%dA\r\033[J", lines)
			fmt.Printf("%s %s %s\r\n", style("1;32", "✔"), style("1", title), style("36", options[cur]))
			finish()
			return cur, true
		case k == keyCancel:
			fmt.Print("\r\n")
			finish()
			os.Exit(130)
		}
	}
}

func selectNumbered(title string, options []string) (int, string) {
	fmt.Println()
	fmt.Println(title)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	for {
		fmt.Print("Enter choice number: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < 1 || n > len(options) {
			fmt.Println("Invalid choice, try again.")
			continue
		}
		return n - 1, options[n-1]
	}
}

// ---------------------------------------------------------------- input

// Input prompts for a line of text. If defaultVal is non-empty it's shown
// and used verbatim when the user just presses Enter.
func Input(prompt, defaultVal string) string {
	q := style("1;36", "?") + " " + style("1", prompt)
	if defaultVal != "" {
		q += " " + style("2", "("+defaultVal+")")
	}
	fmt.Print(q + style("36", " › "))
	line, err := reader.ReadString('\n')
	if err != nil {
		os.Exit(1)
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

// Confirm asks a yes/no question (arrow keys on a terminal, y/n otherwise).
func Confirm(prompt string, defaultYes bool) bool {
	if stdinTTY && vtOK {
		def := 1
		if defaultYes {
			def = 0
		}
		if idx, ok := selectArrows(prompt, []string{"Yes", "No"}, def); ok {
			return idx == 0
		}
	}
	suffix := "y/N"
	if defaultYes {
		suffix = "Y/n"
	}
	fmt.Printf("%s (%s): ", prompt, suffix)
	line, err := reader.ReadString('\n')
	if err != nil {
		os.Exit(1)
	}
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defaultYes
	}
	return line == "y" || line == "yes"
}

// ---------------------------------------------------------------- status

func Info(format string, args ...interface{}) {
	fmt.Println(style("1;34", "→") + " " + fmt.Sprintf(format, args...))
}

func Success(format string, args ...interface{}) {
	fmt.Println(style("1;32", "✔") + " " + style("32", fmt.Sprintf(format, args...)))
}

func Warn(format string, args ...interface{}) {
	fmt.Println(style("1;33", "⚠") + " " + style("33", fmt.Sprintf(format, args...)))
}

func Error(format string, args ...interface{}) {
	fmt.Println(style("1;31", "✘") + " " + style("31", fmt.Sprintf(format, args...)))
}
