package main

import (
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle global messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateFrameSize()
		// Always keep file picker height in sync
		m.filePicker.SetHeight(m.height - 5)
	}

	switch m.state {
	case statePicking:
		newFP, cmd := m.filePicker.Update(msg)
		m.filePicker = newFP
		cmds = append(cmds, cmd)

		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			data, err := os.ReadFile(path)
			if err != nil {
				return m, func() tea.Msg { return errorMsg(err) }
			}
			m.rawData = data
			m.bitChunks = bytesToBitChunks(data)
			m.state = stateEncoding
			m.currentIndex = 0
			return m, tick(m.delay)
		}

	default:
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "o":
				m.state = statePicking
				return m, m.filePicker.Init()
			case "l":
				m.loop = !m.loop
			case "+":
				m.delay += 10 * time.Millisecond
			case "-":
				m.delay -= 10 * time.Millisecond
				if m.delay < 10*time.Millisecond {
					m.delay = 10 * time.Millisecond
				}
			case " ": // Pause/Resume toggle
				if m.state == stateEncoding {
					m.state = stateIdle
				} else if len(m.bitChunks) > 0 {
					m.state = stateEncoding
					return m, tick(m.delay)
				}
			}

		case tickMsg:
			if m.state == stateEncoding {
				m.currentIndex++
				if m.currentIndex*m.frameSize >= len(m.bitChunks) {
					if m.loop {
						m.currentIndex = 0
					} else {
						m.currentIndex-- // Stay on last frame
						m.state = stateIdle
						return m, nil
					}
				}
				return m, tick(m.delay)
			}

		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateFrameSize() {
	// Header/Footer take dynamic space based on width.
	headerHeight := lipgloss.Height(m.renderHeader())
	footerHeight := lipgloss.Height(m.renderFooter())
	availableHeight := m.height - headerHeight - footerHeight
	if availableHeight < 0 {
		availableHeight = 0
	}
	m.frameSize = m.width * availableHeight
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func bytesToBitChunks(data []byte) []bitChunk {
	chunks := make([]bitChunk, 0, (len(data)*8+2)/3)
	var currentBits uint32
	var bitCount int

	for _, b := range data {
		currentBits = (currentBits << 8) | uint32(b)
		bitCount += 8
		for bitCount >= 3 {
			shift := bitCount - 3
			chunk := bitChunk((currentBits >> shift) & 0x07)
			chunks = append(chunks, chunk)
			bitCount -= 3
			currentBits &= (1 << bitCount) - 1
		}
	}
	// Final partial chunk if any
	if bitCount > 0 {
		chunk := bitChunk(currentBits << (3 - bitCount))
		chunks = append(chunks, chunk)
	}
	return chunks
}