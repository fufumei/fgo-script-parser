package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	state  State
	styles Styles

	// Parse from Atlas or local files
	source        Source
	sourceOptions []ListItem

	// Atlas ID type
	atlasIdType        AtlasIdType
	atlasIdTypeOptions []ListItem

	// ID list for atlas parsing
	IdInput textarea.Model

	// filepicker for local parsing.
	filepicker  filepicker.Model
	Attachments list.Model

	// TODO: Change to an "options" struct
	// Print to file or not
	NoFile bool

	loadingSpinner spinner.Model
	help           help.Model
	keymap         KeyMap
	quitting       bool
	abort          bool
	err            error
}

func NewModel() Model {
	styles := NewStyles()

	body := textarea.New()
	body.ShowLineNumbers = false
	body.FocusedStyle.CursorLine = styles.ActiveText
	body.FocusedStyle.Prompt = styles.CurrentLabel
	body.FocusedStyle.Text = styles.ActiveText
	body.BlurredStyle.CursorLine = styles.Text
	body.BlurredStyle.Text = styles.Text
	body.Cursor.Style = styles.Cursor

	attachments := list.New([]list.Item{}, attachmentDelegate{}, 0, 3)
	attachments.DisableQuitKeybindings()
	attachments.SetShowTitle(true)
	attachments.Title = "Files"
	attachments.Styles.Title = styles.CurrentLabel
	attachments.Styles.TitleBar = styles.CurrentLabel
	attachments.SetShowHelp(false)
	attachments.SetShowStatusBar(false)
	attachments.SetStatusBarItemName("script", "scripts")
	attachments.SetShowPagination(false)

	picker := filepicker.New()
	picker.AllowedTypes = []string{".txt"}
	picker.DirAllowed = true

	loadingSpinner := spinner.New()
	loadingSpinner.Style = styles.CurrentLabel
	loadingSpinner.Spinner = spinner.Dot

	m := Model{
		state:  SourceSelect,
		styles: styles,
		source: atlas,
		sourceOptions: []ListItem{
			{
				Title:       "Atlas",
				Description: "Parse directly from Atlas IDs",
			},
			{
				Title:       "Local",
				Description: "Parse from local files on your computer",
			}},
		atlasIdType: war,
		atlasIdTypeOptions: []ListItem{
			{
				Title:       "War",
				Description: "Parse every script in a war (story chapter or event).\nEx: 100 for Fuyuki",
			},
			{
				Title:       "Quest",
				Description: "Parse every script in a quest (war section or interlude etc).\nEx: 1000001 for Fuyuki chapter 1",
			},
			{
				Title:       "Script",
				Description: "Parse specific scripts individually.\nEx: 0100000111 for Fuyuki chapter 1 post battle scene",
			},
		},
		IdInput:        body,
		NoFile:         false,
		Attachments:    attachments,
		filepicker:     picker,
		loadingSpinner: loadingSpinner,
		help:           help.New(),
		keymap:         DefaultKeybinds(),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

type clearErrMsg struct{}

func clearErrAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearErrMsg{}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case parseSuccessMsg:
		// TODO: Display results in a table (bubbles)
		m.quitting = true
		return m, tea.Quit
	case parseFailureMsg:
		m.err = msg
		m.state = ConfirmButton
		return m, clearErrAfter(10 * time.Second)
	case clearErrMsg:
		m.err = nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NextState):
			switch m.state {
			case SourceSelect:
				if m.source == atlas {
					m.state = AtlasTypeSelect
				} else {
					m.state = IdInput
				}
			case AtlasTypeSelect:
				m.state = IdInput
				m.IdInput.Focus()
				m.IdInput.CursorEnd()
				// m.Attachments.SetDelegate(attachmentDelegate{true})
			case IdInput:
				m.IdInput.Blur()
				// m.Attachments.SetDelegate(attachmentDelegate{false})
				m.state = MiscOptions
			case MiscOptions:
				m.state = ConfirmButton
			}

		case key.Matches(msg, m.keymap.PrevState):
			switch m.state {
			case AtlasTypeSelect:
				m.state = SourceSelect
			case IdInput:
				if m.source == atlas {
					m.state = AtlasTypeSelect
				} else {
					m.state = SourceSelect
				}
				m.IdInput.Blur()
				m.Attachments.SetDelegate(attachmentDelegate{false})
			case MiscOptions:
				m.state = IdInput
				m.IdInput.Focus()
				m.IdInput.CursorEnd()
				m.Attachments.SetDelegate(attachmentDelegate{true})
			case ConfirmButton:
				m.state = MiscOptions
			}

		// case key.Matches(msg, m.keymap.Back):
		// 	m.state = IdInput
		// 	m.updateKeymap()
		// 	return m, nil

		// case key.Matches(msg, m.keymap.Attach):
		// 	m.state = PickingFile
		// 	return m, m.filepicker.Init()

		// case key.Matches(msg, m.keymap.Unattach):
		// 	m.Attachments.RemoveItem(m.Attachments.Index())
		// 	m.Attachments.SetHeight(ordered.Max(len(m.Attachments.Items()), 1) + 2)

		case key.Matches(msg, m.keymap.NextOption):
			switch m.state {
			case SourceSelect:
				if m.source == atlas {
					m.source = local
				} else {
					m.source = atlas
				}
			case AtlasTypeSelect:
				if int(m.atlasIdType) < len(m.atlasIdTypeOptions)-1 {
					m.atlasIdType = m.atlasIdType + 1
				}
			}

		case key.Matches(msg, m.keymap.PrevOption):
			switch m.state {
			case SourceSelect:
				if m.source == atlas {
					m.source = local
				} else {
					m.source = atlas
				}
			case AtlasTypeSelect:
				if int(m.atlasIdType) > 0 {
					m.atlasIdType = m.atlasIdType - 1
				}
			}

		case key.Matches(msg, m.keymap.Toggle):
			m.NoFile = !m.NoFile

		case key.Matches(msg, m.keymap.Confirm):
			m.state = Parsing
			m.err = nil
			return m, tea.Batch(
				m.loadingSpinner.Tick,
				m.parseScriptCmd(),
			)

		case key.Matches(msg, m.keymap.Quit):
			m.quitting = true
			m.abort = true
			return m, tea.Quit
		}
	}

	m.updateKeymap()

	var cmds []tea.Cmd
	var cmd tea.Cmd
	cmds = append(cmds, cmd)
	m.IdInput, cmd = m.IdInput.Update(msg)
	cmds = append(cmds, cmd)
	m.filepicker, cmd = m.filepicker.Update(msg)
	cmds = append(cmds, cmd)

	switch m.state {
	// case PickingFile:
	// 	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
	// 		m.Attachments.InsertItem(0, attachment(path))
	// 		m.Attachments.SetHeight(len(m.Attachments.Items()) + 2)
	// 		m.state = IdInput
	// 		m.updateKeymap()
	// 	}
	// case IdInput:
	// 	if m.source == local {
	// 		m.Attachments, cmd = m.Attachments.Update(msg)
	// 		cmds = append(cmds, cmd)
	// 	}
	case Parsing:
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)
	cmds = append(cmds, tea.ClearScrollArea)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// switch m.state {
	// case PickingFile:
	// 	return "\n" + m.styles.CurrentLabel.Render("Attachments") + " " +
	// 		m.styles.CommentStyle.Render(m.filepicker.CurrentDirectory) +
	// 		"\n\n" + m.filepicker.View()
	// }

	var (
		sourceHeader string
		sourceList   []string
		sourceRender string

		atlasTypeHeader string
		atlasTypeList   []string
		atlasTypeRender string

		miscOptionsHeader string
		miscOptionsList   []string
		miscOptionsRender string

		noFileRender string

		confirmButtonRender string
	)
	currentHeader := "> "
	uncheckedBox := "[ ] "
	checkedBox := "[X] "
	sourceHeaderText := "Source "
	atlasTypeHeaderText := "Source Type "
	// idInputHeaderText := "IDs "
	miscOptionsHeaderText := "Options "
	noFileHeaderText := "No file:"
	noFileHeaderDescText := "If checked, print the result directly to the terminal,\notherwise outputs to a csv on the same level as the script."
	confirmButtonText := "Parse"

	// 	 0: current state
	// < 0: previous state
	// > 0: upcoming state
	atlasTypeState := int(AtlasTypeSelect) - int(m.state)
	// idInputState := int(IdInput) - int(m.state)
	miscOptionsState := int(MiscOptions) - int(m.state)
	confirmButtonState := int(ConfirmButton) - int(m.state)

	// ------------------------ //

	if m.state == SourceSelect {
		sourceHeader = m.styles.CurrentLabel.Render(currentHeader + sourceHeaderText)
	} else {
		sourceHeader = m.styles.PreviousLabel.Render(sourceHeaderText)
	}

	for i, o := range m.sourceOptions {
		title := o.Title
		desc := o.Description

		if m.source == Source(i) {
			title = m.styles.SelectedItemTitle.Render(title)
			desc = m.styles.SelectedItemDescription.Render(desc)
		} else {
			title = m.styles.ItemTitle.Render(title)
			desc = m.styles.ItemDescription.Render(desc)
		}
		sourceList = append(sourceList, fmt.Sprintf("%s\n%s\n", title, desc))
	}

	sourceRender =
		lipgloss.JoinVertical(
			lipgloss.Left,
			sourceHeader,
			m.styles.ListBlock.Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					sourceList...,
				)),
		)

	switch {
	case atlasTypeState < 0:
		atlasTypeHeader = m.styles.PreviousLabel.Render(atlasTypeHeaderText)
	case atlasTypeState > 0:
		atlasTypeHeader = m.styles.Disabled.Render(atlasTypeHeaderText)
	case atlasTypeState == 0:
		atlasTypeHeader = m.styles.CurrentLabel.Render(currentHeader + atlasTypeHeaderText)
	}

	for i, o := range m.atlasIdTypeOptions {
		title := o.Title
		desc := o.Description

		switch {
		case atlasTypeState < 0:
			if m.atlasIdType == AtlasIdType(i) {
				title = m.styles.SelectedItemTitle.Render(title)
				desc = m.styles.SelectedItemDescription.Render(desc)
			} else {
				title = m.styles.ItemTitle.Render(title)
				desc = m.styles.ItemDescription.Render(desc)
			}
		case atlasTypeState > 0:
			title = m.styles.DisabledItemTitle.Render(title)
			desc = m.styles.DisabledItemDescription.Render(desc)
		case atlasTypeState == 0:
			if m.atlasIdType == AtlasIdType(i) {
				title = m.styles.SelectedItemTitle.Render(title)
				desc = m.styles.SelectedItemDescription.Render(desc)
			} else {
				title = m.styles.ItemTitle.Render(title)
				desc = m.styles.ItemDescription.Render(desc)
			}

		}

		atlasTypeList = append(atlasTypeList, fmt.Sprintf("%s\n%s\n", title, desc))
	}

	atlasTypeRender =
		lipgloss.JoinVertical(
			lipgloss.Left,
			atlasTypeHeader,
			m.styles.ListBlock.Render(
				lipgloss.JoinVertical(
					lipgloss.Left,
					atlasTypeList...,
				)),
		)

	// TODO: Include description to explain one ID per line,
	// and for local it takes one FULL filepath per line
	// switch {
	// case idInputState < 0:
	// 	idInputHeader = m.styles.PreviousLabel.Render(idInputHeaderText)
	// case idInputState > 0:
	// 	idInputHeader = m.styles.Disabled.Render(idInputHeaderText)
	// case idInputState == 0:
	// 	idInputHeader = m.styles.CurrentLabel.Render(currentHeader + idInputHeaderText)
	// }

	idInput := m.IdInput.View()

	noFileCheckbox := uncheckedBox
	if m.NoFile {
		noFileCheckbox = checkedBox
	}

	switch {
	case miscOptionsState < 0:
		miscOptionsHeader = m.styles.PreviousLabel.Render(miscOptionsHeaderText)
		noFileRender = lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.PreviousLabel.Render(noFileCheckbox+noFileHeaderText),
			m.styles.PreviousLabel.Render(noFileHeaderDescText),
		)
		miscOptionsList = append(miscOptionsList, noFileRender)
	case miscOptionsState > 0:
		miscOptionsHeader = m.styles.Disabled.Render(miscOptionsHeaderText)
		noFileRender = lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.Disabled.Render(noFileCheckbox+noFileHeaderText),
			m.styles.Disabled.Render(noFileHeaderDescText),
		)
		miscOptionsList = append(miscOptionsList, noFileRender)
	case miscOptionsState == 0:
		miscOptionsHeader = m.styles.CurrentLabel.Render(miscOptionsHeaderText)
		noFileRender = lipgloss.JoinVertical(
			lipgloss.Left,
			m.styles.SelectedItemTitle.Render(currentHeader+noFileCheckbox+noFileHeaderText),
			m.styles.SelectedItemDescription.Render(noFileHeaderDescText),
		)
		miscOptionsList = append(miscOptionsList, noFileRender)
	}

	miscOptionsRender =
		lipgloss.JoinVertical(
			lipgloss.Left,
			miscOptionsHeader,
			lipgloss.JoinHorizontal(
				lipgloss.Bottom,
				miscOptionsList...,
			),
		)

	switch {
	case confirmButtonState > 0:
		confirmButtonRender = m.styles.SendButtonStyle.Render(confirmButtonText)
	case confirmButtonState == 0:
		confirmButtonRender = m.styles.SendButtonActiveStyle.Render(confirmButtonText)
	}

	parsingRender := confirmButtonRender
	if m.state == Parsing {
		parsingRender = m.loadingSpinner.View() + "Parsing scripts"
	}

	errRender := ""
	if m.err != nil {
		errRender = m.styles.Error.Render(m.err.Error() + "\n")
	}

	// TODO: Check https://github.com/jahvon/flow
	var states []string
	validStates := []State{SourceSelect, AtlasTypeSelect, IdInput, MiscOptions}
	if m.source == local {
		validStates = slices.DeleteFunc(validStates, func(s State) bool {
			return s == AtlasTypeSelect
		})
	}
	for _, state := range []State{SourceSelect, AtlasTypeSelect, IdInput, MiscOptions} {
		name := ""
		switch state {
		case SourceSelect:
			name = "Source"
		case AtlasTypeSelect:
			name = "Type"
		case IdInput:
			name = "IDs"
		case MiscOptions:
			name = "Options"
		}
		if m.state == state {
			states = append(states, m.styles.ActiveBanner.Render(name))
		} else {
			states = append(states, m.styles.Banner.Render(name))
		}
		// if state != MiscOptions {
		// 	states = append(states, "->")
		// }
	}
	stateFlow := lipgloss.JoinHorizontal(
		lipgloss.Center,
		states...,
	)

	statePaneRender := ""
	switch m.state {
	case SourceSelect:
		statePaneRender = (lipgloss.JoinVertical(
			lipgloss.Left,
			sourceRender,
			m.help.View(m.keymap),
		))
	case AtlasTypeSelect:
		statePaneRender = (lipgloss.JoinVertical(
			lipgloss.Left,
			atlasTypeRender,
			m.help.View(m.keymap),
		))
	case IdInput:
		statePaneRender = (lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinVertical(
				lipgloss.Left,
				m.IdInput.View(),
				idInput,
			),
			// TODO: Display an error directly here if empty when pressing next state
			m.help.View(m.keymap),
		))
	case MiscOptions, ConfirmButton:
		statePaneRender = (lipgloss.JoinVertical(
			lipgloss.Left,
			miscOptionsRender,
			"\n",
			parsingRender,
			errRender,
			m.help.View(m.keymap),
		))
		// TODO: Result state
	}

	return m.styles.Padding.Render(lipgloss.JoinVertical(lipgloss.Left, stateFlow, "\n", statePaneRender))
}
