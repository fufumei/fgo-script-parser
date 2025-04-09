package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Source int

const (
	atlas Source = iota
	local
)

type AtlasIdType int

const (
	war AtlasIdType = iota
	quest
	script
)

type SourceOption struct {
	Title       string
	Description string
	value       Source
}

type AtlasIdTypeOption struct {
	Title       string
	Description string
	value       AtlasIdType
}

var (
	sourceOptions = []SourceOption{
		{
			Title:       "Atlas",
			Description: "Parse directly from Atlas IDs",
			value:       atlas,
		},
		{
			Title:       "Local",
			Description: "Parse from local files on your computer",
			value:       local,
		}}

	atlasIdTypeOptions = []AtlasIdTypeOption{
		{
			Title:       "War",
			Description: "Parse every script in a war (story chapter or event).\nEx: 100 for Fuyuki",
			value:       war,
		},
		{
			Title:       "Quest",
			Description: "Parse every script in a quest (war section or interlude etc).\nEx: 1000001 for Fuyuki chapter 1",
			value:       quest,
		},
		{
			Title:       "Script",
			Description: "Parse specific scripts individually.\nEx: 0100000111 for Fuyuki chapter 1 post battle scene",
			value:       script,
		},
	}
)

type State int

const (
	SourceSelect State = iota
	AtlasTypeSelect
	IdInput
	MiscOptions
	Confirm
	Parsing
)

type Options struct {
	noFile bool
}

type Model struct {
	ready                         bool
	terminalWidth, terminalHeight int

	theme                  Theme
	statePane, optionsPane viewport.Model
	help                   help.Model
	keymap                 KeyMap
	loadingSpinner         spinner.Model

	currentState                    State
	currentOption, currentSubOption int
	options                         Options
	quitting                        bool
	abort                           bool
	err                             error
}

func NewModel() Model {
	return Model{
		theme:  GetTheme(),
		help:   help.New(),
		keymap: DefaultKeybinds(),
	}
}

func (m Model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	cmds = append(cmds, tea.SetWindowTitle("fgo script parser"))
	return tea.Batch(cmds...)
}

func calculateViewportWidths(terminalWidth int) (int, int) {
	paneOne := 17 // TODO: Magic number
	paneTwo := terminalWidth - paneOne
	// paneTwo := math.Floor(float64(terminalWidth) * 0.75)
	return int(paneOne), int(paneTwo)
}

func (m Model) statePaneContent() string {
	var sb strings.Builder
	steps := []string{"Source", "Type", "IDs", "Options", "Confirm"}
	sb.WriteString(renderPaneTitle("Step", m.theme))
	paneWidth, _ := calculateViewportWidths(m.terminalWidth)

	for i, step := range steps {
		prefix := "◌ "
		if State(i) == m.currentState {
			prefix = "◉ "
		}

		if State(i) == m.currentState {
			sb.WriteString(renderSelection(prefix+truncateText(step, paneWidth), m.theme))
			sb.WriteString("\n")
		} else {
			sb.WriteString(renderInactive(prefix+truncateText(step, paneWidth), m.theme))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (m Model) optionsPaneContent() string {
	var sb strings.Builder
	steps := []string{"a", "ab", "IDs", "Options", "Confirm"}
	sb.WriteString(renderPaneTitle("", m.theme))
	_, paneWidth := calculateViewportWidths(m.terminalWidth)

	for i, step := range steps {
		prefix := "◌ "
		if State(i) == m.currentState {
			prefix = "◉ "
		}

		if State(i) == m.currentState {
			sb.WriteString(renderSelection(prefix+truncateText(step, paneWidth), m.theme))
			sb.WriteString("\n")
		} else {
			sb.WriteString(renderInactive(prefix+truncateText(step, paneWidth), m.theme))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		case "down":
			if m.currentState < 4 {
				m.currentState++
			}
		case "up":
			if m.currentState > 0 {
				m.currentState--
			}
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		v1, v2 := calculateViewportWidths(msg.Width)

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.statePane = viewport.New(v1, msg.Height-verticalMarginHeight)
			m.statePane.Style = paneStyle(0, m.theme)
			m.statePane.YPosition = headerHeight
			m.statePane.SetContent(m.statePaneContent())

			m.optionsPane = viewport.New(v2, msg.Height-verticalMarginHeight)
			m.optionsPane.Style = paneStyle(1, m.theme)
			m.optionsPane.YPosition = headerHeight
			m.optionsPane.SetContent(m.optionsPaneContent())
			m.ready = true
		} else {
			m.statePane.Width = v1
			m.statePane.Height = msg.Height - verticalMarginHeight

			m.optionsPane.Width = v2
			m.optionsPane.Height = msg.Height - verticalMarginHeight
		}
	}

	// Handle keyboard events in the viewport
	m.statePane, cmd = m.statePane.Update(msg)
	cmds = append(cmds, cmd)
	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

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

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

func (m Model) headerView() string {
	title := titleStyle.Render("FGO Script Parser")
	line := strings.Repeat("─", max(0, m.terminalWidth-lipgloss.Width(title)))
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
	return footerStyle.Render(m.help.View(m.keymap))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
