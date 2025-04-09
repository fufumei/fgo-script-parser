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
}

func GetTheme() Theme {
	theme := Theme{
		BodyColor:      "#D3C6AA",
		EmphasisColor:  "#E67E80",
		BorderColor:    "#5C6A72",
		PrimaryColor:   "#7FBBB3",
		SecondaryColor: "#83C092",
		TertiaryColor:  "#D699B6",
		SuccessColor:   "#8DA101",
		WarningColor:   "#5C6A72",
		InfoColor:      "#3A94C5",
		ErrorColor:     "#F85552",
		White:          "#DFDDC8",
		Gray:           "#5C6A72",
		Black:          "#343F44",
		SpinnerType:    spinner.Points,
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

func renderPaneTitle(s string, theme Theme) string {
	var title string
	title = s

	style := lipgloss.NewStyle().Foreground(theme.SecondaryColor).Padding(0, 1).Bold(true)
	return style.Render(title) + "\n\n"
}

func renderSelection(s string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.PrimaryColor)
	return style.Render(s)
}

func renderInactive(s string, theme Theme) string {
	style := lipgloss.NewStyle().Foreground(theme.Gray)
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
