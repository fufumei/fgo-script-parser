package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const FooterHeight = 2

type Theme struct {
	Interface InterfaceStyles
	Text      TextStyles
	Table     TableStyles
	Help      help.Styles
}

// The styles for the interface wrapper
type InterfaceStyles struct {
	Title       lipgloss.Style
	Border      lipgloss.Style
	StatePane   lipgloss.Style
	OptionsPane lipgloss.Style
	Footer      lipgloss.Style
}

type TextStyles struct {
	StatePaneTitle    lipgloss.Style
	OptionsPaneTitle  lipgloss.Style
	Default           lipgloss.Style
	Highlighted       lipgloss.Style
	Darkened          lipgloss.Style
	OptionDescription lipgloss.Style
	Notification      lipgloss.Style
	Error             lipgloss.Style
}

type TableStyles struct {
	Header   lipgloss.Style
	Selected lipgloss.Style
}

func DefaultTheme(terminalWidth int) Theme {
	var t Theme

	var (
		gold      = lipgloss.Color("#f4cf0b")
		border    = lipgloss.Color("#2C5CA4")
		gray      = lipgloss.Color("#75828a")
		white     = lipgloss.Color("#D1E3FA")
		darkBlue  = lipgloss.Color("#0f81cf")
		lightBlue = lipgloss.Color("#9EDEFF")
		red       = lipgloss.Color("#F85552")
		green     = lipgloss.Color("#8DA101")
	)

	// Interface styles
	b := lipgloss.RoundedBorder()
	b.Right = "├"
	t.Interface.Title = lipgloss.NewStyle().BorderStyle(b).BorderForeground(border).Foreground(gold).Padding(0, 1)
	t.Interface.Border = lipgloss.NewStyle().Foreground(border)
	t.Interface.StatePane = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderForeground(border).BorderRight(true)
	t.Interface.OptionsPane = lipgloss.NewStyle().Padding(0, 0, 0, 3)
	t.Interface.Footer = lipgloss.NewStyle().Foreground(gray).Border(lipgloss.NormalBorder()).BorderTop(true).BorderBottom(false).
		BorderLeft(false).BorderRight(false).BorderForeground(border).Height(FooterHeight - 1). // top border
		Width(terminalWidth)

	// Text styles
	t.Text.StatePaneTitle = lipgloss.NewStyle().Foreground(lightBlue).Padding(0, 1).Bold(true)
	t.Text.OptionsPaneTitle = lipgloss.NewStyle().Foreground(lightBlue)
	t.Text.Default = lipgloss.NewStyle().Foreground(white)
	t.Text.Highlighted = lipgloss.NewStyle().Foreground(darkBlue)
	t.Text.Darkened = lipgloss.NewStyle().Foreground(gray)
	t.Text.OptionDescription = lipgloss.NewStyle().PaddingLeft(2).Foreground(gray)
	t.Text.Notification = lipgloss.NewStyle().Foreground(green)
	t.Text.Error = lipgloss.NewStyle().Foreground(red)

	// Table styles
	t.Table.Header = table.DefaultStyles().Header.Foreground(lightBlue)
	t.Table.Selected = table.DefaultStyles().Selected.Foreground(darkBlue)

	// Help styles
	t.Help = help.New().Styles

	return t
}

func truncateText(s string, w int) string {
	padding := 10
	if runewidth.StringWidth(s) <= w-padding {
		// Don't truncate strings that fit
		return s
	}

	runes := []rune(s)
	width := 0
	for i := len(runes) - 1; i >= 0; i-- {
		r := runes[i]
		width += runewidth.RuneWidth(r)
		if width >= w-padding {
			return "..." + string(runes[i+1:])
		}
	}
	return string(runes)
}
