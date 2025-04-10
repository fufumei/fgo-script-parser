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
	Quit          key.Binding
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
		k.Quit,
	}
}

// FullHelp returns the key bindings for the full help screen.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.NextState,
			k.Confirm,
			k.Quit,
		},
	}
}

func (m *Model) updateKeymap() {
	stateHasOptions := m.currentState == SourceSelect || m.currentState == AtlasTypeSelect
	hasNextstate := true

	if m.currentState == Results {
		hasNextstate = false
	} else if m.currentState == Confirm {
		if len(m.results) > 0 {
			hasNextstate = true
		} else {
			hasNextstate = false
		}
	} else {
		hasNextstate = true
	}

	m.keymap.NextState.SetEnabled(hasNextstate)
	m.keymap.PrevState.SetEnabled(m.currentState != SourceSelect)
	m.keymap.NextOption.SetEnabled(stateHasOptions)
	m.keymap.PrevOption.SetEnabled(stateHasOptions)
	m.keymap.Toggle.SetEnabled(m.currentState == MiscOptions)
	m.keymap.Confirm.SetEnabled(m.currentState == Confirm)
}
