//go:build windows

package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
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

// isDoubleClick Windows下检测是否双击运行（非终端启动）
func isDoubleClick() bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")

	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd == 0 {
		return true
	}

	user32 := syscall.NewLazyDLL("user32.dll")
	getWindowThreadProcessId := user32.NewProc("GetWindowThreadProcessId")

	var consoleOwnerPid uint32
	getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&consoleOwnerPid)))

	return int(consoleOwnerPid) == os.Getpid()
}

// checkTerminal 检测是否从终端启动，禁止双击运行
func checkTerminal() {
	if isDoubleClick() {
		fmt.Println("[错误] 不支持双击运行此程序")
		fmt.Println("[提示] 请在 CMD 或 PowerShell 中运行")
		fmt.Println("[提示] 打开方式: Win+R -> cmd -> cd到程序目录 -> 运行此程序")
		fmt.Println()
		fmt.Println()
		fmt.Println("程序将在 3 秒后退出...")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}
}
