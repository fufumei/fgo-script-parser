package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	selectSource State = iota
	selectAtlasIdType
	enteringIds
	pickingFile
	selectNoFile
	hoveringConfirmButton
	parsing
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

type ListItem struct {
	Title       string
	Description string
}

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
	IdInput textinput.Model

	// filepicker for local parsing.
	// filepicker     filepicker.Model

	// Print to file or not
	NoFile bool

	loadingSpinner spinner.Model
	help           help.Model
	keymap         KeyMap
	quitting       bool
	abort          bool
}

func NewModel() Model {
	styles := NewStyles()

	body := textinput.New()
	// body.ShowLineNumbers = false
	// body.FocusedStyle.CursorLine = styles.ActiveText
	// body.FocusedStyle.Prompt = styles.ActiveLabel
	// body.FocusedStyle.Text = styles.ActiveText
	// body.BlurredStyle.CursorLine = styles.Text
	// body.BlurredStyle.Text = styles.Text
	// body.Cursor.Style = styles.Cursor
	body.Prompt = "ID: "
	body.Placeholder = "_ "
	body.PromptStyle = styles.Disabled
	body.Cursor.Style = styles.Cursor
	body.TextStyle = styles.Text

	loadingSpinner := spinner.New()
	loadingSpinner.Style = styles.ActiveLabel
	loadingSpinner.Spinner = spinner.Dot

	m := Model{
		state:  selectSource,
		styles: styles,
		source: atlas,
		sourceOptions: []ListItem{
			{Title: "Atlas", Description: "Parse directly from Atlas IDs"},
			{Title: "Local", Description: "Parse from local files on your computer\n(NOTE: Not implemented currently)"}},
		atlasIdType: war,
		atlasIdTypeOptions: []ListItem{
			{Title: "War", Description: "Parse every script in a war (story chapter or event).\nEx: 100 for Fuyuki"},
			{Title: "Quest", Description: "Parse every script in a quest (war section or interlude etc).\nEx: 1000001 for Fuyuki chapter 1"},
			{Title: "Script", Description: "Parse specific scripts individually.\nEx: 0100000111 for Fuyuki chapter 1 post battle scene"},
		},
		IdInput:        body,
		NoFile:         false,
		loadingSpinner: loadingSpinner,
		help:           help.New(),
		keymap:         DefaultKeybinds(),
	}

	return m
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update is the update loop for the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case parseSuccessMsg:
		{
			m.quitting = true
			return m, tea.Quit
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NextInput):
			switch m.state {
			case selectSource:
				m.state = selectAtlasIdType
			case selectAtlasIdType:
				m.state = enteringIds

				m.IdInput.PromptStyle = m.styles.ActiveLabel
				m.IdInput.TextStyle = m.styles.ActiveText
				m.IdInput.Focus()
				m.IdInput.CursorEnd()
			case enteringIds:
				m.IdInput.Blur()
				m.state = selectNoFile
			case selectNoFile:
				m.state = hoveringConfirmButton
			}

		case key.Matches(msg, m.keymap.PrevInput):
			switch m.state {
			case selectAtlasIdType:
				m.state = selectSource
			case enteringIds:
				m.IdInput.Blur()
				m.state = selectAtlasIdType
				m.IdInput.PromptStyle = m.styles.Disabled
				m.IdInput.TextStyle = m.styles.Text
			case selectNoFile:
				m.state = enteringIds
				m.IdInput.PromptStyle = m.styles.ActiveLabel
				m.IdInput.TextStyle = m.styles.ActiveText
				m.IdInput.Focus()
				m.IdInput.CursorEnd()
			case hoveringConfirmButton:
				m.state = selectNoFile
			}

		case key.Matches(msg, m.keymap.NextOption):
			switch m.state {
			case selectSource:
				if m.source == atlas {
					m.source = local
				} else {
					m.source = atlas
				}
			case selectAtlasIdType:
				if int(m.atlasIdType) < len(m.atlasIdTypeOptions)-1 {
					m.atlasIdType = m.atlasIdType + 1
				}
			}

		case key.Matches(msg, m.keymap.PrevOption):
			switch m.state {
			case selectSource:
				if m.source == atlas {
					m.source = local
				} else {
					m.source = atlas
				}
			case selectAtlasIdType:
				if int(m.atlasIdType) > 0 {
					m.atlasIdType = m.atlasIdType - 1
				}
			}

		case key.Matches(msg, m.keymap.Confirm):
			switch m.state {
			case selectNoFile:
				m.NoFile = !m.NoFile
			case hoveringConfirmButton:
				m.state = parsing
				return m, tea.Batch(
					m.loadingSpinner.Tick,
					m.parseScriptCmd(),
				// call parsing functions
				// change to result view (split view?)
				)
			}

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
	// m.filepicker, cmd = m.filepicker.Update(msg)
	// cmds = append(cmds, cmd)

	switch m.state {
	// case pickingFile:
	// 	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
	// 		m.Attachments.InsertItem(0, attachment(path))
	// 		m.Attachments.SetHeight(len(m.Attachments.Items()) + 2)
	// 		m.state = editingAttachments
	// 		m.updateKeymap()
	// 	}
	// case editingAttachments:
	// 	m.Attachments, cmd = m.Attachments.Update(msg)
	// 	cmds = append(cmds, cmd)
	case parsing:
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	s := ""

	// TODO: Add a caret (>) to whichever state it's currently on

	// Source selection
	s += m.styles.ActiveLabel.Render("Source: ")
	s += "\n"

	sourceListText := ""
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
		sourceListText += fmt.Sprintf("%s\n%s\n\n", title, desc)
	}
	s += m.styles.ListBlock.Render(sourceListText)
	s += "\n"

	// Atlas ID type selection
	atlasIdTypeDisabled := int(m.state) < int(selectAtlasIdType)

	if atlasIdTypeDisabled {
		s += m.styles.Disabled.Render("ID type: ")
	} else {
		s += m.styles.ActiveLabel.Render("ID type: ")
	}
	s += "\n"

	idTypeListText := ""
	for i, o := range m.atlasIdTypeOptions {
		title := o.Title
		desc := o.Description

		if atlasIdTypeDisabled {
			title = m.styles.DisabledItemTitle.Render(title)
			desc = m.styles.DisabledItemDescription.Render(desc)
		} else if m.atlasIdType == AtlasIdType(i) {
			title = m.styles.SelectedItemTitle.Render(title)
			desc = m.styles.SelectedItemDescription.Render(desc)
		} else {
			title = m.styles.ItemTitle.Render(title)
			desc = m.styles.ItemDescription.Render(desc)
		}
		idTypeListText += fmt.Sprintf("%s\n%s\n\n", title, desc)
	}
	s += m.styles.ListBlock.Render(idTypeListText)
	s += "\n"

	// TODO: Weird padding on first line
	// if int(m.state) < int(enteringIds) {
	// 	s += m.styles.Disabled.Render("IDs:\n")
	// } else {
	// 	s += m.styles.ActiveLabel.Render("IDs:\n")
	// }
	s += m.IdInput.View()
	s += "\n\n"

	// No file checkbox
	checkboxLabel := "No file:"
	checkboxDesc := "If checked, print the result directly to the terminal,\notherwise outputs to a csv on the same level as the script.\n(NOTE: This option does nothing currently.)"

	noFileDisabled := int(m.state) < int(selectNoFile)
	noFileSelected := "[ ]"
	if m.NoFile {
		noFileSelected = "[X]"
	}

	if noFileDisabled {
		checkboxLabel = m.styles.DisabledItemTitle.Render(checkboxLabel)
		checkboxDesc = m.styles.DisabledItemDescription.Render(checkboxDesc)
		noFileSelected = m.styles.DisabledItemTitle.Render(noFileSelected)
	} else {
		checkboxLabel = m.styles.ItemTitle.Render(checkboxLabel)
		checkboxDesc = m.styles.ItemDescription.Render(checkboxDesc)
		noFileSelected = m.styles.ItemTitle.Render(noFileSelected)
	}
	s += fmt.Sprintf("%s %s\n%s", checkboxLabel, noFileSelected, checkboxDesc)
	s += "\n\n"

	if m.state == hoveringConfirmButton {
		s += sendButtonActiveStyle.Render("Parse")
	} else if m.state == hoveringConfirmButton && false {
		s += sendButtonInactiveStyle.Render("Parse")
	} else {
		s += sendButtonStyle.Render("Parse")
	}
	s += "\n\n"

	if m.state == parsing {
		s += "\n " + m.loadingSpinner.View() + "Parsing script"
	}
	s += "\n\n"

	s += m.help.View(m.keymap)

	return m.styles.Padding.Render(s)
}
