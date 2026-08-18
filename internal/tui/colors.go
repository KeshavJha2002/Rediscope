package tui

import "regexp"

// ANSI Color & Style Definitions
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Reverse   = "\033[7m"

	// Foreground Colors (Extended 256-color & Standard)
	FgBorder     = "\033[38;5;67m"  // Slate / Steel Blue border
	FgBorderDim  = "\033[38;5;240m" // Sub-divider gray
	FgTitleRed   = "\033[1;38;5;203m" // Redis red accent
	FgTitleWhite = "\033[1;97m"       // Bright bold white
	FgVersion    = "\033[38;5;215m"   // Warm peach / amber version
	FgSource     = "\033[38;5;114m"   // Mint green source
	FgPollLive   = "\033[1;38;5;84m"  // Emerald green poll
	FgPollPaused = "\033[1;38;5;208m" // Orange poll paused

	FgNavHeader  = "\033[1;38;5;117m" // Sky blue nav header
	FgNavActive  = "\033[1;38;5;220m" // Gold / Yellow active pointer
	FgNavActiveBg= "\033[48;5;24m\033[1;97m" // Navy blue bar for active row
	FgNavNum     = "\033[38;5;75m"    // Cyan for bracketed numbers
	FgNavText    = "\033[38;5;252m"   // Crisp off-white nav label

	FgCtxHeader  = "\033[1;38;5;141m" // Lavender context header
	FgCtxKey     = "\033[38;5;245m"   // Slate gray context keys
	FgCtxVal     = "\033[38;5;254m"   // Bright text
	FgCtxLive    = "\033[1;38;5;114m" // Green live target
	FgCtxDb      = "\033[1;38;5;75m"  // Cyan db scope
	FgCtxFreezeOff= "\033[38;5;245m"  // Gray freeze OFF
	FgCtxFreezeOn = "\033[1;38;5;196m" // Bright Red freeze ON

	FgBodyHeader = "\033[1;38;5;117m" // Sky blue body header
	FgBodyDim    = "\033[38;5;242m"   // Dim body placeholder

	FgKeyBracket = "\033[1;38;5;75m"  // Cyan shortcut keys
	FgKeyLabel   = "\033[38;5;250m"   // Light gray shortcut labels
	FgStatusBadge= "\033[1;30m\033[48;5;220m" // High-contrast amber badge
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes all ANSI escape sequences from a string.
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// VisibleWidth returns the character display width of a string ignoring ANSI codes.
func VisibleWidth(s string) int {
	return len([]rune(StripANSI(s)))
}
