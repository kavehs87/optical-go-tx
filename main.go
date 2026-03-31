package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
)

type styles struct {
	title     lipgloss.Style
	statusBar lipgloss.Style
	help      lipgloss.Style
	quitText  lipgloss.Style
}

func newStyles(isDark bool) styles {
	var s styles
	s.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("57")).
		Align(lipgloss.Center)
	s.statusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("247")).
		Background(lipgloss.Color("235"))
	s.help = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4).Bold(true)
	return s
}

func (m model) Init() tea.Cmd {
	return nil
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
