package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	prefix           = "◌ "
	selectedPrefix   = "◉ "
	checkbox         = "☐ "
	selectedCheckbox = "☑  "
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
	// Nicer with a map but iteration over map is random order
	steps := []struct {
		state State
		name  string
	}{
		{state: SourceSelect, name: "Source"},
		{state: AtlasTypeSelect, name: "Type"},
		{state: IdInput, name: "IDs"},
		{state: MiscOptions, name: "Options"},
		{state: Confirm, name: "Parse"},
		{state: Results, name: "Results"},
	}

	var sb strings.Builder
	sb.WriteString(m.theme.Text.StatePaneTitle.Render("Steps") + "\n\n")
	paneWidth, _ := calculateViewportWidths(m.terminalWidth)

	for _, step := range steps {
		if step.state == Results && len(m.results) == 0 {
			continue
		}

		if step.state == m.currentState {
			sb.WriteString(m.theme.Text.Highlighted.Render(selectedPrefix + truncateText(step.name, paneWidth)))
		} else {
			sb.WriteString(m.theme.Text.Darkened.Render(prefix + truncateText(step.name, paneWidth)))
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
	options := []struct {
		title       string
		description string
		source      Source
	}{
		{title: "Atlas", description: "Parse from scripts fetched from Atlas DB", source: atlas},
		{title: "Local", description: "Parse from local files on your computer", source: local},
	}

	var sb strings.Builder
	sb.WriteString(m.theme.Text.OptionsPaneTitle.Render("Source"))
	sb.WriteString("\n")
	sb.WriteString(m.theme.Text.Default.Render("The source from which to fetch scripts.\nBoth options accept a list of scripts to parse.\nNote that parsing from Atlas requires an internet connection."))
	sb.WriteString("\n\n")

	for _, o := range options {
		if m.selectedSource == o.source {
			sb.WriteString(fmt.Sprintf(selectedPrefix+"%s\n%s\n", m.theme.Text.Highlighted.Render(o.title), m.theme.Text.OptionDescription.Render(o.description)))
		} else {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", m.theme.Text.Default.Render(o.title), m.theme.Text.OptionDescription.Render(o.description)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) atlasIdTypeSelectContent() string {
	options := []struct {
		title       string
		description string
		atlasType   AtlasIdType
	}{
		{title: "War", description: "Parse every script in a war (story chapter or event).\nEx: 100 for Fuyuki", atlasType: war},
		{title: "Quest", description: "Parse every script in a quest (war section or interlude etc).\nEx: 1000001 for Fuyuki chapter 1", atlasType: quest},
		{title: "Script", description: "Parse a list of specific scripts.\nEx: 0100000111 for Fuyuki chapter 1 post battle scene", atlasType: script},
	}

	var sb strings.Builder
	sb.WriteString(m.theme.Text.OptionsPaneTitle.Render("Type"))
	sb.WriteString("\n")
	sb.WriteString(m.theme.Text.Default.Render("The type of Atlas ID to input.\nIt's currently not possible to parse multiple types at once."))
	sb.WriteString("\n\n")

	for _, o := range options {
		if m.selectedAtlasIdType == o.atlasType {
			sb.WriteString(fmt.Sprintf(selectedPrefix+"%s\n%s\n", m.theme.Text.Highlighted.Render(o.title), m.theme.Text.OptionDescription.Render(o.description)))
		} else {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", m.theme.Text.Default.Render(o.title), m.theme.Text.OptionDescription.Render(o.description)))
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

func (m Model) miscOptionsContent() string {
	options := []struct {
		title       string
		description string
		option      OptionsEnum
	}{
		{title: "Include word count", description: "Calculates the approximate English word count per result.\nEnglish word count is conventionally half the character count.", option: IncludeWordCount},
		{title: "No output file", description: "Print results only to the terminal.\nIf unchecked, also outputs results to a file next to the application executable.", option: NoFile},
		{title: "Don't overwrite output file", description: "Append a random string to the output file name to prevent overwriting previous results.", option: UniqueFileName},
	}

	// TODO: Set up sections for these, like NoFile and UniqueFileName under a "Output File" section
	var sb strings.Builder
	sb.WriteString(m.theme.Text.OptionsPaneTitle.Render("Options"))
	sb.WriteString("\n")
	sb.WriteString(m.theme.Text.Default.Render("Miscellaneous options for parsing."))
	sb.WriteString("\n\n")

	for _, o := range options {
		prefix := checkbox
		switch o.option {
		case NoFile:
			if m.options.noFile {
				prefix = selectedCheckbox
			}
		case IncludeWordCount:
			if m.options.includeWordCount {
				prefix = selectedCheckbox
			}
		case UniqueFileName:
			if m.options.uniqueFileName {
				prefix = selectedCheckbox
			}
		}

		if o.option == UniqueFileName && m.options.noFile {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", m.theme.Text.Darkened.Render(o.title), m.theme.Text.OptionDescription.Render(o.description)))
		} else if m.currentOption == o.option {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", m.theme.Text.Highlighted.Render(o.title), m.theme.Text.OptionDescription.Render(o.description)))
		} else {
			sb.WriteString(fmt.Sprintf(prefix+"%s\n%s\n", m.theme.Text.Default.Render(o.title), m.theme.Text.OptionDescription.Render(o.description)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) parseContent() string {
	var sb strings.Builder
	if m.currentState == Confirm {
		// button := lipgloss.NewStyle().Background(m.theme.BorderColor).Foreground(m.theme.White).Padding(0, 2).Render("Parse")
		sb.WriteString("Parse")
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
	title := m.theme.Interface.Title.Render("FGO Script Parser")
	line := strings.Repeat(m.theme.Interface.Border.Render("─"), max(0, m.terminalWidth-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	notif := ""
	if m.err != nil {
		notif = m.theme.Text.Error.Render(strings.ToUpper(m.err.Error()[:1]) + m.err.Error()[1:])
		return m.theme.Interface.Footer.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.help.View(m.keymap),
				notif,
			),
		)
	} else if m.notification.message != "" {
		notif = m.theme.Text.Notification.Render(m.notification.message)
		return m.theme.Interface.Footer.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.help.View(m.keymap),
				notif,
			),
		)
	}

	return m.theme.Interface.Footer.Render(m.help.View(m.keymap))
}

func (m Model) idInputDescriptionView() string {
	var sb strings.Builder
	switch m.selectedSource {
	case atlas:
		switch m.selectedAtlasIdType {
		case war:
			sb.WriteString("Enter the war IDs to parse from.")
		case quest:
			sb.WriteString("Enter the quest IDs to parse from.")
		case script:
			sb.WriteString("Enter the script IDs to parse from.")
		}
		sb.WriteString("\nOnly one ID per line.")
	case local:
		sb.WriteString("Enter the filepaths to local files to parse from.")
		sb.WriteString("\nFilepath can point to a directory or directly to a file (must include file extension).")
		sb.WriteString("\nOnly one filepath per line.\nNote that filepaths must be absolute.")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.theme.Text.OptionsPaneTitle.Render("IDs"),
		m.theme.Text.Default.Render(sb.String()),
		"\n",
	)
}

func calculateViewportWidths(terminalWidth int) (int, int) {
	paneOne := 17 // TODO: Magic number
	paneTwo := terminalWidth - paneOne
	return int(paneOne), int(paneTwo)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
