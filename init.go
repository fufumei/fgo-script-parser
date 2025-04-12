package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"
)

func (m Model) Init() tea.Cmd {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, tea.SetWindowTitle("fgo script parser"))
	return tea.Batch(cmds...)
}
