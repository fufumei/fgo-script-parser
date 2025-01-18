package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/go-zoox/fetch"
	"golang.org/x/exp/slices"
)

// fgo-script-parser atlas [--no-file] [--war <ID>] [--quest <ID>] [<script ID]
// fgo-script-parser <path> [--no-file] [--ignore-splits]

var CLI struct {
	NoFile bool `name:"no-file" default:"true" help:"If true, print the result directly to the terminal, otherwise outputs to a csv on the same level as the script."`

	Atlas struct {
		War    string `required short:"w" xor:"type" help:"A war ID to query against Atlas."`
		Quest  string `required short:"q" xor:"type" help:"A quest ID to query against Atlas."`
		Script string `required short:"s" xor:"type" help:"A script ID to query against Atlas."`
	} `cmd:"" help:"Fetches scripts from Atlas API to parse."`
	Local struct {
		IgnoreSplits bool `name:"ignore-splits" default:"false" hidden:""`

		Path string `arg:"" name:"path" type:"path"`
	} `cmd:"" help:"Parses locally stored scripts."`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "atlas":
		ParseFromAtlas()
	case "local <path>":
		ParseFromLocal()
	default:
		panic(ctx.Command())
	}
}

func ParseFromAtlas() {
	var writer *csv.Writer
	if !CLI.NoFile {
		file := CreateFile()
		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	writer.Comma = '\t'
	writer.Write([]string{"Name", "Lines", "Characters"})

	var scripts []Script
	var name string
	if CLI.Atlas.War != "" {
		s, n := FetchWarScripts(CLI.Atlas.War)
		name = n
		scripts = append(scripts, s...)
		// The quest list for OC2 does not include the appendix
		if CLI.Atlas.War == "403" {
			scripts = append(scripts, FetchQuestScripts("4000327")...)
		}
	} else if CLI.Atlas.Quest != "" {
		scripts = append(scripts, FetchQuestScripts(CLI.Atlas.Quest)...)
		name = CLI.Atlas.Quest
	} else if CLI.Atlas.Script != "" {
		FetchScript(CLI.Atlas.Script, writer)
		writer.Flush()
		return
	}
	ParseScripts(scripts, name, writer)
	writer.Flush()
}

func ParseFromLocal() {
	var writer *csv.Writer
	if !CLI.NoFile {
		file := CreateFile()
		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	writer.Comma = '\t'
	writer.Write([]string{"Name", "Lines", "Characters"})

	argInfo, err := os.Stat(CLI.Local.Path)
	if err != nil {
		log.Fatalf("Could not get file info for %s.", CLI.Local.Path)
	}

	// If given path is a file, just open and count it, else traverse the directory
	if argInfo.IsDir() {
		TraverseDirectories(CLI.Local.Path, writer)
		writer.Flush()
	} else {
		data, err := os.ReadFile(CLI.Local.Path)
		if err != nil {
			log.Fatalf("Can't read file: %s", CLI.Local.Path)
		}
		lines, characters := CleanAndCountScript(string(data))
		fmt.Printf("%s\t%d\t%d", strings.TrimSuffix(filepath.Base(CLI.Local.Path), filepath.Ext(CLI.Local.Path)), lines, characters)
	}
}

func CreateFile() *os.File {
	file, err := os.Create("script-length.csv")
	if err != nil {
		log.Fatalf("Could not create output file. %s", err)
	}
	return file
}

func ParseScripts(scripts []Script, name string, writer *csv.Writer) {
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
			l, c := CleanAndCountScript(r.String())
			ch <- Count{l, c}
			wg.Done()
		}(script)
	}

	wg.Wait()
	close(ch)

	for c := range ch {
		lines += c.lines
		characters += c.characters
	}

	writer.Write([]string{name, fmt.Sprint(lines), fmt.Sprint(characters)})
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

func FetchScript(id string, writer *csv.Writer) {
	r, err := fetch.Get(fmt.Sprintf("https://static.atlasacademy.io/JP/Script/%s/%s.txt", id[0:2], id))
	if err != nil {
		log.Fatalf("Error fetching script " + id)
		return
	}
	l, c := CleanAndCountScript(r.String())
	writer.Write([]string{id, fmt.Sprint(l), fmt.Sprint(c)})
}

func TraverseDirectories(path string, writer *csv.Writer) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Unable to get entries for directory ", path)
	}

	hasDirectory := slices.ContainsFunc(entries, func(e fs.DirEntry) bool {
		return e.IsDir()
	})

	// If directories exist, call this recursively for each one until we hit the lowest level
	if hasDirectory {
		for _, e := range entries {
			TraverseDirectories(filepath.Join(path, e.Name()), writer)
		}
		return
	}

	lines := 0
	characters := 0
	// TODO: This could be done with goroutines but it's pretty fast already
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(path, e.Name()))
		if err != nil {
			log.Fatalf("Can't read file: %s", filepath.Join(path, e.Name()))
		}
		l, c := CleanAndCountScript(string(data))
		lines += l
		characters += c
	}
	writer.Write([]string{filepath.Base(path), fmt.Sprint(lines), fmt.Sprint(characters)})
}

func CleanAndCountScript(data string) (int, int) {
	r, _ := regexp.Compile(`(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]|(？.+?：.+)`)
	matches := r.FindAllString(data, -1)
	lines := len(matches)
	characters := 0
	r, _ = regexp.Compile(`(\[[^#&]+?\]|[\[\]#&:]|？.+?：|^＠.+|\n)`)
	for _, m := range matches {
		characters += len([]rune(r.ReplaceAllString(m, "")))
	}

	return lines, characters
}
