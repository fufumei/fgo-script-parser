package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	} else if m.quitting {
		return ""
	}

	m.statePane.SetContent(m.statePaneContent())
	m.optionsPane.SetContent(m.optionsPaneContent())

	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.headerView(),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.statePane.View(),
			m.optionsPane.View(),
		),
		m.footerView(),
	)
}

func (m Model) statePaneContent() string {
	var sb strings.Builder
	steps := []string{"Source", "Type", "IDs", "Options", "Parse"}
	if len(m.results) > 0 {
		steps = append(steps, "Results")
	}
	sb.WriteString(renderPaneTitle("Steps", m.theme))
	paneWidth, _ := calculateViewportWidths(m.terminalWidth)

	for i, step := range steps {
		prefix := "◌ "
		if State(i) == m.currentState {
			prefix = "◉ "
		}

		if State(i) == m.currentState {
			sb.WriteString(renderSelected(prefix+truncateText(step, paneWidth), m.theme))
		} else {
			sb.WriteString(renderInactiveState(prefix+truncateText(step, paneWidth), m.theme))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) optionsPaneContent() string {
	switch m.currentState {
	case SourceSelect:
		return m.sourceSelectContent()
	case AtlasTypeSelect:
		return m.atlasIdTypeSelectContent()
	case IdInput:
		return m.idInputContent()
	case MiscOptions:
		return m.miscOptionsContent()
	case Confirm, Parsing:
		return m.parseContent()
	case Results:
		return m.resultsContent()
	}

	return "Something went wrong..."
}

func (m Model) sourceSelectContent() string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Foreground(m.theme.TertiaryColor).Render("Source"))
	sb.WriteString("\n")
	sb.WriteString(renderDefault("The source from which to fetch scripts.\nBoth options accept a list of scripts to parse.\nNote that parsing from Atlas requires an internet connection.", m.theme))
	sb.WriteString("\n\n")

	for _, o := range sourceOptions {
		prefix := "◌ "
		if m.selectedSource == o.value {
			prefix = "◉ "
		}

		if m.selectedSource == o.value {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", renderSelected(o.Title, m.theme), renderDescription(o.Description, m.theme)))
		} else {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", renderDefault(o.Title, m.theme), renderDescription(o.Description, m.theme)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) atlasIdTypeSelectContent() string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Foreground(m.theme.TertiaryColor).Render("Type"))
	sb.WriteString("\n")
	sb.WriteString(renderDefault("The type of Atlas ID to input.\nIt's currently not possible to parse multiple types at once.", m.theme))
	sb.WriteString("\n\n")

	for _, o := range atlasIdTypeOptions {
		prefix := "◌ "
		if m.selectedAtlasIdType == o.value {
			prefix = "◉ "
		}

		if m.selectedAtlasIdType == o.value {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", renderSelected(o.Title, m.theme), renderDescription(o.Description, m.theme)))
		} else {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", renderDefault(o.Title, m.theme), renderDescription(o.Description, m.theme)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) idInputContent() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.idInputDescriptionView(),
		m.IdInput.View(),
	)
}

// TODO: Clean this up like the above
func (m Model) miscOptionsContent() string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Foreground(m.theme.TertiaryColor).Render("Options"))
	sb.WriteString("\n")
	sb.WriteString(renderDefault("Miscellaneous options for parsing.", m.theme))
	sb.WriteString("\n\n")

	prefix := "☐ "
	if m.options.noFile {
		prefix = renderSelected("☑  ", m.theme)
	}
	title := renderSelected("No output file", m.theme)
	desc := renderDescription("If checked, the result will only print to the terminal,\notherwise also outputs to a csv on the same level as the script.", m.theme)

	sb.WriteString(
		lipgloss.JoinVertical(
			lipgloss.Left,
			prefix+title,
			desc,
		))

	return sb.String()
}

func (m Model) parseContent() string {
	var sb strings.Builder

	if m.currentState == Confirm {
		button := lipgloss.NewStyle().Background(m.theme.BorderColor).Foreground(m.theme.White).Padding(0, 2).Render("Parse")
		sb.WriteString(button)
	} else if m.currentState == Parsing {
		sb.WriteString(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.loadingSpinner.View()+" Parsing scripts...",
				"Elapsed time: "+m.timer.View(),
			),
		)

	}

	return sb.String()
}

func (m Model) resultsContent() string {
	return m.resultsTable.View()
}

func (m Model) headerView() string {
	title := headerStyle("FGO Script Parser", m.theme)
	line := strings.Repeat(lipgloss.NewStyle().Foreground(m.theme.BorderColor).Render("─"), max(0, m.terminalWidth-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(m.theme.Gray).
		Border(lipgloss.NormalBorder()).
		BorderTop(true).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderForeground(m.theme.BorderColor).
		Height(FooterHeight - 1). // top border
		Width(m.terminalWidth)

	notif := ""
	if m.err != nil {
		notif = renderError(m.err.Error(), m.theme)
		return footerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.help.View(m.keymap),
				notif,
			),
		)
	} else if m.notification.message != "" {
		notif = renderNotification(m.notification.message, m.theme)
		return footerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.help.View(m.keymap),
				notif,
			),
		)
	}

	return footerStyle.Render(m.help.View(m.keymap))
}

func (m Model) idInputDescriptionView() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(m.theme.TertiaryColor).Render("IDs"),
		renderDefault("Enter the IDs (if Atlas) or filepaths (if local) to parse from.\nOne ID/filepath per line.\nNote that filepaths cannot be relative.", m.theme),
		"\n",
	)
}

func calculateViewportWidths(terminalWidth int) (int, int) {
	paneOne := 17 // TODO: Magic number
	paneTwo := terminalWidth - paneOne
	// paneTwo := math.Floor(float64(terminalWidth) * 0.75)
	return int(paneOne), int(paneTwo)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
