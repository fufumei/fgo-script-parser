package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// State is the current state of the application.
type State int

const (
	hoveringSource State = iota
	pickingSource
	editingSourceType
	editingIds
	pickingFile
	hoveringConfirmButton
	parsing
)

type Source int

const (
	atlas Source = iota
	local
)

type Model struct {
	// state represents the current state of the application.
	state State

	// Parse from Atlas or local files
	source      Source
	SourceInput textinput.Model

	// // To represents the recipient's email address.
	// // This can be a comma-separated list of addresses.
	// To textinput.Model
	// // Subject represents the email's subject.
	// Subject textinput.Model
	// // Body represents the email's body.
	// // This can be written in markdown and will be converted to HTML.
	// Body textarea.Model
	// // Attachments represents the email's attachments.
	// // This is a list of file paths which are picked with a filepicker.
	// Attachments list.Model

	// showCc bool
	// Cc     textinput.Model
	// Bcc    textinput.Model

	// // filepicker is used to pick file attachments.
	// filepicker     filepicker.Model
	// loadingSpinner spinner.Model
	help     help.Model
	keymap   KeyMap
	quitting bool
	abort    bool
	// err            error
}

// NewModel returns a new model for the application.
func NewModel() Model {
	sourceInput := textinput.New()
	sourceInput.Prompt = "Source "
	// sourceInput.Placeholder = "me@example.com"
	// sourceInput.PromptStyle = labelStyle.Copy()
	sourceInput.PromptStyle = labelStyle
	sourceInput.TextStyle = textStyle
	sourceInput.Cursor.Style = cursorStyle
	sourceInput.PlaceholderStyle = placeholderStyle
	sourceInput.SetValue("Atlas")

	m := Model{
		state:       hoveringSource,
		source:      atlas,
		SourceInput: sourceInput,
		help:        help.New(),
		keymap:      DefaultKeybinds(),
	}

	m.focusActiveInput()

	return m
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) blurInputs() {
	m.SourceInput.Blur()
	// m.To.Blur()
	// m.Subject.Blur()
	// m.Body.Blur()
	// if m.showCc {
	// 	m.Cc.Blur()
	// 	m.Bcc.Blur()
	// }
	m.SourceInput.PromptStyle = labelStyle
	// m.To.PromptStyle = labelStyle
	// if m.showCc {
	// 	m.Cc.PromptStyle = labelStyle
	// 	m.Cc.TextStyle = textStyle
	// 	m.Bcc.PromptStyle = labelStyle
	// 	m.Bcc.TextStyle = textStyle
	// }
	// m.Subject.PromptStyle = labelStyle
	m.SourceInput.TextStyle = textStyle
	// m.To.TextStyle = textStyle
	// m.Subject.TextStyle = textStyle
	// m.Attachments.Styles.Title = labelStyle
	// m.Attachments.SetDelegate(attachmentDelegate{false})
}

// Update is the update loop for the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NextInput):
			m.blurInputs()
			// Switch to next state
			switch m.state {
			case hoveringSource:
				m.state = hoveringConfirmButton
			}
			m.focusActiveInput()

		case key.Matches(msg, m.keymap.PrevInput):
			m.blurInputs()
			// Switch to prev state
			switch m.state {
			case hoveringConfirmButton:
				m.state = hoveringSource
			}
			m.focusActiveInput()

		case key.Matches(msg, m.keymap.SelectSource):
			// m.state = pickingSource
			if m.source == atlas {
				m.source = local
				m.SourceInput.SetValue("Local")
			} else {
				m.source = atlas
				m.SourceInput.SetValue("Atlas")
			}
			// return m, simple list view?

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
	m.SourceInput, cmd = m.SourceInput.Update(msg)
	cmds = append(cmds, cmd)
	// m.To, cmd = m.To.Update(msg)
	// cmds = append(cmds, cmd)
	// if m.showCc {
	// 	m.Cc, cmd = m.Cc.Update(msg)
	// 	cmds = append(cmds, cmd)
	// 	m.Bcc, cmd = m.Bcc.Update(msg)
	// 	cmds = append(cmds, cmd)
	// }
	// m.Subject, cmd = m.Subject.Update(msg)
	// cmds = append(cmds, cmd)
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

// View displays the application.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// switch m.state {
	// case pickingFile:
	// 	return "\n" + activeLabelStyle.Render("Attachments") + " " + commentStyle.Render(m.filepicker.CurrentDirectory) +
	// 		"\n\n" + m.filepicker.View()
	// case sendingEmail:
	// 	return "\n " + m.loadingSpinner.View() + "Sending email"
	// }

	var s strings.Builder

	s.WriteString(m.SourceInput.View())
	s.WriteString("\n")
	// s.WriteString(m.To.View())
	// s.WriteString("\n")
	// if m.showCc {
	// 	s.WriteString(m.Cc.View())
	// 	s.WriteString("\n")
	// 	s.WriteString(m.Bcc.View())
	// 	s.WriteString("\n")
	// }
	// s.WriteString(m.Subject.View())
	// s.WriteString("\n\n")
	// s.WriteString(m.Body.View())
	// s.WriteString("\n\n")
	// s.WriteString(m.Attachments.View())
	// s.WriteString("\n")
	if m.state == hoveringConfirmButton {
		s.WriteString(sendButtonActiveStyle.Render("Parse"))
	} else if m.state == hoveringConfirmButton && false {
		s.WriteString(sendButtonInactiveStyle.Render("Parse"))
	} else {
		s.WriteString(sendButtonStyle.Render("Parse"))
	}
	s.WriteString("\n\n")
	s.WriteString(m.help.View(m.keymap))

	// if m.err != nil {
	// 	s.WriteString("\n\n")
	// 	s.WriteString(errorStyle.Render(m.err.Error()))
	// }

	return paddedStyle.Render(s.String())
}

func (m *Model) focusActiveInput() {
	switch m.state {
	case hoveringSource:
		m.SourceInput.PromptStyle = activeLabelStyle
		m.SourceInput.TextStyle = activeTextStyle
		m.SourceInput.Focus()
		m.SourceInput.CursorEnd()
		// case editingTo:
		// 	m.To.PromptStyle = activeLabelStyle
		// 	m.To.TextStyle = activeTextStyle
		// 	m.To.Focus()
		// 	m.To.CursorEnd()
		// case editingCc:
		// 	m.Cc.PromptStyle = activeLabelStyle
		// 	m.Cc.TextStyle = activeTextStyle
		// 	m.Cc.Focus()
		// 	m.Cc.CursorEnd()
		// case editingBcc:
		// 	m.Bcc.PromptStyle = activeLabelStyle
		// 	m.Bcc.TextStyle = activeTextStyle
		// 	m.Bcc.Focus()
		// 	m.Bcc.CursorEnd()
		// case editingSubject:
		// 	m.Subject.PromptStyle = activeLabelStyle
		// 	m.Subject.TextStyle = activeTextStyle
		// 	m.Subject.Focus()
		// 	m.Subject.CursorEnd()
		// case editingBody:
		// 	m.Body.Focus()
		// 	m.Body.CursorEnd()
		// case editingAttachments:
		// 	m.Attachments.Styles.Title = activeLabelStyle
		// 	m.Attachments.SetDelegate(attachmentDelegate{true})
	}
}
