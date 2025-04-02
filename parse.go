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

func (m Model) parseScriptCmd() tea.Cmd {
	return func() tea.Msg {
		var results []ParseResult
		if strings.TrimSpace(m.IdInput.Value()) == "" {
			return parseFailureMsg(errors.New("IDs cannot be empty"))
		}

		if m.source == atlas {
			results = m.ParseFromAtlas()
		} else {
			results = m.ParseFromLocal()
		}

		var writer *csv.Writer
		if !m.NoFile {
			file := CreateFile()
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

		writer.Flush()
		return parseSuccessMsg(results)
	}
}

func (m Model) ParseFromAtlas() []ParseResult {
	var results []ParseResult

	scripts := make(map[string]Script)
	for _, id := range strings.Split(m.IdInput.Value(), "\n") {
		if m.atlasIdType == war {
			s, n := FetchWarScripts(id)
			for _, script := range s {
				scripts[script.ScriptId] = script
			}
			// The quest list for OC2 does not include the appendix
			if id == "403" {
				s = FetchQuestScripts("4000327")
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
		} else if m.atlasIdType == quest {
			s := FetchQuestScripts(id)
			for _, script := range s {
				scripts[script.ScriptId] = script
			}

			var scr []Script
			for _, v := range scripts {
				scr = append(scr, v)
			}
			result := ParseScripts(scr, id)
			results = append(results, result)
		} else if m.atlasIdType == script {
			result := FetchSingleScript(id)
			results = append(results, result)
		}
	}

	return results
}

func (m Model) ParseFromLocal() []ParseResult {
	var results []ParseResult

	for _, path := range strings.Split(m.IdInput.Value(), "\n\r") {
		path = strings.Trim(path, "\"")
		argInfo, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Could not get file info for %s. %s", path, err)
		}

		// If given path is a file, just open and count it, else traverse the directory
		if argInfo.IsDir() {
			TraverseDirectories(path, &results)
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				// TODO: Return failure message instead
				log.Fatalf("Can't read file: %s", path)
			}
			count := CleanAndCountScript(string(data))
			results = append(results, ParseResult{
				name:  strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
				count: count,
			})
		}
	}

	return results
}

func CreateFile() *os.File {
	file, err := os.Create("script-length.csv")
	if err != nil {
		log.Fatalf("Could not create output file. %s", err)
	}
	return file
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
				log.Fatalf("Error fetching script " + script.ScriptId)
				return
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

func FetchWarScripts(id string) ([]Script, string) {
	var result Response
	response, err := fetch.Get(fmt.Sprintf("https://api.atlasacademy.io/nice/JP/war/%s?lang=en", id))
	if err != nil {
		log.Fatalf("Could not get data for war with ID %s", id)
	}
	err = response.UnmarshalJSON(&result)
	if err != nil {
		log.Fatal("Error unmarshaling JSON:", err)
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
	return scripts, result.Name
}

func FetchQuestScripts(id string) []Script {
	var result Quest
	response, err := fetch.Get(fmt.Sprintf("https://api.atlasacademy.io/nice/JP/quest/%s?lang=en", id))
	if err != nil {
		log.Fatalf("Could not get data for quest with ID %s", id)
	}
	err = response.UnmarshalJSON(&result)
	if err != nil {
		log.Fatal("Error unmarshaling JSON:", err)
	}

	var scripts []Script
	for _, phase := range result.PhaseScripts {
		scripts = append(scripts, phase.Scripts...)
	}

	return scripts
}

func FetchSingleScript(id string) ParseResult {
	r, err := fetch.Get(fmt.Sprintf("https://static.atlasacademy.io/JP/Script/%s/%s.txt", id[0:2], id))
	if err != nil {
		log.Fatalf("Error fetching script " + id)
		return ParseResult{} // TODO: Actually return error here
	}
	count := CleanAndCountScript(r.String())

	return ParseResult{
		name:  id,
		count: count,
	}
}

func TraverseDirectories(path string, results *[]ParseResult) {
	entries, err := os.ReadDir(path)
	// TODO: Return failure message
	if err != nil {
		fmt.Println("Unable to get entries for directory ", path)
	}

	hasDirectory := slices.ContainsFunc(entries, func(e fs.DirEntry) bool {
		return e.IsDir()
	})

	// If directories exist, call this recursively for each one until we hit the lowest level
	if hasDirectory {
		for _, e := range entries {
			TraverseDirectories(filepath.Join(path, e.Name()), results)
		}
		return
	}

	lines := 0
	characters := 0
	// This could be done with goroutines but it's pretty fast already
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(path, e.Name()))
		if err != nil {
			// TODO: Return failure message
			log.Fatalf("Can't read file: %s", filepath.Join(path, e.Name()))
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
	// writer.Write([]string{filepath.Base(path), fmt.Sprint(lines), fmt.Sprint(characters)})
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
