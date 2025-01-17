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

	"github.com/alecthomas/kong"
	"github.com/go-zoox/fetch"
	"golang.org/x/exp/slices"
)

// fgo-script-parser atlas [--war <ID>] [--quest <ID>] [<script ID]
// fgo-script-parser <path> [--no-file] [--ignore-splits]

var CLI struct {
	NoFile bool `name:"no-file" default:"true"`

	Atlas struct {
		War    string `required short:"w" xor:"type"`
		Quest  string `required short:"q" xor:"type" hidden:""`
		Script string `required short:"s" xor:"type" hidden:""`
	} `cmd:""`
	Local struct {
		IgnoreSplits bool `name:"ignore-splits" default:"false"`

		Path string `arg:"" name:"path" type:"path"`
	} `cmd:""`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "atlas":
		RunAtlas()
	case "local <path>":
		RunLocal()
	default:
		panic(ctx.Command())
	}
}

func RunAtlas() {
	var writer *csv.Writer
	if !CLI.NoFile {
		file := CreateFile()
		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	writer.Comma = '\t'
	writer.Write([]string{"Name", "Lines", "Characters"})

	// TODO: Only handles war IDs for now
	FetchScripts(CLI.Atlas.War, writer)
	writer.Flush()
}

func RunLocal() {
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
		log.Fatalf("Could not get file info for %s", CLI.Local.Path)
	}

	// If given path is a file, just open and count it, else traverse the directory
	if argInfo.IsDir() {
		TraverseDirectories(os.Args[1], writer)
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

func FetchScripts(warId string, writer *csv.Writer) {
	type Scripts struct {
		ScriptId string `json:"scriptId"`
		Script   string `json:"script"`
	}

	type PhaseScripts struct {
		Phase   int `json:"phase"`
		Scripts []Scripts
	}

	type Quests struct {
		Id           int    `json:"id"`
		Type         string `json:"type"`
		PhaseScripts []PhaseScripts
	}

	type Response struct {
		Name  string `json:"name"`
		Spots []struct {
			Quests []Quests
		}
	}

	var result Response
	response, err := fetch.Get("https://api.atlasacademy.io/nice/JP/war/" + warId + "?lang=en")
	if err != nil {
		log.Fatalf("Could not get data for war with ID %s", warId)
	}
	err = response.UnmarshalJSON(&result)
	if err != nil {
		log.Fatal("Error unmarshaling JSON:", err)
	}

	fmt.Printf("Extracting scripts...\n")
	var scripts []Scripts
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
	fmt.Printf("Finished extracting scripts.\n\n")

	lines := 0
	characters := 0
	// One line in OC2 is missing the @ and can't be counted regularly
	if warId == "403" {
		lines = 1
	}
	// TODO: This could be done async in parallel
	for _, script := range scripts {
		fmt.Printf("Fetching %s...\n", script.ScriptId)
		r, err := fetch.Get(script.Script)
		if err != nil {
			log.Fatalf("Error fetching script " + script.ScriptId)
			return
		}
		fmt.Printf("Counting %s...\n", script.ScriptId)
		l, c := CleanAndCountScript(r.String())
		lines += l
		characters += c
		fmt.Printf("Next script...\n")
	}
	writer.Write([]string{result.Name, fmt.Sprint(lines), fmt.Sprint(characters)})
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
	// One line in OC2 is missing the @ and can't be counted regularly
	if slices.ContainsFunc(entries, func(e fs.DirEntry) bool {
		return strings.HasPrefix(e.Name(), "040003")
	}) {
		lines = 1
	}

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
