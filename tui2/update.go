package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type clearErrMsg struct{}

func clearErrAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearErrMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case parseSuccessMsg:
		// TODO: Add hotkeys to copy single or all values
		m.results = msg

		_, w2 := calculateViewportWidths(m.terminalWidth)
		columns := []table.Column{
			{Title: "Name", Width: int((float64(w2)) * 0.5)},
			{Title: "Lines", Width: int((float64(w2)) * 0.25)},
			{Title: "Characters", Width: int((float64(w2)) * 0.25)},
		}

		var rows []table.Row
		for _, r := range msg {
			rows = append(rows, table.Row{r.name, fmt.Sprint(r.count.lines), fmt.Sprint(r.count.characters)})
		}

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.terminalHeight-verticalMarginHeight),
		)
		m.resultsTable = t
		// TODO: Add table styling

		m.currentState = Results
		m.resultsTable.Focus()
	case parseFailureMsg:
		m.err = msg
		m.currentState = Confirm
		cmds = append(cmds, tea.WindowSize(), clearErrAfter(10*time.Second))
	case clearErrMsg:
		m.err = nil
		cmds = append(cmds, tea.WindowSize())
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
				if m.selectedSource == atlas {
					m.selectedSource = local
				} else {
					m.selectedSource = atlas
				}
			case AtlasTypeSelect:
				if int(m.selectedAtlasIdType) < len(atlasIdTypeOptions)-1 {
					m.selectedAtlasIdType = m.selectedAtlasIdType + 1
				}
			}

		case key.Matches(msg, m.keymap.PrevOption):
			switch m.currentState {
			case SourceSelect:
				if m.selectedSource == atlas {
					m.selectedSource = local
				} else {
					m.selectedSource = atlas
				}
			case AtlasTypeSelect:
				if int(m.selectedAtlasIdType) > 0 {
					m.selectedAtlasIdType = m.selectedAtlasIdType - 1
				}
			}

		case key.Matches(msg, m.keymap.Toggle):
			m.options.noFile = !m.options.noFile

		case key.Matches(msg, m.keymap.BlurInput):
			m.IdInput.Blur()

		case key.Matches(msg, m.keymap.FocusInput):
			m.IdInput.Focus()
			m.IdInput.CursorEnd()
			m.updateKeymap()
			return m, nil // Prevent new line from double enter input

		case key.Matches(msg, m.keymap.Confirm):
			m.currentState = Parsing
			m.err = nil
			return m, tea.Batch(
				m.loadingSpinner.Tick,
				// TODO: This is buggy sometimes, only showing 0
				m.timer.Reset(),
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

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.statePane = viewport.New(w1, msg.Height-verticalMarginHeight)
			m.statePane.Style = paneStyle(0, m.theme)
			m.statePane.YPosition = headerHeight
			m.statePane.SetContent(m.statePaneContent())

			m.optionsPane = viewport.New(w2, msg.Height-verticalMarginHeight)
			m.optionsPane.Style = paneStyle(1, m.theme)
			m.optionsPane.YPosition = headerHeight
			m.optionsPane.SetContent(m.optionsPaneContent())

			m.IdInput.SetHeight(msg.Height - verticalMarginHeight - idInputDscriptionHeight)
			m.IdInput.SetWidth(w2)
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
			m.IdInput.SetWidth(w2)
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
