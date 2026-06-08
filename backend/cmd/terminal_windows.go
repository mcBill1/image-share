//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

// enableANSI Windows下启用ANSI转义序列支持
func enableANSI() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")

	// STD_OUTPUT_HANDLE = -11 (0xFFFFFFF5)
	stdout, _, _ := getStdHandle.Call(uintptr(0xFFFFFFF5))

	var mode uint32
	getConsoleMode.Call(stdout, uintptr(unsafe.Pointer(&mode)))
	mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
	setConsoleMode.Call(stdout, uintptr(mode))
}