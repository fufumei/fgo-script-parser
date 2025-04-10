package main

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	HeaderHeight = 3
	FooterHeight = 2
)

type Theme struct {
	SpinnerType spinner.Spinner

	BodyColor      lipgloss.Color
	EmphasisColor  lipgloss.Color
	BorderColor    lipgloss.Color
	PrimaryColor   lipgloss.Color
	SecondaryColor lipgloss.Color
	TertiaryColor  lipgloss.Color
	SuccessColor   lipgloss.Color
	WarningColor   lipgloss.Color
	ErrorColor     lipgloss.Color
	InfoColor      lipgloss.Color
	White          lipgloss.Color
	Gray           lipgloss.Color
	Black          lipgloss.Color
	Gold           lipgloss.Color
}

func GetDefaultTheme() Theme {
	theme := Theme{
		BodyColor:      "#D3C6AA",
		EmphasisColor:  "#E67E80",
		BorderColor:    "#2C5CA4",
		PrimaryColor:   "#D1E3FA",
		SecondaryColor: "#0f81cf",
		TertiaryColor:  "#9EDEFF",
		SuccessColor:   "#8DA101",
		WarningColor:   "#5C6A72",
		InfoColor:      "#3A94C5",
		ErrorColor:     "#F85552",
		White:          "#DFDDC8",
		Gray:           "#75828a",
		Black:          "#343F44",
		Gold:           "#f4cf0b",
		SpinnerType:    spinner.Line,
	}
	return theme
}

func paneStyle(pos int, theme Theme) lipgloss.Style {
	if pos == 0 {
		return lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(theme.BorderColor).BorderRight(true)
	} else {
		return lipgloss.NewStyle().Padding(0, 0, 0, 3)
	}
}

func headerStyle(s string, theme Theme) string {
	b := lipgloss.RoundedBorder()
	b.Right = "├"
	style := lipgloss.NewStyle().BorderStyle(b).BorderForeground(theme.BorderColor).Foreground(theme.Gold).Padding(0, 1)
	return style.Render(s)
}

func renderPaneTitle(s string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.TertiaryColor).Padding(0, 1).Bold(true)
	return style.Render(s) + "\n\n"
}

func renderInactiveState(s string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.Gray)
	return style.Render(s)
}

func renderSelected(s string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.SecondaryColor)
	return style.Render(s)
}

func renderDescription(s string, theme Theme) string {
	style := lipgloss.NewStyle().PaddingLeft(2).Foreground(theme.Gray)
	return style.Render(s)
}

func renderDefault(s string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.PrimaryColor)
	return style.Render(s)
}

func renderError(s string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.ErrorColor)
	return style.Render(s)
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
