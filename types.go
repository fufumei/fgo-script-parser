package main

// TUI types
type State int

const (
	SourceSelect State = iota
	AtlasTypeSelect
	IdInput
	MiscOptions
	ConfirmButton
	Parsing
	PickingFile
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

// Parsing types
type Script struct {
	ScriptId string `json:"scriptId"`
	Script   string `json:"script"`
}

type PhaseScript struct {
	Phase   int `json:"phase"`
	Scripts []Script
}

type Quest struct {
	Id           int    `json:"id"`
	Type         string `json:"type"`
	PhaseScripts []PhaseScript
}

type Response struct {
	Name  string `json:"name"`
	Spots []struct {
		Quests []Quest
	}
}

type Count struct {
	lines      int
	characters int
}
