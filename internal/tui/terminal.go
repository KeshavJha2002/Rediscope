package tui

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	AltScreenOn  = "\033[?1049h"
	AltScreenOff = "\033[?1049l"
	CursorHide   = "\033[?25l"
	CursorShow   = "\033[?25h"
	ClearScreen  = "\033[2J\033[H"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// GetTerminalSize returns the current (rows, cols) of stdout. Default fallback: (24, 110).
func GetTerminalSize() (int, int) {
	var ws winsize
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)
	if err != 0 || ws.Row == 0 || ws.Col == 0 {
		return 24, 110
	}
	return int(ws.Row), int(ws.Col)
}

// SetRawMode puts the terminal in raw mode for direct single-keystroke inputs.
func SetRawMode(fd int) (*syscall.Termios, error) {
	var old syscall.Termios
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&old)),
	)
	if err != 0 {
		return nil, err
	}

	raw := old
	// Disable echo, canonical mode, extended input processing, and interrupt signals if needed
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Iflag &^= syscall.IXON | syscall.ICRNL | syscall.BRKINT | syscall.INPCK | syscall.ISTRIP
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	_, _, err = syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&raw)),
	)
	if err != 0 {
		return nil, err
	}

	return &old, nil
}

// RestoreTerminalMode restores previous terminal settings.
func RestoreTerminalMode(fd int, old *syscall.Termios) error {
	if old == nil {
		return nil
	}
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(old)),
	)
	if err != 0 {
		return err
	}
	return nil
}

// IsTerminal checks if the file descriptor is a terminal.
func IsTerminal(f *os.File) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&termios)),
	)
	return err == 0
}
