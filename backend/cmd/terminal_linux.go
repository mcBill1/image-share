//go:build linux

package main

// enableANSI Linux下无需额外启用ANSI支持
func enableANSI() {
	// Linux terminals natively support ANSI escape sequences
}

// checkTerminal Linux下无需检查终端
func checkTerminal() {
	// No need to check terminal on Linux
}
