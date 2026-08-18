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

// Render generates the complete terminal screen string matching the exact layout spec with color styling.
func (r *Renderer) Render(state *State, totalWidth, totalHeight int) string {
	if totalWidth < 80 {
		totalWidth = 80
	}
	if totalHeight < 16 {
		totalHeight = 16
	}

	navW := r.NavWidth
	bodyW := totalWidth - 3 - navW
	if bodyW < 20 {
		bodyW = 20
		totalWidth = navW + bodyW + 3
	}

	numBodyRows := totalHeight - 6
	if numBodyRows < 12 {
		numBodyRows = 12
	}

	var sb strings.Builder

	// 1. Top border
	sb.WriteString(FgBorder)
	sb.WriteString("┌")
	sb.WriteString(strings.Repeat("─", totalWidth-2))
	sb.WriteString("┐")
	sb.WriteString(Reset)
	sb.WriteString("\n")

	// 2. Header Line
	// Colored Header Components
	hdrLeftFormatted := fmt.Sprintf(" %sREDISCOPE%s %s%s%s   %s%s%s   %sSource:%s %s%s%s",
		FgTitleRed, Reset,
		FgVersion, strings.TrimPrefix(state.AppTitle, "REDISCOPE "), Reset,
		FgTitleWhite, state.RedisVersion, Reset,
		FgCtxKey, Reset,
		FgSource, state.Source, Reset,
	)
	
	var pollBadge string
	if state.IsPaused {
		pollBadge = fmt.Sprintf("%sPoll:PAUSED%s   ", FgPollPaused, Reset)
	} else {
		pollBadge = fmt.Sprintf("%sPoll:%.1fs%s   ", FgPollLive, state.PollPeriod, Reset)
	}

	visLeft := VisibleWidth(hdrLeftFormatted)
	visRight := VisibleWidth(pollBadge)
	innerHdrWidth := totalWidth - 2

	padLen := innerHdrWidth - visLeft - visRight
	if padLen < 1 {
		padLen = 1
	}

	sb.WriteString(FgBorder + "│" + Reset)
	sb.WriteString(hdrLeftFormatted)
	sb.WriteString(strings.Repeat(" ", padLen))
	sb.WriteString(pollBadge)
	sb.WriteString(FgBorder + "│" + Reset)
	sb.WriteString("\n")

	// 3. Header Divider
	sb.WriteString(FgBorder)
	sb.WriteString("├")
	sb.WriteString(strings.Repeat("─", navW))
	sb.WriteString("┬")
	sb.WriteString(strings.Repeat("─", bodyW))
	sb.WriteString("┤")
	sb.WriteString(Reset)
	sb.WriteString("\n")

	// 4. Build Left Pane Rows
	leftRows := r.buildLeftPaneRows(state, navW, numBodyRows)

	// 5. Build Right Body Rows
	rightRows := r.buildRightBodyRows(state, bodyW, numBodyRows)

	// 6. Output Body Lines
	for i := 0; i < numBodyRows; i++ {
		lText := leftRows[i]
		rText := rightRows[i]

		lPadded := padToWidth(lText, navW)
		rPadded := padToWidth(rText, bodyW)

		sb.WriteString(FgBorder + "│" + Reset)
		sb.WriteString(lPadded)
		sb.WriteString(FgBorder + "│" + Reset)
		sb.WriteString(rPadded)
		sb.WriteString(FgBorder + "│" + Reset)
		sb.WriteString("\n")
	}

	// 7. Footer Divider
	sb.WriteString(FgBorder)
	sb.WriteString("├")
	sb.WriteString(strings.Repeat("─", navW))
	sb.WriteString("┴")
	sb.WriteString(strings.Repeat("─", bodyW))
	sb.WriteString("┤")
	sb.WriteString(Reset)
	sb.WriteString("\n")

	// 8. Footer Line
	footerText := r.buildColoredFooter(state, totalWidth-2)
	fPadded := padToWidth(footerText, totalWidth-2)
	sb.WriteString(FgBorder + "│" + Reset)
	sb.WriteString(fPadded)
	sb.WriteString(FgBorder + "│" + Reset)
	sb.WriteString("\n")

	// 9. Bottom Border
	sb.WriteString(FgBorder)
	sb.WriteString("└")
	sb.WriteString(strings.Repeat("─", totalWidth-2))
	sb.WriteString("┘")
	sb.WriteString(Reset)

	return sb.String()
}

func (r *Renderer) buildColoredFooter(state *State, innerWidth int) string {
	var statusPart string
	if state.StatusMessage != "" {
		statusPart = fmt.Sprintf(" %s %s %s", FgStatusBadge, state.StatusMessage, Reset)
	}

	// 1. Full 3-space version
	fullText := fmt.Sprintf("%s %s[↑/↓]%s %smove%s   %s[←/→]%s %sdrill/back%s   %s[Enter]%s %sselect%s   %s[s]%s %sfreeze%s   %s[r]%s %srefresh%s   %s[p]%s %spause%s   %s[+/-]%s %sspeed%s   %s[?]%s %shelp%s",
		statusPart,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
	)
	if VisibleWidth(fullText) <= innerWidth {
		return fullText
	}

	// 2. 2-space version for ~100-112 col terminals
	midText := fmt.Sprintf("%s %s[↑/↓]%s %smove%s  %s[←/→]%s %sdrill/back%s  %s[Enter]%s %sselect%s  %s[s]%s %sfreeze%s  %s[r]%s %srefresh%s  %s[p]%s %spause%s  %s[+/-]%s %sspeed%s  %s[?]%s %shelp%s",
		statusPart,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
	)
	if VisibleWidth(midText) <= innerWidth {
		return midText
	}

	// 3. Compact 1-space version
	compactText := fmt.Sprintf("%s %s[↑/↓]%s %smove%s %s[←/→]%s %sdrill%s %s[Enter]%s %sselect%s %s[s]%s %sfreeze%s %s[r]%s %srefresh%s %s[p]%s %spause%s %s[+/-]%s %sspeed%s %s[?]%s %shelp%s",
		statusPart,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
	)
	if VisibleWidth(compactText) <= innerWidth {
		return compactText
	}

	// 4. Short version for 80-col terminals
	shortText := fmt.Sprintf("%s %s[↑/↓]%s %smove%s %s[Enter]%s %ssel%s %s[s]%s %sfrz%s %s[p]%s %spause%s %s[+/-]%s %sspd%s %s[?]%s %shlp%s",
		statusPart,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
		FgKeyBracket, Reset, FgKeyLabel, Reset,
	)
	return shortText
}

func (r *Renderer) buildLeftPaneRows(state *State, navW, maxRows int) []string {
	rows := make([]string, maxRows)

	// Section 1 Header
	rows[0] = fmt.Sprintf(" %sFEATURE / NAV%s", FgNavHeader, Reset)
	rows[1] = fmt.Sprintf("%s%s%s", FgBorderDim, strings.Repeat("─", navW), Reset)

	// Nav Items
	rIdx := 2
	for i, item := range state.NavItems {
		if rIdx >= maxRows {
			break
		}
		if i == state.ActiveNavIndex {
			// Active Highlighted Row: Gold pointer with Cyan bracketed number and Crisp text
			rows[rIdx] = fmt.Sprintf(" %s>%s %s%s%s%s%s%s",
				FgNavActive, Reset,
				FgKeyBracket, item.Label[:3], Reset,
				FgTitleWhite, item.Label[3:], Reset,
			)
		} else {
			// Inactive Row: Dim cyan number + off-white label
			rows[rIdx] = fmt.Sprintf("   %s%s%s%s%s%s",
				FgNavNum, item.Label[:3], Reset,
				FgNavText, item.Label[3:], Reset,
			)
		}
		rIdx++
	}

	// Blank separator
	if rIdx < maxRows {
		rows[rIdx] = ""
		rIdx++
	}

	// Section 2: Context Header
	if rIdx < maxRows {
		rows[rIdx] = fmt.Sprintf("       %sContext%s        ", FgCtxHeader, Reset)
		rIdx++
	}
	if rIdx < maxRows {
		rows[rIdx] = fmt.Sprintf("%s%s%s", FgBorderDim, strings.Repeat("-", navW), Reset)
		rIdx++
	}

	// Freeze color
	var freezeVal string
	if state.Freeze {
		freezeVal = fmt.Sprintf("%sON%s", FgCtxFreezeOn, Reset)
	} else {
		freezeVal = fmt.Sprintf("%sOFF%s", FgCtxFreezeOff, Reset)
	}

	// Context Fields
	contextEntries := []string{
		fmt.Sprintf(" %sTarget:%s %s%s%s", FgCtxKey, Reset, FgCtxLive, state.Target, Reset),
		fmt.Sprintf(" %sScope:%s %s%s%s", FgCtxKey, Reset, FgCtxDb, state.Scope, Reset),
		fmt.Sprintf(" %sFreeze:%s %s", FgCtxKey, Reset, freezeVal),
		fmt.Sprintf(" %sTier:%s %s%s%s", FgCtxKey, Reset, FgCtxVal, state.Tier, Reset),
		fmt.Sprintf(" %sSelection:%s %s%s%s", FgCtxKey, Reset, FgCtxKey, state.Selection, Reset),
	}

	for _, entry := range contextEntries {
		if rIdx >= maxRows {
			break
		}
		rows[rIdx] = entry
		rIdx++
	}

	// Fill remaining rows
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
	formattedTitle := fmt.Sprintf("%s%s%s", FgBodyHeader, title, Reset)
	rows[0] = centerFormattedText(formattedTitle, bodyW)

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
	visW := VisibleWidth(s)
	if visW >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visW)
}

func centerFormattedText(s string, width int) string {
	visW := VisibleWidth(s)
	if visW >= width {
		return s
	}
	leftPad := (width - visW) / 2
	rightPad := width - visW - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}
