package main

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	NextState     key.Binding
	PrevState     key.Binding
	NextOption    key.Binding
	PrevOption    key.Binding
	NextSubOption key.Binding
	PrevSubOption key.Binding
	Confirm       key.Binding
	Toggle        key.Binding
	// Attach        key.Binding
	// Unattach      key.Binding
	// Back          key.Binding
	Quit key.Binding
}

func DefaultKeybinds() KeyMap {
	return KeyMap{
		NextState: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next step"),
		),
		PrevState: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev step"),
		),
		NextOption: key.NewBinding(
			key.WithKeys("down"),
		),
		PrevOption: key.NewBinding(
			key.WithKeys("up"),
		),
		NextSubOption: key.NewBinding(
			key.WithKeys("right"),
		),
		PrevSubOption: key.NewBinding(
			key.WithKeys("left"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "toggle"),
			key.WithDisabled(),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
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
		k.NextState,
		k.PrevState,
		k.Toggle,
		k.Confirm,
		// k.Attach, k.Unattach,
		k.Quit,
	}
}

// FullHelp returns the key bindings for the full help screen.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.NextState,
			k.Confirm,
			// k.Attach,
			// k.Unattach,
			k.Quit,
		},
	}
}

// func (m *Model) updateKeymap() {
// 	m.keymap.NextState.SetEnabled(m.state != ConfirmButton)
// 	m.keymap.PrevState.SetEnabled(m.state != SourceSelect)
// 	m.keymap.NextOption.SetEnabled(m.state == SourceSelect || m.state == AtlasTypeSelect)
// 	m.keymap.PrevOption.SetEnabled(m.state == SourceSelect || m.state == AtlasTypeSelect)
// 	m.keymap.Confirm.SetEnabled(m.state == ConfirmButton)
// 	m.keymap.Toggle.SetEnabled(m.state == MiscOptions)
// 	m.keymap.Back.SetEnabled(m.state == PickingFile)
// 	m.keymap.Attach.SetEnabled(m.state == IdInput && m.source == local)
// 	m.keymap.Unattach.SetEnabled(m.state == IdInput && m.source == local && len(m.Attachments.Items()) > 0)

// 	m.filepicker.KeyMap.Up.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.Down.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.Back.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.Select.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.Open.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.PageUp.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.PageDown.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.GoToTop.SetEnabled(m.state == PickingFile && m.source == local)
// 	m.filepicker.KeyMap.GoToLast.SetEnabled(m.state == PickingFile && m.source == local)
// }
