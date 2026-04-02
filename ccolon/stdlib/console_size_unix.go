//go:build !windows

package stdlib

import (
	"syscall"
	"unsafe"
)

func getTerminalSize() (width, height int) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	ws := &winsize{}
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if err != 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}
