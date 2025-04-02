package main

import (
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type attachment string

func (a attachment) FilterValue() string {
	return string(a)
}

type attachmentDelegate struct {
	focused bool
}

func (d attachmentDelegate) Height() int {
	return 1
}

func (d attachmentDelegate) Spacing() int {
	return 0
}

func (d attachmentDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	styles := NewStyles()
	path := item.(attachment).FilterValue()
	style := styles.PreviousLabel
	if m.Index() == index && d.focused {
		style = styles.CurrentLabel
	}

	if m.Index() == index {
		_, _ = w.Write([]byte(style.Render("• " + path)))
	} else {
		_, _ = w.Write([]byte(style.Render("  " + path)))
	}
}

func (d attachmentDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
