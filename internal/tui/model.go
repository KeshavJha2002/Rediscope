package tui

import "fmt"

// NavItem represents a selectable navigation menu entry.
type NavItem struct {
	Index int
	Key   string
	Label string
}

// DefaultNavItems defines the 7 core feature navigation views.
var DefaultNavItems = []NavItem{
	{Index: 1, Key: "1", Label: "[1] Namespace"},
	{Index: 2, Key: "2", Label: "[2] Serialized"},
	{Index: 3, Key: "3", Label: "[3] Physical Mem"},
	{Index: 4, Key: "4", Label: "[4] Mutable cmd"},
	{Index: 5, Key: "5", Label: "[5] All cmds"},
	{Index: 6, Key: "6", Label: "[6] Snapshot Diff"},
	{Index: 7, Key: "7", Label: "[7] Cross-View"},
}

// State represents the complete runtime visual and context state of the TUI.
type State struct {
	// Header fields
	AppTitle     string // e.g. "REDISCOPE v2:alpha"
	RedisVersion string // e.g. "Redis 7.4.2"
	Source       string // e.g. "127.0.0.1:6379"
	PollPeriod   float64 // e.g. 1.0 -> "Poll:1.0s"
	IsPaused     bool

	// Navigation state
	NavItems       []NavItem
	ActiveNavIndex int // 0 to len(NavItems)-1

	// Context fields
	Target    string // e.g. "live"
	Scope     string // e.g. "db[0]"
	Freeze    bool   // false -> "OFF", true -> "ON"
	Tier      string // e.g. "root"
	Selection string // e.g. "none"

	// Body content
	BodyTitle string
	BodyLines []string

	// Status / Notification overlay
	StatusMessage string
	ShowHelp      bool
}

// NewDefaultState creates an initialized state matching the default design specification.
func NewDefaultState() *State {
	return &State{
		AppTitle:       "REDISCOPE v2:alpha",
		RedisVersion:   "Redis 7.4.2",
		Source:         "127.0.0.1:6379",
		PollPeriod:     1.0,
		IsPaused:       false,
		NavItems:       DefaultNavItems,
		ActiveNavIndex: 0,
		Target:         "live",
		Scope:          "db[0]",
		Freeze:         false,
		Tier:           "root",
		Selection:      "none",
		BodyTitle:      "BODY",
		BodyLines:      nil,
		StatusMessage:  "",
		ShowHelp:       false,
	}
}

// FormatPollPeriod returns the formatted poll string e.g. "Poll:1.0s".
func (s *State) FormatPollPeriod() string {
	if s.IsPaused {
		return "Poll:PAUSED"
	}
	return fmt.Sprintf("Poll:%.1fs", s.PollPeriod)
}

// FormatFreeze returns "ON" or "OFF".
func (s *State) FormatFreeze() string {
	if s.Freeze {
		return "ON"
	}
	return "OFF"
}

// SelectedNav returns the currently active NavItem.
func (s *State) SelectedNav() NavItem {
	if s.ActiveNavIndex >= 0 && s.ActiveNavIndex < len(s.NavItems) {
		return s.NavItems[s.ActiveNavIndex]
	}
	return s.NavItems[0]
}
