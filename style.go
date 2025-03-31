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

	ActiveLabel lipgloss.Style

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
}

func NewStyles() (s Styles) {
	s.Padding = lipgloss.NewStyle().Padding(1)

	s.Disabled = lipgloss.NewStyle().Foreground(grayColor)

	s.ActiveLabel = lipgloss.NewStyle().Foreground(labelColor)

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

	s.ActiveText = lipgloss.NewStyle().Foreground(whiteColor)
	s.Text = lipgloss.NewStyle().Foreground(lightGrayColor)
	s.Cursor = lipgloss.NewStyle().Foreground(whiteColor)

	return s
}

var (
	disabledStyle = lipgloss.NewStyle().Foreground(grayColor)

	errorHeaderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1F1F1")).Background(lipgloss.Color("#FF5F87")).Bold(true).Padding(0, 1).SetString("ERROR")
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	commentStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#757575")).PaddingLeft(1)

	sendButtonActiveStyle   = lipgloss.NewStyle().Background(labelColor).Foreground(yellowColor).Padding(0, 2)
	sendButtonInactiveStyle = lipgloss.NewStyle().Background(darkGrayColor).Foreground(lightGrayColor).Padding(0, 2)
	sendButtonStyle         = lipgloss.NewStyle().Background(darkGrayColor).Foreground(grayColor).Padding(0, 2)

	inlineCodeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Background(lipgloss.Color("#3A3A3A")).Padding(0, 1)
	linkStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AF87")).Underline(true)
)
