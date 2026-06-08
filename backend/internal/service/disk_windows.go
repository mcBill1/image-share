//go:build windows

package service

import (
	"syscall"
	"unsafe"
)

// getDiskUsage 获取磁盘总空间和可用空间 (Windows)
func getDiskUsage(path string) (total, free uint64, err error) {
	var freeBytes uint64
	var totalBytes uint64
	var availBytes uint64

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}

	ret, _, callErr := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&availBytes)),
	)
	if ret == 0 {
		return 0, 0, callErr
	}

	return totalBytes, freeBytes, nil
}
