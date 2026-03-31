package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	var v tea.View

	if m.width == 0 || m.height == 0 {
		v.Content = "Initializing..."
		return v
	}

	header := m.renderHeader()
	content := m.renderContent()
	footer := m.renderFooter()

	// Ensure content has a fixed height based on available space
	availableHeight := m.height - 2
	if availableHeight < 0 {
		availableHeight = 0
	}
	styledContent := lipgloss.NewStyle().Height(availableHeight).MaxHeight(availableHeight).Render(content)

	// Join all sections vertically. JoinVertical handles newlines between elements.
	v.Content = lipgloss.JoinVertical(lipgloss.Left, header, styledContent, footer)
	return v
}

func (m model) renderHeader() string {
	status := "[IDLE]"
	if m.state == stateEncoding {
		status = "[ENCODING]"
	} else if m.state == statePicking {
		status = "[PICKING]"
	}

	title := fmt.Sprintf(" Optical Transfer Encoder %s ", status)
	help := " 'o':open | 'l':loop | 'q':quit | '+/-':speed | 'space':pause "
	delayStr := fmt.Sprintf(" Delay: %v ", m.delay)
	
	loopStatus := "[LOOP: OFF]"
	if m.loop {
		loopStatus = "[LOOP: ON]"
	}

	content := lipgloss.JoinHorizontal(lipgloss.Center, title, help, delayStr, " ", loopStatus)
	return m.styles.title.Width(m.width).Render(content)
}

func (m model) renderContent() string {
	if m.state == statePicking {
		return m.filePicker.View()
	}

	if len(m.bitChunks) == 0 {
		return "\n   No data loaded. Press 'o' to select a file."
	}

	// Adjust available area for machine markers
	// We'll use the top-left block as a "Sync Marker" that toggles 
	// between White (7) and Black (0) every frame.
	syncColor := "0"
	if m.currentIndex%2 == 0 {
		syncColor = "7"
	}
	syncMarker := lipgloss.NewStyle().Background(lipgloss.Color(syncColor)).Render(" ")

	start := m.currentIndex * m.frameSize
	end := start + m.frameSize
	if end > len(m.bitChunks) {
		end = len(m.bitChunks)
	}

	var s strings.Builder
	for i := start; i < end; i++ {
		// Replace first block with sync marker
		if i == start {
			s.WriteString(syncMarker)
			continue
		}

		chunk := m.bitChunks[i]
		colorCode := fmt.Sprintf("%d", chunk)
		style := lipgloss.NewStyle().Background(lipgloss.Color(colorCode))
		s.WriteString(style.Render(" "))

		if (i-start+1)%m.width == 0 && i < end-1 {
			s.WriteRune('\n')
		}
	}

	return s.String()
}

func (m model) renderFooter() string {
	if len(m.bitChunks) == 0 {
		return m.styles.statusBar.Width(m.width).Render(" Ready")
	}

	totalFrames := (len(m.bitChunks) + m.frameSize - 1) / m.frameSize
	
	// Create a machine-readable binary progress bar
	// Every 3 bits of the frame index is encoded as one color block
	var machineProgress strings.Builder
	idx := uint32(m.currentIndex)
	for j := 0; j < 4; j++ { // Encode up to 4095 frames (12 bits)
		chunk := (idx >> (j * 3)) & 0x07
		style := lipgloss.NewStyle().Background(lipgloss.Color(fmt.Sprintf("%d", chunk)))
		machineProgress.WriteString(style.Render(" "))
	}

	progress := float64(m.currentIndex+1) / float64(totalFrames)
	barWidth := m.width - 40
	if barWidth < 10 {
		barWidth = 10
	}
	
	filled := int(progress * float64(barWidth))
	empty := barWidth - filled
	bar := "[" + strings.Repeat("=", filled) + strings.Repeat("-", empty) + "]"
	
	footerContent := fmt.Sprintf(" PAGE-SYNC: %s   | %s %d/%d frames", machineProgress.String(), bar, m.currentIndex+1, totalFrames)
	return m.styles.statusBar.Width(m.width).Render(footerContent)
}
