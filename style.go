package main

import (
	"github.com/charmbracelet/lipgloss"
)

const labelColor = lipgloss.Color("99")
const yellowColor = lipgloss.Color("#ECFD66")
const whiteColor = lipgloss.Color("255")
const grayColor = lipgloss.Color("241")
const darkGrayColor = lipgloss.Color("236")
const lightGrayColor = lipgloss.Color("247")

type Styles struct {
	Padding lipgloss.Style

	Disabled lipgloss.Style

	CurrentLabel  lipgloss.Style
	PreviousLabel lipgloss.Style

	ListBlock               lipgloss.Style
	ListItem                lipgloss.Style
	ItemTitle               lipgloss.Style
	SelectedItemTitle       lipgloss.Style
	DisabledItemTitle       lipgloss.Style
	ItemDescription         lipgloss.Style
	SelectedItemDescription lipgloss.Style
	DisabledItemDescription lipgloss.Style

	ActiveText lipgloss.Style
	Text       lipgloss.Style
	Cursor     lipgloss.Style

	CheckboxLabel       lipgloss.Style
	CheckboxDescription lipgloss.Style

	SendButtonActiveStyle   lipgloss.Style
	SendButtonInactiveStyle lipgloss.Style
	SendButtonStyle         lipgloss.Style

	Error        lipgloss.Style
	CommentStyle lipgloss.Style
	ActiveBanner lipgloss.Style
	Banner       lipgloss.Style
}

func NewStyles() (s Styles) {
	s.Padding = lipgloss.NewStyle().Padding(1)

	s.Disabled = lipgloss.NewStyle().Foreground(grayColor)

	s.CurrentLabel = lipgloss.NewStyle().Foreground(labelColor)
	s.PreviousLabel = lipgloss.NewStyle().Foreground(whiteColor)

	s.ListBlock = lipgloss.NewStyle().PaddingLeft(2)
	s.ItemTitle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)
	s.SelectedItemTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 0, 0, 1)
	s.DisabledItemTitle = s.ItemTitle.Foreground(grayColor)
	s.ItemDescription = s.ItemTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	s.SelectedItemDescription = s.SelectedItemTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"})
	s.DisabledItemDescription = s.ItemDescription.Foreground(grayColor)

	s.ActiveBanner = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true).
		BorderForeground(labelColor).
		Foreground(labelColor)

	s.Banner = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true).
		BorderForeground(whiteColor).
		Foreground(whiteColor)

	s.ActiveText = lipgloss.NewStyle().Foreground(whiteColor)
	s.Text = lipgloss.NewStyle().Foreground(lightGrayColor)
	s.Cursor = lipgloss.NewStyle().Foreground(whiteColor)

	s.SendButtonActiveStyle = lipgloss.NewStyle().Background(labelColor).Foreground(yellowColor).Padding(0, 2)
	s.SendButtonInactiveStyle = lipgloss.NewStyle().Background(darkGrayColor).Foreground(lightGrayColor).Padding(0, 2)
	s.SendButtonStyle = lipgloss.NewStyle().Background(darkGrayColor).Foreground(grayColor).Padding(0, 2)

	s.Error = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	s.CommentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#757575")).PaddingLeft(1)

	return s
}
