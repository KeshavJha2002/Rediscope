package tui

import (
	"fmt"
	"strings"
)

// Renderer renders the full static frame and dynamic views into string buffers.
type Renderer struct {
	NavWidth int
}

// NewRenderer creates a Renderer with default 22-character left pane.
func NewRenderer() *Renderer {
	return &Renderer{
		NavWidth: 22,
	}
}

// Render generates the complete terminal screen string matching the exact layout spec.
func (r *Renderer) Render(state *State, totalWidth, totalHeight int) string {
	if totalWidth < 80 {
		totalWidth = 80
	}
	if totalHeight < 16 {
		totalHeight = 16
	}

	navW := r.NavWidth
	// totalWidth = 1 (left border) + navW (22) + 1 (divider) + bodyW + 1 (right border)
	bodyW := totalWidth - 3 - navW
	if bodyW < 20 {
		bodyW = 20
		totalWidth = navW + bodyW + 3
	}

	// Calculate vertical budget
	// 1: Top border
	// 2: Header line
	// 3: Header divider
	// 4..N-3: Body rows (numBodyRows)
	// N-2: Footer divider
	// N-1: Footer line
	// N: Bottom border
	numBodyRows := totalHeight - 6
	if numBodyRows < 12 {
		numBodyRows = 12
	}

	var sb strings.Builder

	// 1. Top border
	sb.WriteString("┌")
	sb.WriteString(strings.Repeat("─", totalWidth-2))
	sb.WriteString("┐\n")

	// 2. Header Line
	// Format: │ REDISCOPE v2:alpha   Redis 7.4.2   Source: 127.0.0.1:6379                                    Poll:1.0s       │
	hdrLeft := fmt.Sprintf(" %s   %s   Source: %s", state.AppTitle, state.RedisVersion, state.Source)
	hdrRight := fmt.Sprintf("%s   ", state.FormatPollPeriod())
	innerHdrWidth := totalWidth - 2

	padLen := innerHdrWidth - runewidth(hdrLeft) - runewidth(hdrRight)
	if padLen < 1 {
		padLen = 1
	}
	sb.WriteString("│")
	sb.WriteString(hdrLeft)
	sb.WriteString(strings.Repeat(" ", padLen))
	sb.WriteString(hdrRight)
	sb.WriteString("│\n")

	// 3. Header Divider
	sb.WriteString("├")
	sb.WriteString(strings.Repeat("─", navW))
	sb.WriteString("┬")
	sb.WriteString(strings.Repeat("─", bodyW))
	sb.WriteString("┤\n")

	// 4. Build Left Pane Rows
	leftRows := r.buildLeftPaneRows(state, navW, numBodyRows)

	// 5. Build Right Body Rows
	rightRows := r.buildRightBodyRows(state, bodyW, numBodyRows)

	// 6. Output Body Lines
	for i := 0; i < numBodyRows; i++ {
		lText := leftRows[i]
		rText := rightRows[i]

		// Ensure exact widths
		lPadded := padToWidth(lText, navW)
		rPadded := padToWidth(rText, bodyW)

		sb.WriteString("│")
		sb.WriteString(lPadded)
		sb.WriteString("│")
		sb.WriteString(rPadded)
		sb.WriteString("│\n")
	}

	// 7. Footer Divider
	sb.WriteString("├")
	sb.WriteString(strings.Repeat("─", navW))
	sb.WriteString("┴")
	sb.WriteString(strings.Repeat("─", bodyW))
	sb.WriteString("┤\n")

	// 8. Footer Line
	footerText := r.buildFooterText(state, totalWidth-2)
	fPadded := padToWidth(footerText, totalWidth-2)
	sb.WriteString("│")
	sb.WriteString(fPadded)
	sb.WriteString("│\n")

	// 9. Bottom Border
	sb.WriteString("└")
	sb.WriteString(strings.Repeat("─", totalWidth-2))
	sb.WriteString("┘")

	return sb.String()
}

func (r *Renderer) buildFooterText(state *State, innerWidth int) string {
	// 1. Full 3-space version (matching user prompt)
	fullText := " [↑/↓] move   [←/→] drill/back   [Enter] select   [s] freeze   [r] refresh   [p] pause   [+/-] speed   [?] help"
	if state.StatusMessage != "" {
		fullText = fmt.Sprintf(" [%s]%s", state.StatusMessage, fullText)
	}
	if runewidth(fullText) <= innerWidth {
		return fullText
	}

	// 2. 2-space version for ~100-112 col terminals
	midText := " [↑/↓] move  [←/→] drill/back  [Enter] select  [s] freeze  [r] refresh  [p] pause  [+/-] speed  [?] help"
	if state.StatusMessage != "" {
		midText = fmt.Sprintf(" [%s]%s", state.StatusMessage, midText)
	}
	if runewidth(midText) <= innerWidth {
		return midText
	}

	// 3. Compact 1-space version for medium screens (~85-100 col)
	compactText := " [↑/↓] move [←/→] drill/back [Enter] select [s] freeze [r] refresh [p] pause [+/-] speed [?] help"
	if state.StatusMessage != "" {
		compactText = fmt.Sprintf(" [%s]%s", state.StatusMessage, compactText)
	}
	if runewidth(compactText) <= innerWidth {
		return compactText
	}

	// 4. Short version for 80-col terminals
	shortText := " [↑/↓] move [Enter] select [s] freeze [p] pause [+/-] speed [?] help"
	if state.StatusMessage != "" {
		shortText = fmt.Sprintf(" [%s]%s", state.StatusMessage, shortText)
	}
	return shortText
}

func (r *Renderer) buildLeftPaneRows(state *State, navW, maxRows int) []string {
	rows := make([]string, maxRows)

	// Static Template Header
	rows[0] = " FEATURE / NAV"
	rows[1] = strings.Repeat("─", navW)

	// Nav Items
	rIdx := 2
	for i, item := range state.NavItems {
		if rIdx >= maxRows {
			break
		}
		if i == state.ActiveNavIndex {
			rows[rIdx] = fmt.Sprintf(" > %s", item.Label)
		} else {
			rows[rIdx] = fmt.Sprintf("   %s", item.Label)
		}
		rIdx++
	}

	// Blank separator
	if rIdx < maxRows {
		rows[rIdx] = ""
		rIdx++
	}

	// Context Header
	if rIdx < maxRows {
		rows[rIdx] = centerText("Context", navW)
		rIdx++
	}
	if rIdx < maxRows {
		rows[rIdx] = strings.Repeat("-", navW)
		rIdx++
	}

	// Context Fields
	contextEntries := []string{
		fmt.Sprintf(" Target: %s", state.Target),
		fmt.Sprintf(" Scope: %s", state.Scope),
		fmt.Sprintf(" Freeze: %s", state.FormatFreeze()),
		fmt.Sprintf(" Tier: %s", state.Tier),
		fmt.Sprintf(" Selection: %s", state.Selection),
	}

	for _, entry := range contextEntries {
		if rIdx >= maxRows {
			break
		}
		rows[rIdx] = entry
		rIdx++
	}

	// Fill remaining rows with blank strings
	for ; rIdx < maxRows; rIdx++ {
		rows[rIdx] = ""
	}

	return rows
}

func (r *Renderer) buildRightBodyRows(state *State, bodyW, maxRows int) []string {
	rows := make([]string, maxRows)

	// Row 0: Center BODY title
	title := state.BodyTitle
	if title == "" {
		title = "BODY"
	}
	rows[0] = centerText(title, bodyW)

	// Body content lines
	contentIdx := 0
	for i := 1; i < maxRows; i++ {
		if contentIdx < len(state.BodyLines) {
			rows[i] = " " + state.BodyLines[contentIdx]
			contentIdx++
		} else {
			rows[i] = ""
		}
	}

	return rows
}

func padToWidth(s string, width int) string {
	w := runewidth(s)
	if w >= width {
		// Truncate if exceeds
		runes := []rune(s)
		if len(runes) > width {
			return string(runes[:width])
		}
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func centerText(s string, width int) string {
	w := runewidth(s)
	if w >= width {
		return s
	}
	leftPad := (width - w) / 2
	rightPad := width - w - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

func runewidth(s string) int {
	return len([]rune(s))
}
