package tui

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App manages the interactive runtime loop for Rediscope TUI.
type App struct {
	State    *State
	Renderer *Renderer
	Stdout   io.Writer
	Stdin    io.Reader
}

// NewApp creates an App instance with default configuration.
func NewApp() *App {
	return &App{
		State:    NewDefaultState(),
		Renderer: NewRenderer(),
		Stdout:   os.Stdout,
		Stdin:    os.Stdin,
	}
}

// RenderToString renders the current layout string for testing or headless pipelines.
func (a *App) RenderToString(width, height int) string {
	return a.Renderer.Render(a.State, width, height)
}

// Run starts the interactive full-screen TUI loop.
func (a *App) Run() error {
	fIn, isFileIn := a.Stdin.(*os.File)
	if !isFileIn || !IsTerminal(fIn) {
		// Non-interactive fallback: render single frame to Stdout
		rows, cols := GetTerminalSize()
		fmt.Fprintln(a.Stdout, a.Renderer.Render(a.State, cols, rows))
		return nil
	}

	// 1. Enter Alternate Screen Buffer & Hide Cursor
	fmt.Fprint(a.Stdout, AltScreenOn+CursorHide)
	defer fmt.Fprint(a.Stdout, CursorShow+AltScreenOff)

	// 2. Put terminal into Raw Mode
	rawState, err := SetRawMode(int(fIn.Fd()))
	if err == nil {
		defer RestoreTerminalMode(int(fIn.Fd()), rawState)
	}

	// 3. Setup window resize (SIGWINCH) listener
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	defer signal.Stop(sigCh)

	// 4. Initial draw
	a.redraw()

	// 5. Input reading buffer
	inputBuf := make([]byte, 32)

	for {
		// Non-blocking select between resize signals, timers, and input
		// Read input byte
		n, err := fIn.Read(inputBuf)
		if err != nil && err != io.EOF {
			break
		}

		if n > 0 {
			shouldExit := a.handleInput(inputBuf[:n])
			if shouldExit {
				break
			}
			a.redraw()
		}

		// Check for resize signals
		select {
		case <-sigCh:
			a.redraw()
		default:
		}
	}

	return nil
}

func (a *App) redraw() {
	rows, cols := GetTerminalSize()
	frame := a.Renderer.Render(a.State, cols, rows)
	// Clear and jump to home (1,1) in alternate buffer
	fmt.Fprint(a.Stdout, ClearScreen+frame)
}

func (a *App) handleInput(buf []byte) bool {
	if len(buf) == 0 {
		return false
	}

	// Single byte characters
	b := buf[0]

	// Check for ESC key or ANSI arrow sequences
	if b == 27 { // ESC
		if len(buf) >= 3 && buf[1] == '[' {
			switch buf[2] {
			case 'A': // UP arrow
				a.prevNav()
				return false
			case 'B': // DOWN arrow
				a.nextNav()
				return false
			case 'C': // RIGHT arrow (drill)
				a.State.StatusMessage = fmt.Sprintf("Drill: %s", a.State.SelectedNav().Label)
				return false
			case 'D': // LEFT arrow (back)
				a.State.StatusMessage = "Back: root"
				return false
			}
		}
		// Standalone ESC -> Exit
		if len(buf) == 1 {
			return true
		}
		return false
	}

	switch b {
	case 'q', 'Q', 3: // 'q' or Ctrl+C
		return true

	case 'k', 'K': // Up (vim style)
		a.prevNav()

	case 'j', 'J': // Down (vim style)
		a.nextNav()

	case '1', '2', '3', '4', '5', '6', '7': // Number key direct jump
		idx := int(b - '1')
		if idx >= 0 && idx < len(a.State.NavItems) {
			a.State.ActiveNavIndex = idx
			a.State.StatusMessage = fmt.Sprintf("Nav: %s", a.State.SelectedNav().Label)
		}

	case 's', 'S': // Toggle Freeze
		a.State.Freeze = !a.State.Freeze
		a.State.StatusMessage = fmt.Sprintf("Freeze: %s", a.State.FormatFreeze())

	case 'p', 'P': // Toggle Pause
		a.State.IsPaused = !a.State.IsPaused
		if a.State.IsPaused {
			a.State.StatusMessage = "PAUSED"
		} else {
			a.State.StatusMessage = "LIVE"
		}

	case '+', '=': // Faster poll rate
		a.State.PollPeriod = max(0.2, a.State.PollPeriod-0.2)
		a.State.StatusMessage = fmt.Sprintf("Poll: %.1fs", a.State.PollPeriod)

	case '-', '_': // Slower poll rate
		a.State.PollPeriod = min(10.0, a.State.PollPeriod+0.2)
		a.State.StatusMessage = fmt.Sprintf("Poll: %.1fs", a.State.PollPeriod)

	case 'r', 'R': // Refresh
		a.State.StatusMessage = "REFRESHED (" + time.Now().Format("15:04:05") + ")"

	case '?': // Toggle help
		a.State.ShowHelp = !a.State.ShowHelp
		if a.State.ShowHelp {
			a.State.BodyTitle = "HELP / SHORTCUTS"
			a.State.BodyLines = []string{
				"",
				" KEYBOARD CONTROLS:",
				" -------------------",
				" [↑ / ↓] or [k / j] : Navigate Left Menu",
				" [1 - 7]            : Direct jump to View 1-7",
				" [s]                : Toggle Freeze (ON/OFF)",
				" [p]                : Toggle Pause / Resume",
				" [+ / -]            : Adjust Polling Speed",
				" [r]                : Manual Refresh",
				" [q] or [ESC]       : Exit TUI Dashboard",
				"",
				" Press '?' to close this help window.",
			}
		} else {
			a.State.BodyTitle = "BODY"
			a.State.BodyLines = nil
		}

	case '\r', '\n': // Enter key (select)
		a.State.StatusMessage = fmt.Sprintf("Selected: %s", a.State.SelectedNav().Label)
	}

	return false
}

func (a *App) nextNav() {
	if a.State.ActiveNavIndex < len(a.State.NavItems)-1 {
		a.State.ActiveNavIndex++
	} else {
		a.State.ActiveNavIndex = 0
	}
	a.State.StatusMessage = ""
}

func (a *App) prevNav() {
	if a.State.ActiveNavIndex > 0 {
		a.State.ActiveNavIndex--
	} else {
		a.State.ActiveNavIndex = len(a.State.NavItems) - 1
	}
	a.State.StatusMessage = ""
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
