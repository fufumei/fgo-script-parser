package main

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
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

type Options struct {
	noFile           bool
	includeWordCount bool
	// Ignore subdirectory split for local files
	// Map known main story chapter names (can work for local too with some regex)
}

type OptionsEnum int

const (
	NoFile OptionsEnum = iota
	IncludeWordCount
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

	currentOption       OptionsEnum
	currentState        State
	selectedSource      Source
	selectedAtlasIdType AtlasIdType
	options             Options
	results             []ParseResult
	quitting            bool
	abort               bool
	err                 error
	notification        clipboardMsg
}

func NewModel() Model {
	body := textarea.New()
	body.ShowLineNumbers = true
	body.Prompt = ""

	return Model{
		theme:          GetDefaultTheme(),
		IdInput:        body,
		loadingSpinner: spinner.New(),
		help:           help.New(),
		keymap:         DefaultKeybinds(),
		currentState:   SourceSelect,
		timer:          stopwatch.NewWithInterval(time.Millisecond),
	}
}
