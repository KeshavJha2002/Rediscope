package test

import (
	"strings"
	"testing"

	"rediscope/internal/tui"
)

func TestTUILayoutRenderingExactMatch(t *testing.T) {
	state := tui.NewDefaultState()
	renderer := tui.NewRenderer()

	width := 114
	height := 24
	rendered := renderer.Render(state, width, height)
	clean := tui.StripANSI(rendered)

	// 1. Verify Top Border
	if !strings.HasPrefix(clean, "┌") {
		t.Fatalf("expected top border starting with '┌', got:\n%s", clean)
	}

	// 2. Verify Header Contents
	if !strings.Contains(clean, "REDISCOPE v2:alpha") {
		t.Errorf("missing app title 'REDISCOPE v2:alpha'")
	}
	if !strings.Contains(clean, "Redis 7.4.2") {
		t.Errorf("missing Redis version 'Redis 7.4.2'")
	}
	if !strings.Contains(clean, "Source: 127.0.0.1:6379") {
		t.Errorf("missing Source '127.0.0.1:6379'")
	}
	if !strings.Contains(clean, "Poll:1.0s") {
		t.Errorf("missing Poll period 'Poll:1.0s'")
	}

	// 3. Verify Left Pane - Navigation Section
	if !strings.Contains(clean, "FEATURE / NAV") {
		t.Errorf("missing 'FEATURE / NAV' header")
	}
	if !strings.Contains(clean, "> [1] Namespace") {
		t.Errorf("missing active navigation '> [1] Namespace'")
	}
	if !strings.Contains(clean, "  [2] Serialized") {
		t.Errorf("missing nav item '[2] Serialized'")
	}
	if !strings.Contains(clean, "  [3] Physical Mem") {
		t.Errorf("missing nav item '[3] Physical Mem'")
	}
	if !strings.Contains(clean, "  [4] Mutable cmd") {
		t.Errorf("missing nav item '[4] Mutable cmd'")
	}
	if !strings.Contains(clean, "  [5] All cmds") {
		t.Errorf("missing nav item '[5] All cmds'")
	}
	if !strings.Contains(clean, "  [6] Snapshot Diff") {
		t.Errorf("missing nav item '[6] Snapshot Diff'")
	}
	if !strings.Contains(clean, "  [7] Cross-View") {
		t.Errorf("missing nav item '[7] Cross-View'")
	}

	// 4. Verify Left Pane - Context Section
	if !strings.Contains(clean, "Context") {
		t.Errorf("missing 'Context' section header")
	}
	if !strings.Contains(clean, "Target: live") {
		t.Errorf("missing context field 'Target: live'")
	}
	if !strings.Contains(clean, "Scope: db[0]") {
		t.Errorf("missing context field 'Scope: db[0]'")
	}
	if !strings.Contains(clean, "Freeze: OFF") {
		t.Errorf("missing context field 'Freeze: OFF'")
	}
	if !strings.Contains(clean, "Tier: root") {
		t.Errorf("missing context field 'Tier: root'")
	}
	if !strings.Contains(clean, "Selection: none") {
		t.Errorf("missing context field 'Selection: none'")
	}

	// 5. Verify Body Section
	if !strings.Contains(clean, "BODY") {
		t.Errorf("missing center 'BODY' label")
	}

	// 6. Verify Exact Match of Full Footer Keybindings on width 114
	expectedFooter := "[↑/↓] move   [←/→] drill/back   [Enter] select   [s] freeze   [r] refresh   [p] pause   [+/-] speed   [?] help"
	lines := strings.Split(clean, "\n")
	if !strings.Contains(clean, expectedFooter) {
		t.Errorf("missing full footer keybindings: %q in:\n%s", expectedFooter, lines[len(lines)-2])
	}

	// 7. Verify Bottom Border
	if !strings.HasSuffix(strings.TrimSpace(clean), "┘") {
		t.Errorf("expected bottom border ending with '┘'")
	}

	// 8. Verify Line Count matches requested height
	if len(lines) != height {
		t.Errorf("expected %d lines, got %d lines", height, len(lines))
	}
}

func TestTUIStateTransitions(t *testing.T) {
	app := tui.NewApp()

	// Initial active nav is 0 ([1] Namespace)
	if app.State.ActiveNavIndex != 0 {
		t.Fatalf("expected initial nav index 0, got %d", app.State.ActiveNavIndex)
	}

	// Move down ('j')
	app.State.ActiveNavIndex = 1
	rendered := app.RenderToString(110, 24)
	clean := tui.StripANSI(rendered)
	if !strings.Contains(clean, "  [1] Namespace") || !strings.Contains(clean, "> [2] Serialized") {
		t.Errorf("expected nav to highlight item 2:\n%s", clean)
	}

	// Toggle freeze
	app.State.Freeze = true
	renderedFreeze := app.RenderToString(110, 24)
	cleanFreeze := tui.StripANSI(renderedFreeze)
	if !strings.Contains(cleanFreeze, "Freeze: ON") {
		t.Errorf("expected 'Freeze: ON' after toggling freeze")
	}

	// Toggle pause
	app.State.IsPaused = true
	renderedPause := app.RenderToString(110, 24)
	cleanPause := tui.StripANSI(renderedPause)
	if !strings.Contains(cleanPause, "Poll:PAUSED") {
		t.Errorf("expected 'Poll:PAUSED' after toggling pause")
	}
}
