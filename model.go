package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type State int

const (
	selectSource State = iota
	selectAtlasIdType
	enteringIds
	pickingFile
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

	// // filepicker is used to pick file attachments.
	// filepicker     filepicker.Model
	// loadingSpinner spinner.Model
	help     help.Model
	keymap   KeyMap
	quitting bool
	abort    bool
	// err            error
}

func NewModel() Model {
	m := Model{
		state:  selectSource,
		styles: NewStyles(),
		source: atlas,
		sourceOptions: []ListItem{
			{Title: "Atlas", Description: "Parse directly from Atlas IDs"},
			{Title: "Local", Description: "Parse from local files on your computer"}},
		atlasIdType: war,
		atlasIdTypeOptions: []ListItem{
			{Title: "War", Description: "Parse every script in a war (story chapter or event).\nEx: 100 for Fuyuki"},
			{Title: "Quest", Description: "Parse every script in a quest (war section or interlude etc).\nEx: 1000001 for Fuyuki chapter 1"},
			{Title: "Script", Description: "Parse specific scripts individually.\nEx: 0100000111 for Fuyuki chapter 1 post battle scene"},
		},
		help:   help.New(),
		keymap: DefaultKeybinds(),
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
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NextInput):
			switch m.state {
			case selectSource:
				m.state = selectAtlasIdType
			case selectAtlasIdType:
				m.state = hoveringConfirmButton
			}

		case key.Matches(msg, m.keymap.PrevInput):
			switch m.state {
			case selectAtlasIdType:
				m.state = selectSource
			case hoveringConfirmButton:
				m.state = selectAtlasIdType
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
			m.state = parsing
			return m, tea.Batch(
			// loading spinner
			// call parsing functions
			// change to result view (split view?)
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
	// m.Body, cmd = m.Body.Update(msg)
	// cmds = append(cmds, cmd)
	// m.filepicker, cmd = m.filepicker.Update(msg)
	// cmds = append(cmds, cmd)

	// switch m.state {
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
	// case sendingEmail:
	// 	m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
	// 	cmds = append(cmds, cmd)
	// }

	m.help, cmd = m.help.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	s := ""

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

	IdTypeListText := ""
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
		IdTypeListText += fmt.Sprintf("%s\n%s\n\n", title, desc)
	}
	s += m.styles.ListBlock.Render(IdTypeListText)
	s += "\n"

	if m.state == hoveringConfirmButton {
		s += sendButtonActiveStyle.Render("Parse")
	} else if m.state == hoveringConfirmButton && false {
		s += sendButtonInactiveStyle.Render("Parse")
	} else {
		s += sendButtonStyle.Render("Parse")
	}

	s += "\n\n"
	s += m.help.View(m.keymap)

	return m.styles.Padding.Render(s)
}
