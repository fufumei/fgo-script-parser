package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-zoox/fetch"
	"golang.org/x/exp/slices"
)

type ParseResult struct {
	name  string
	count Count
}

type parseSuccessMsg []ParseResult

type parseFailureMsg error

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

func (m Model) parseScriptCmd() tea.Cmd {
	return func() tea.Msg {
		var results []ParseResult
		var err error
		if strings.TrimSpace(m.IdInput.Value()) == "" {
			return parseFailureMsg(errors.New("IDs cannot be empty"))
		}

		if m.selectedSource == atlas {
			results, err = m.ParseFromAtlas()
		} else {
			results, err = m.ParseFromLocal()
		}

		if err != nil {
			return parseFailureMsg(err)
		}

		var writer *csv.Writer
		if !m.options.noFile {
			file, err := CreateFile()
			if err != nil {
				return parseFailureMsg(err)
			}
			writer = csv.NewWriter(file)
		} else {
			writer = csv.NewWriter(os.Stdout)
		}
		writer.Comma = '\t'
		writer.Write([]string{"Name", "Lines", "Characters"})

		// TODO: Sort
		for _, r := range results {
			writer.Write([]string{r.name, fmt.Sprint(r.count.lines), fmt.Sprint(r.count.characters)})
		}

		if !m.options.noFile {
			writer.Flush()
		}
		return parseSuccessMsg(results)
	}
}

func (m Model) ParseFromAtlas() ([]ParseResult, error) {
	var results []ParseResult

	scripts := make(map[string]Script)
	for id := range strings.SplitSeq(m.IdInput.Value(), "\n") {
		if m.selectedAtlasIdType == war {
			s, n, err := FetchWarScripts(id)
			if err != nil {
				return nil, err
			}
			for _, script := range s {
				scripts[script.ScriptId] = script
			}
			// The quest list for OC2 does not include the appendix
			if id == "403" {
				s, err = FetchQuestScripts("4000327")
				if err != nil {
					return nil, err
				}
				for _, script := range s {
					scripts[script.ScriptId] = script
				}
			}

			var scr []Script
			for _, v := range scripts {
				scr = append(scr, v)
			}
			result := ParseScripts(scr, n)
			results = append(results, result)
		} else if m.selectedAtlasIdType == quest {
			s, err := FetchQuestScripts(id)
			if err != nil {
				return nil, err
			}
			for _, script := range s {
				scripts[script.ScriptId] = script
			}

			var scr []Script
			for _, v := range scripts {
				scr = append(scr, v)
			}
			result := ParseScripts(scr, id)
			results = append(results, result)
		} else if m.selectedAtlasIdType == script {
			result, err := FetchSingleScript(id)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (m Model) ParseFromLocal() ([]ParseResult, error) {
	var results []ParseResult

	for _, path := range strings.Split(m.IdInput.Value(), "\n") {
		path = strings.Trim(path, "\"")
		argInfo, err := os.Stat(path)
		if err != nil {
			return nil, parseFailureMsg(errors.New(fmt.Sprintf("Could not get file info for %s. %s", path, err)))
		}

		// If given path is a file, just open and count it, else traverse the directory
		if argInfo.IsDir() {
			TraverseDirectories(path, &results)
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, parseFailureMsg(errors.New(fmt.Sprintf("Can't read file: %s. %s", path, err)))
			}
			count := CleanAndCountScript(string(data))
			results = append(results, ParseResult{
				name:  strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
				count: count,
			})
		}
	}

	return results, nil
}

func CreateFile() (*os.File, error) {
	file, err := os.Create("script-length.csv")
	if err != nil {
		return nil, parseFailureMsg(errors.New(fmt.Sprintf("Could not create output file. %s", err)))
	}
	return file, nil
}

func ParseScripts(scripts []Script, name string) ParseResult {
	lines := 0
	characters := 0
	ch := make(chan Count, len(scripts))
	wg := sync.WaitGroup{}
	for _, script := range scripts {
		wg.Add(1)

		go func(script Script) {
			r, err := fetch.Get(script.Script)
			if err != nil {
				// TODO: Should this stop the whole thing?
				log.Fatalf("Error fetching script " + script.ScriptId)
				return
				// return nil, parseFailureMsg(errors.New(fmt.Sprintf("Error fetching script %s. %s", script.ScriptId, err)))
			}
			count := CleanAndCountScript(r.String())
			ch <- count
			wg.Done()
		}(script)
	}

	wg.Wait()
	close(ch)

	for c := range ch {
		lines += c.lines
		characters += c.characters
	}

	return ParseResult{
		name: name,
		count: Count{
			lines:      lines,
			characters: characters,
		},
	}
}

func FetchWarScripts(id string) ([]Script, string, error) {
	var result Response
	response, err := fetch.Get(fmt.Sprintf("https://api.atlasacademy.io/nice/JP/war/%s?lang=en", id))
	if response.StatusCode() == 404 {
		return nil, "", parseFailureMsg(errors.New(fmt.Sprintf("Could not get data for war with ID %s. Make sure the ID is correct.", id)))
	} else if err != nil {
		return nil, "", parseFailureMsg(errors.New(fmt.Sprintf("Could not get data for war with ID %s. %s", id, err)))
	}
	err = response.UnmarshalJSON(&result)
	if err != nil {
		return nil, "", parseFailureMsg(errors.New(fmt.Sprintf("Error unmarshaling JSON: %s", err)))
	}

	var scripts []Script
	for _, spot := range result.Spots {
		for _, quest := range spot.Quests {
			// This works for both main story and event quests
			if quest.Type == "main" {
				for _, phase := range quest.PhaseScripts {
					scripts = append(scripts, phase.Scripts...)
				}
			}
		}
	}
	return scripts, result.Name, nil
}

func FetchQuestScripts(id string) ([]Script, error) {
	var result Quest
	response, err := fetch.Get(fmt.Sprintf("https://api.atlasacademy.io/nice/JP/quest/%s?lang=en", id))
	if response.StatusCode() == 404 {
		return nil, parseFailureMsg(errors.New(fmt.Sprintf("Could not get data for quest with ID %s. Make sure the ID is correct.", id)))
	} else if err != nil {
		return nil, parseFailureMsg(errors.New(fmt.Sprintf("Could not get data for quest with ID %s. %s", id, err)))
	}
	err = response.UnmarshalJSON(&result)
	if err != nil {
		return nil, parseFailureMsg(errors.New(fmt.Sprintf("Error unmarshaling JSON: %s", err)))
	}

	var scripts []Script
	for _, phase := range result.PhaseScripts {
		scripts = append(scripts, phase.Scripts...)
	}

	return scripts, nil
}

func FetchSingleScript(id string) (ParseResult, error) {
	response, err := fetch.Get(fmt.Sprintf("https://static.atlasacademy.io/JP/Script/%s/%s.txt", id[0:2], id))
	if response.StatusCode() == 404 {
		return ParseResult{}, parseFailureMsg(errors.New(fmt.Sprintf("Error fetching script %s. Make sure the ID is correct.", id)))
	} else if err != nil {
		return ParseResult{}, parseFailureMsg(errors.New(fmt.Sprintf("Error fetching script %s. %s", id, err)))
	}
	count := CleanAndCountScript(response.String())

	return ParseResult{
		name:  id,
		count: count,
	}, nil
}

func TraverseDirectories(path string, results *[]ParseResult) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return parseFailureMsg(errors.New(fmt.Sprintf("Unable to get entries for directory %s", path)))
	}

	hasDirectory := slices.ContainsFunc(entries, func(e fs.DirEntry) bool {
		return e.IsDir()
	})

	// If directories exist, call this recursively for each one until we hit the lowest level
	if hasDirectory {
		for _, e := range entries {
			err = TraverseDirectories(filepath.Join(path, e.Name()), results)
			if err != nil {
				break
			}
		}
		return err
	}

	lines := 0
	characters := 0
	// This could be done with goroutines but it's pretty fast already
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(path, e.Name()))
		if err != nil {
			return parseFailureMsg(errors.New(fmt.Sprintf("Can't read file: %s", filepath.Join(path, e.Name()))))
		}
		count := CleanAndCountScript(string(data))
		lines += count.lines
		characters += count.characters
	}
	*results = append(*results, ParseResult{
		name: filepath.Base(path),
		count: Count{
			lines:      lines,
			characters: characters,
		},
	})

	return nil
}

func CleanAndCountScript(data string) Count {
	r, _ := regexp.Compile(`(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]|(？.+?：.+)`)
	matches := r.FindAllString(data, -1)
	lines := len(matches)
	characters := 0
	r, _ = regexp.Compile(`(\[[^#&]+?\]|[\[\]#&:]|？.+?：|^＠.+|\n)`)
	for _, m := range matches {
		characters += len([]rune(r.ReplaceAllString(m, "")))
	}

	return Count{
		lines:      lines,
		characters: characters,
	}
}
