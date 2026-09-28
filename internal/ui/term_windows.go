//go:build windows

package ui

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	getMode  = kernel32.NewProc("GetConsoleMode")
	setMode  = kernel32.NewProc("SetConsoleMode")
	setOutCP = kernel32.NewProc("SetConsoleOutputCP")
)

const (
	enableProcessedInput    = 0x0001
	enableLineInput         = 0x0002
	enableEchoInput         = 0x0004
	enableVirtualTermInput  = 0x0200
	enableVirtualTermOutput = 0x0004
)

// makeRaw switches the console to single-keypress mode with VT input, so
// arrow keys arrive as ANSI escape sequences (ESC [ A / ESC [ B).
func makeRaw() (func(), error) {
	h := os.Stdin.Fd()
	var old uint32
	if r, _, e := getMode.Call(h, uintptr(unsafe.Pointer(&old))); r == 0 {
		return nil, e
	}
	nw := (old &^ (enableProcessedInput | enableLineInput | enableEchoInput)) | enableVirtualTermInput
	if r, _, e := setMode.Call(h, uintptr(nw)); r == 0 {
		return nil, e
	}
	return func() { setMode.Call(h, uintptr(old)) }, nil
}

// enableVT turns on ANSI escape processing and UTF-8 output.
func enableVT() bool {
	h := os.Stdout.Fd()
	var mode uint32
	if r, _, _ := getMode.Call(h, uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	if r, _, _ := setMode.Call(h, uintptr(mode|enableVirtualTermOutput)); r == 0 {
		return false
	}
	setOutCP.Call(65001)
	return true
}

type coord struct{ X, Y int16 }
type smallRect struct{ Left, Top, Right, Bottom int16 }
type consoleInfo struct {
	Size       coord
	CursorPos  coord
	Attributes uint16
	Window     smallRect
	MaxSize    coord
}

// termSize returns the visible console rows and columns (24x80 if unknown).
func termSize() (rows, cols int) {
	var info consoleInfo
	r, _, _ := kernel32.NewProc("GetConsoleScreenBufferInfo").Call(os.Stdout.Fd(), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 24, 80
	}
	rows = int(info.Window.Bottom-info.Window.Top) + 1
	cols = int(info.Window.Right-info.Window.Left) + 1
	if rows <= 0 || cols <= 0 {
		return 24, 80
	}
	return rows, cols
}