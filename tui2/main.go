package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
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
	Results
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
	IdInput                textarea.Model
	help                   help.Model
	keymap                 KeyMap
	loadingSpinner         spinner.Model
	timer                  stopwatch.Model
	resultsTable           table.Model

	currentOption, currentSubOption int
	currentState                    State
	selectedSource                  Source
	selectedAtlasIdType             AtlasIdType
	options                         Options
	results                         []ParseResult
	quitting                        bool
	abort                           bool
	err                             error
}

func NewModel() Model {
	body := textarea.New()
	body.ShowLineNumbers = true
	body.Prompt = ""

	loadingSpinner := spinner.New()

	return Model{
		theme:          GetDefaultTheme(),
		IdInput:        body,
		loadingSpinner: loadingSpinner,
		help:           help.New(),
		keymap:         DefaultKeybinds(),
		currentState:   SourceSelect,
		timer:          stopwatch.NewWithInterval(time.Millisecond),
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
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Foreground(m.theme.TertiaryColor).Render("IDs"))
	sb.WriteString("\n")
	sb.WriteString(renderDefault("Enter the IDs (if Atlas) or filepaths (if local) to parse from.\nOne ID/filepath per line.\nNote that filepaths cannot be relative.", m.theme))
	sb.WriteString("\n\n")

	sb.WriteString(m.IdInput.View())

	return sb.String()
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

// TODO: Prevent changing state while entering IDs, have to press esc to blur and enter to focus again?
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case parseSuccessMsg:
		// TODO: Add hotkeys to copy single or all values
		m.results = msg

		// TODO: Calculate heights based on available height
		columns := []table.Column{
			{Title: "Name", Width: 25},
			{Title: "Lines", Width: 25},
			{Title: "Characters", Width: 25},
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
				// TODO: Doesn't work
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

		case key.Matches(msg, m.keymap.Confirm):
			m.currentState = Parsing
			m.err = nil
			return m, tea.Batch(
				m.loadingSpinner.Tick,
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

			m.IdInput.SetHeight(msg.Height - verticalMarginHeight)
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

type clearErrMsg struct{}

func clearErrAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearErrMsg{}
	})
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

	errRender := ""
	if m.err != nil {
		errRender = renderError(m.err.Error(), m.theme)
		return footerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.help.View(m.keymap),
				errRender,
			),
		)
	}

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
