// Package ui provides tiny, dependency-free terminal prompt helpers:
// numbered select menus, editable text input with a default value,
// and yes/no confirmations. Deliberately simple so giteasy has zero
// third-party dependencies.
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

// Select shows a numbered list and returns the chosen index and value.
// The user types a number and presses Enter. Ctrl+C exits the program.
func Select(title string, options []string) (int, string) {
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
		line = strings.TrimSpace(line)
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(options) {
			fmt.Println("Invalid choice, try again.")
			continue
		}
		return n - 1, options[n-1]
	}
}

// Input prompts for a line of text. If defaultVal is non-empty it's shown
// in brackets and used verbatim when the user just presses Enter.
func Input(prompt, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}
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

// Confirm asks a yes/no question. defaultYes controls what bare Enter means.
func Confirm(prompt string, defaultYes bool) bool {
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

// Println / Printf pass-throughs kept here so command code doesn't need
// to import "fmt" separately for simple status lines, keeping a single
// consistent look (e.g. we could add color/prefixing later in one place).
func Info(format string, args ...interface{}) {
	fmt.Printf("→ "+format+"\n", args...)
}

func Success(format string, args ...interface{}) {
	fmt.Printf("✔ "+format+"\n", args...)
}

func Warn(format string, args ...interface{}) {
	fmt.Printf("⚠ "+format+"\n", args...)
}

func Error(format string, args ...interface{}) {
	fmt.Printf("✘ "+format+"\n", args...)
}
