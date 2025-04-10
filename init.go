package main

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, tea.SetWindowTitle("fgo script parser"))
	return tea.Batch(cmds...)
}
