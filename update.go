package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

type clearErrMsg struct{}

func clearErrAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearErrMsg{}
	})
}

type notificationMsg struct {
	message string
}

func clearNotifAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return notificationMsg{}
	})
}

func sendNotificationMsg(msg string) tea.Cmd {
	return func() tea.Msg {
		return notificationMsg{message: msg}
	}
}

func (m Model) copyToClipboard() tea.Msg {
	row := m.resultsTable.SelectedRow()
	clipboard.Write(clipboard.FmtText, fmt.Appendf(nil, "%s, %s, %s", row[0], row[1], row[2]))
	return notificationMsg{message: "Row copied to clipboard!"}
}

func getTableColumns(totalWidth int, includeWordCount bool) []table.Column {
	if includeWordCount {
		return []table.Column{
			{Title: "Id", Width: int((float64(totalWidth)) * 0.1)},
			{Title: "Name", Width: int((float64(totalWidth)) * 0.4)},
			{Title: "Lines", Width: int((float64(totalWidth)) * 0.1)},
			{Title: "Characters", Width: int((float64(totalWidth)) * 0.15)},
			{Title: "Words", Width: int((float64(totalWidth)) * 0.25)},
		}
	} else {
		return []table.Column{
			{Title: "Id", Width: int((float64(totalWidth)) * 0.1)},
			{Title: "Name", Width: int((float64(totalWidth)) * 0.5)},
			{Title: "Lines", Width: int((float64(totalWidth)) * 0.15)},
			{Title: "Characters", Width: int((float64(totalWidth)) * 0.25)},
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case parseSuccessMsg:
		_, w2 := calculateViewportWidths(m.terminalWidth)
		var columns []table.Column
		var rows []table.Row

		columns = getTableColumns(w2, m.options.includeWordCount)

		if m.options.includeWordCount {
			for _, r := range msg {
				rows = append(rows, table.Row{r.id, r.name, fmt.Sprint(r.count.lines), fmt.Sprint(r.count.characters), fmt.Sprint(r.count.characters / 2)})
			}
		} else {
			for _, r := range msg {
				rows = append(rows, table.Row{r.id, r.name, fmt.Sprint(r.count.lines), fmt.Sprint(r.count.characters)})
			}
		}

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		styles := table.DefaultStyles()
		styles.Header = styles.Header.Foreground(m.theme.TertiaryColor)
		styles.Selected = styles.Selected.Foreground(m.theme.SecondaryColor)

		keys := table.KeyMap{
			LineUp:   key.NewBinding(key.WithKeys("up")),
			LineDown: key.NewBinding(key.WithKeys("down")),
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithHeight(m.terminalHeight-verticalMarginHeight),
			table.WithWidth(w2),
			table.WithStyles(styles),
			table.WithFocused(true),
			table.WithKeyMap(keys),
		)
		m.resultsTable = t
		m.results = msg
		m.currentState = Results
		cmds = append(cmds, m.timer.Stop(), m.timer.Reset(), sendNotificationMsg(fmt.Sprintf("Parsing completed in %s", m.timer.Elapsed().String())))
	case parseFailureMsg:
		m.err = msg
		m.currentState = Confirm
		cmds = append(cmds, tea.WindowSize(), clearErrAfter(5*time.Second), m.timer.Stop(), m.timer.Reset())
	case clearErrMsg:
		m.err = nil
		cmds = append(cmds, tea.WindowSize())
	case notificationMsg:
		m.notification = msg
		cmds = append(cmds, tea.WindowSize(), clearNotifAfter(2*time.Second))
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NextState):
			switch m.currentState {
			case SourceSelect:
				if m.selectedSource == atlas {
					m.currentState = AtlasTypeSelect
				} else {
					m.currentState = IdInput
					m.IdInput.Focus()
					m.IdInput.CursorEnd()
				}
			case AtlasTypeSelect:
				m.currentState = IdInput
				m.IdInput.Focus()
				m.IdInput.CursorEnd()
			case IdInput:
				m.IdInput.Blur()
				m.currentState = MiscOptions
			case MiscOptions:
				m.currentState = Confirm
			case Confirm:
				if len(m.results) > 0 {
					m.currentState = Results
				}
			}

		case key.Matches(msg, m.keymap.PrevState):
			switch m.currentState {
			case AtlasTypeSelect:
				m.currentState = SourceSelect
			case IdInput:
				if m.selectedSource == atlas {
					m.currentState = AtlasTypeSelect
				} else {
					m.currentState = SourceSelect
				}
				m.IdInput.Blur()
			case MiscOptions:
				m.currentState = IdInput
				m.IdInput.Focus()
				m.IdInput.CursorEnd()
			case Confirm:
				m.currentState = MiscOptions
			case Results:
				m.currentState = Confirm
			}

		case key.Matches(msg, m.keymap.NextOption):
			switch m.currentState {
			case SourceSelect:
				if int(m.selectedSource) < SourceMaxCount-1 {
					m.selectedSource = m.selectedSource + 1
				}
			case AtlasTypeSelect:
				if int(m.selectedAtlasIdType) < AtlasIdTypeMaxCount-1 {
					m.selectedAtlasIdType = m.selectedAtlasIdType + 1
				}
			case MiscOptions:
				if int(m.currentOption) < OptionsMaxCount-1 {
					if !(m.options.noFile && m.currentOption+1 == UniqueFileName) {
						m.currentOption = m.currentOption + 1
					}
				}
			}

		case key.Matches(msg, m.keymap.PrevOption):
			switch m.currentState {
			case SourceSelect:
				if int(m.selectedSource) > 0 {
					m.selectedSource = m.selectedSource - 1
				}
			case AtlasTypeSelect:
				if int(m.selectedAtlasIdType) > 0 {
					m.selectedAtlasIdType = m.selectedAtlasIdType - 1
				}
			case MiscOptions:
				if int(m.currentOption) > 0 {
					if !(m.options.noFile && m.currentOption-1 == UniqueFileName) {
						m.currentOption = m.currentOption - 1
					}

				}
			}

		case key.Matches(msg, m.keymap.Toggle):
			switch m.currentOption {
			case NoFile:
				m.options.noFile = !m.options.noFile
				m.options.uniqueFileName = false
			case IncludeWordCount:
				m.options.includeWordCount = !m.options.includeWordCount
			case UniqueFileName:
				m.options.uniqueFileName = !m.options.uniqueFileName
			}

		case key.Matches(msg, m.keymap.BlurInput):
			m.IdInput.Blur()

		case key.Matches(msg, m.keymap.FocusInput):
			m.IdInput.Focus()
			m.IdInput.CursorEnd()
			m.updateKeymap()
			return m, nil // Prevent new line from double enter input

		case key.Matches(msg, m.keymap.ClearInput):
			m.IdInput.Reset()
			m.IdInput.Focus()
			m.IdInput.CursorEnd()
			m.updateKeymap()

		case key.Matches(msg, m.keymap.Copy):
			cmds = append(cmds, m.copyToClipboard)

		case key.Matches(msg, m.keymap.Confirm):
			m.currentState = Parsing
			m.err = nil
			return m, tea.Batch(
				m.loadingSpinner.Tick,
				m.timer.Start(),
				m.parseScriptCmd(),
			)

		case key.Matches(msg, m.keymap.Quit):
			m.quitting = true
			m.abort = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		idInputDscriptionHeight := lipgloss.Height(m.idInputDescriptionView() + "\n")
		verticalMarginHeight := headerHeight + footerHeight
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		w1, w2 := calculateViewportWidths(msg.Width)
		paneMargin := 5 // Magic number

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.statePane = viewport.New(w1, msg.Height-verticalMarginHeight)
			m.statePane.Style = m.theme.paneStyle(0)
			m.statePane.YPosition = headerHeight
			m.statePane.SetContent(m.statePaneContent())

			m.optionsPane = viewport.New(w2, msg.Height-verticalMarginHeight)
			m.optionsPane.Style = m.theme.paneStyle(1)
			m.optionsPane.YPosition = headerHeight
			m.optionsPane.SetContent(m.optionsPaneContent())

			m.IdInput.SetHeight(msg.Height - verticalMarginHeight - idInputDscriptionHeight)
			m.IdInput.SetWidth(w2 - paneMargin)
			m.IdInput.FocusedStyle.CursorLine = lipgloss.NewStyle().Foreground(m.theme.SecondaryColor)

			m.loadingSpinner.Style = lipgloss.NewStyle().Foreground(m.theme.SecondaryColor)
			m.loadingSpinner.Spinner = m.theme.SpinnerType

			m.ready = true
		} else {
			m.statePane.Width = w1
			m.statePane.Height = msg.Height - verticalMarginHeight

			m.optionsPane.Width = w2
			m.optionsPane.Height = msg.Height - verticalMarginHeight

			m.IdInput.SetHeight(msg.Height - verticalMarginHeight - idInputDscriptionHeight)
			m.IdInput.SetWidth(w2 - paneMargin)

			m.resultsTable.SetColumns(getTableColumns(w2, m.options.includeWordCount))
		}
	}

	m.updateKeymap()

	// Handle keyboard events in the viewport
	m.statePane, cmd = m.statePane.Update(msg)
	cmds = append(cmds, cmd)
	m.IdInput, cmd = m.IdInput.Update(msg)
	cmds = append(cmds, cmd)
	if m.currentState == Parsing {
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.timer, cmd = m.timer.Update(msg)
	cmds = append(cmds, cmd)
	m.resultsTable, cmd = m.resultsTable.Update(msg)
	cmds = append(cmds, cmd)
	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
