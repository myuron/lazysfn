package ui

// bulletWidth is the number of display columns occupied by " ●".
// The BLACK CIRCLE (U+25CF) is a full-width character that occupies 2 columns,
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
// The BLACK CIRCLE (●) is a full-width character and counts as 2 display columns.
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
