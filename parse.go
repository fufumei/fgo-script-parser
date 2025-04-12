package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
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
	id    string
	name  string
	count Count
}

type parseSuccessMsg []ParseResult

type parseFailureMsg error

type Script struct {
	ScriptId string `json:"scriptId"`
	Script   string `json:"script"`
}

type Quest struct {
	Name         string `json:"name"`
	Id           int    `json:"id"`
	Type         string `json:"type"`
	PhaseScripts []struct {
		Phase   int `json:"phase"`
		Scripts []Script
	}
}

type Response struct {
	Name     string `json:"name"`
	LongName string `json:"longName"`
	Spots    []struct {
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

		switch m.selectedSource {
		case atlas:
			results, err = m.ParseFromAtlas()
		case local:
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

		// TODO: Don't include ID for local parsing
		if m.options.includeWordCount {
			writer.Write([]string{"Id", "Name", "Lines", "Characters, Words"})
			for _, r := range results {
				writer.Write([]string{r.id, r.name, fmt.Sprint(r.count.lines), fmt.Sprint(r.count.characters)})
			}
		} else {
			writer.Write([]string{"Id", "Name", "Lines", "Characters"})
			for _, r := range results {
				writer.Write([]string{r.id, r.name, fmt.Sprint(r.count.lines), fmt.Sprint(r.count.characters), fmt.Sprint(r.count.characters / 2)})
			}
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
		// Skip empty rows
		if strings.Trim(id, " ") == "" {
			continue
		}

		switch m.selectedAtlasIdType {
		case war:
			s, name, err := FetchWarScripts(id)
			if err != nil {
				return nil, err
			}
			for _, script := range s {
				scripts[script.ScriptId] = script
			}
			// The quest list for OC2 does not include the appendix
			if id == "403" {
				s, _, err = FetchQuestScripts("4000327")
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
			result, err := ParseScripts(scr, name)
			if err != nil {
				return nil, err
			}
			result.id = id
			results = append(results, result)
		case quest:
			s, name, err := FetchQuestScripts(id)
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
			result, err := ParseScripts(scr, name)
			if err != nil {
				return nil, err
			}
			result.id = id
			results = append(results, result)
		case script:
			result, err := FetchSingleScript(id)
			if err != nil {
				return nil, err
			}
			result.id = id
			results = append(results, result)
		}
	}

	return results, nil
}

func (m Model) ParseFromLocal() ([]ParseResult, error) {
	var results []ParseResult

	for path := range strings.SplitSeq(m.IdInput.Value(), "\n") {
		// Get rid of empty rows
		if strings.Trim(path, " ") == "" {
			continue
		}
		path = strings.Trim(path, "\"")
		argInfo, err := os.Stat(path)
		if err != nil {
			return nil, parseFailureMsg(fmt.Errorf("could not get file info for %s. %s", path, err))
		}

		// If given path is a file, just open and count it, else traverse the directory
		if argInfo.IsDir() {
			TraverseDirectories(path, &results)
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, parseFailureMsg(fmt.Errorf("can't read file: %s. %s", path, err))
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
		return nil, parseFailureMsg(fmt.Errorf("could not create output file. %s", err))
	}
	return file, nil
}

func ParseScripts(scripts []Script, name string) (ParseResult, error) {
	lines := 0
	characters := 0
	ch := make(chan Count, len(scripts))
	wg := sync.WaitGroup{}
	for _, script := range scripts {
		wg.Add(1)
		go func(script Script) {
			// TODO: Handle error
			r, _ := fetch.Get(script.Script)
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
	}, nil
}

func FetchWarScripts(id string) ([]Script, string, error) {
	var result Response
	response, err := fetch.Get(fmt.Sprintf("https://api.atlasacademy.io/nice/JP/war/%s?lang=en", id))
	if response.StatusCode() == 404 {
		return nil, "", parseFailureMsg(fmt.Errorf("could not get data for war with ID %s. Make sure the ID is correct", id))
	} else if err != nil {
		return nil, "", parseFailureMsg(fmt.Errorf("could not get data for war with ID %s. %s", id, err))
	}
	err = response.UnmarshalJSON(&result)
	if err != nil {
		return nil, "", parseFailureMsg(fmt.Errorf("error unmarshaling JSON: %s", err))
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
	// TODO: Might be needed for the other fetching functions as well
	name := result.Name
	if result.Name == "-" {
		name = result.LongName
	}
	return scripts, name, nil
}

func FetchQuestScripts(id string) ([]Script, string, error) {
	var result Quest
	response, err := fetch.Get(fmt.Sprintf("https://api.atlasacademy.io/nice/JP/quest/%s?lang=en", id))
	if response.StatusCode() == 404 {
		return nil, "", parseFailureMsg(fmt.Errorf("could not get data for quest with ID %s. Make sure the ID is correct", id))
	} else if err != nil {
		return nil, "", parseFailureMsg(fmt.Errorf("could not get data for quest with ID %s. %s", id, err))
	}
	err = response.UnmarshalJSON(&result)
	if err != nil {
		return nil, "", parseFailureMsg(fmt.Errorf("error unmarshaling JSON: %s", err))
	}

	var scripts []Script
	for _, phase := range result.PhaseScripts {
		scripts = append(scripts, phase.Scripts...)
	}

	return scripts, result.Name, nil
}

func FetchSingleScript(id string) (ParseResult, error) {
	response, err := fetch.Get(fmt.Sprintf("https://static.atlasacademy.io/JP/Script/%s/%s.txt", id[0:2], id))
	if response.StatusCode() == 404 {
		return ParseResult{}, parseFailureMsg(fmt.Errorf("error fetching script %s. Make sure the ID is correct", id))
	} else if err != nil {
		return ParseResult{}, parseFailureMsg(fmt.Errorf("error fetching script %s. %s", id, err))
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
		return parseFailureMsg(fmt.Errorf("unable to get entries for directory %s", path))
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
			return parseFailureMsg(fmt.Errorf("can't read file: %s", filepath.Join(path, e.Name())))
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
