//go:build windows

package repl

import (
	"os"
	"syscall"
)

const enableVirtualTerminalProcessing = 0x0004

func terminalSupportsColors() bool {
	handle := syscall.Handle(os.Stdout.Fd())
	var mode uint32
	if err := syscall.GetConsoleMode(handle, &mode); err != nil {
		return false
	}
	if mode&enableVirtualTerminalProcessing != 0 {
		return true
	}
	setConsoleMode := syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")
	if err := setConsoleMode.Find(); err != nil {
		return false
	}
	ok, _, _ := setConsoleMode.Call(uintptr(handle), uintptr(mode|enableVirtualTerminalProcessing))
	return ok != 0
}
