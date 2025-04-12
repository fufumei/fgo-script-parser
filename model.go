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
	SourceMaxCount int = iota
)

type AtlasIdType int

const (
	war AtlasIdType = iota
	quest
	script
	AtlasIdTypeMaxCount int = iota
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
	OptionsMaxCount int = iota
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
	currentOption       OptionsEnum
	currentState        State
	selectedSource      Source
	selectedAtlasIdType AtlasIdType
	options             Options
	results             []ParseResult
	notification        notificationMsg

	theme                  Theme
	help                   help.Model
	keymap                 KeyMap
	statePane, optionsPane viewport.Model
	IdInput                textarea.Model
	loadingSpinner         spinner.Model
	timer                  stopwatch.Model
	resultsTable           table.Model

	ready                         bool
	terminalWidth, terminalHeight int
	quitting                      bool
	abort                         bool
	err                           error
}

func NewModel() Model {
	body := textarea.New()
	body.ShowLineNumbers = true
	body.Prompt = ""

	return Model{
		theme:          DefaultTheme(),
		IdInput:        body,
		loadingSpinner: spinner.New(),
		help:           help.New(),
		keymap:         DefaultKeybinds(),
		currentState:   SourceSelect,
		timer:          stopwatch.NewWithInterval(time.Millisecond),
	}
}
