package main

import (
	"encoding/csv"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type parseSuccessMsg struct{}

func (m Model) parseScriptCmd() tea.Cmd {
	return func() tea.Msg {
		if m.source == atlas {
			m.ParseFromAtlas2()
		} else {
			// ParseFromLocal()
		}
		return parseSuccessMsg{}
	}

}

func (m Model) ParseFromAtlas2() {
	var writer *csv.Writer
	if !m.NoFile {
		file := CreateFile()
		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	writer.Comma = '\t'
	writer.Write([]string{"Name", "Lines", "Characters"})

	scripts := make(map[string]Script)
	var name string
	if m.atlasIdType == war {
		s, n := FetchWarScripts(m.IdInput.Value())
		name = n
		for _, script := range s {
			scripts[script.ScriptId] = script
		}
		// The quest list for OC2 does not include the appendix
		if m.IdInput.Value() == "403" {
			s = FetchQuestScripts("4000327")
			for _, script := range s {
				scripts[script.ScriptId] = script
			}
		}
	} else if m.atlasIdType == quest {
		s := FetchQuestScripts(m.IdInput.Value())
		for _, script := range s {
			scripts[script.ScriptId] = script
		}
		name = m.IdInput.Value()
	} else if m.atlasIdType == script {
		FetchScript(m.IdInput.Value(), writer)
		writer.Flush()
		return
	}

	var scr []Script
	for _, v := range scripts {
		scr = append(scr, v)
	}
	ParseScripts(scr, name, writer)
	writer.Flush()
}
