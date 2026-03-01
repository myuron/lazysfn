package ui

// bulletWidth is the number of display columns occupied by " ●".
// U+25CF has East Asian Width "Ambiguous"; display width is typically 1 in
// Western terminals and 2 in CJK terminals. We assume 2 columns for the bullet
// plus 1 column for the preceding space, giving 3 total.
const bulletWidth = 3

// calcPanelWidths returns left and right panel widths from the total terminal width.
// Ratio is left:right = 1:3.
func calcPanelWidths(totalWidth int) (leftW, rightW int) {
	leftW = totalWidth / 4
	rightW = totalWidth - leftW
	return
}

// formatSMLine formats a single state machine row for the left panel.
// If status is empty, returns name only. Otherwise appends " ●" with the name
// truncated so the bullet fits within panelWidth display columns.
// The BLACK CIRCLE (U+25CF, ●) has East Asian Width "Ambiguous" and is assumed
// to occupy 2 display columns in CJK terminals.
func formatSMLine(name, status string, panelWidth int) string {
	if status == "" {
		return name
	}

	// Reserve bulletWidth display columns (" ●") from the panel for the bullet indicator.
	nameWidth := panelWidth - bulletWidth
	if nameWidth < 0 {
		nameWidth = 0
	}

	runes := []rune(name)
	if len(runes) > nameWidth {
		runes = runes[:nameWidth]
	}

	return string(runes) + " \u25cf"
}
