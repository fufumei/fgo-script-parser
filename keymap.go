package main

import "github.com/charmbracelet/bubbles/key"

// KeyMap represents the key bindings for the application.
type KeyMap struct {
	NextInput    key.Binding
	PrevInput    key.Binding
	SelectSource key.Binding
	Confirm      key.Binding
	// Send      key.Binding
	// Attach    key.Binding
	// Unattach  key.Binding
	// Back      key.Binding
	Quit key.Binding
}

// DefaultKeybinds returns the default key bindings for the application.
func DefaultKeybinds() KeyMap {
	return KeyMap{
		NextInput: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		PrevInput: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev"),
		),
		SelectSource: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "toggle source"),
			key.WithDisabled(),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "parse scripts"),
			key.WithDisabled(),
		),
		// Attach: key.NewBinding(
		// 	key.WithKeys("enter"),
		// 	key.WithHelp("enter", "attach file"),
		// 	key.WithDisabled(),
		// ),
		// Unattach: key.NewBinding(
		// 	key.WithKeys("x"),
		// 	key.WithHelp("x", "remove"),
		// 	key.WithDisabled(),
		// ),
		// Back: key.NewBinding(
		// 	key.WithKeys("esc"),
		// 	key.WithHelp("esc", "back"),
		// 	key.WithDisabled(),
		// ),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "quit"),
		),
	}
}

// ShortHelp returns the key bindings for the short help screen.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NextInput,
		k.Quit,
		k.SelectSource,
		k.Confirm,
		k.Quit,
	}
}

// FullHelp returns the key bindings for the full help screen.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextInput, k.SelectSource, k.Confirm, k.Quit},
	}
}

func (m *Model) updateKeymap() {
	m.keymap.SelectSource.SetEnabled(m.state == hoveringSource)
	m.keymap.Confirm.SetEnabled(m.state == hoveringConfirmButton)
	// m.keymap.Unattach.SetEnabled(m.state == editingAttachments && len(m.Attachments.Items()) > 0)
	// m.keymap.Back.SetEnabled(m.state == pickingFile)

	// m.filepicker.KeyMap.Up.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.Down.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.Back.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.Select.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.Open.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.PageUp.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.PageDown.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.GoToTop.SetEnabled(m.state == pickingFile)
	// m.filepicker.KeyMap.GoToLast.SetEnabled(m.state == pickingFile)
}
