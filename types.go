package main

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
