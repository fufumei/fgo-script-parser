package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type SourceOption struct {
	Title       string
	Description string
	value       Source
}

type AtlasIdTypeOption struct {
	Title       string
	Description string
	value       AtlasIdType
}

var (
	sourceOptions = []SourceOption{
		{
			Title:       "Atlas",
			Description: "Parse directly from Atlas IDs",
			value:       atlas,
		},
		{
			Title:       "Local",
			Description: "Parse from local files on your computer",
			value:       local,
		}}

	atlasIdTypeOptions = []AtlasIdTypeOption{
		{
			Title:       "War",
			Description: "Parse every script in a war (story chapter or event).\nEx: 100 for Fuyuki",
			value:       war,
		},
		{
			Title:       "Quest",
			Description: "Parse every script in a quest (war section or interlude etc).\nEx: 1000001 for Fuyuki chapter 1",
			value:       quest,
		},
		{
			Title:       "Script",
			Description: "Parse specific scripts individually.\nEx: 0100000111 for Fuyuki chapter 1 post battle scene",
			value:       script,
		},
	}
)

func main() {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
