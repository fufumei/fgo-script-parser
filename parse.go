package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-zoox/fetch"
	"golang.org/x/exp/slices"
)

func main() {
	sourcePtr := flag.String("source", "atlas", "Atlas or local files")
	warPtr := flag.Int("war", 0, "the war ID in atlas")
	flag.Parse()

	if *warPtr == 0 {
		fmt.Println("War ID required")
		return
	}

	file, err := os.Create("script-length.csv")
	if err != nil {
		log.Fatalf("Could not create output file...")
	}
	w := csv.NewWriter(file)
	w.Comma = '\t'
	w.Write([]string{"Name", "Lines", "Characters"})
	if *sourcePtr == "atlas" {
		FetchScripts(strconv.Itoa(*warPtr))
	}
	// argInfo, err := os.Stat(os.Args[1])
	// if err != nil {
	// 	log.Fatalf(fmt.Sprintf("Could not get file info for %s", os.Args[1]))
	// }

	// // If given path is a file, just open and count it, else traverse the directory
	// if argInfo.IsDir() {
	// 	file, err := os.Create("script-length.csv")
	// 	if err != nil {
	// 		log.Fatalf("Could not create output file...")
	// 	}
	// 	w := csv.NewWriter(file)
	// 	w.Comma = '\t'
	// 	w.Write([]string{"Name", "Lines", "Characters"})

	// 	TraverseDirectories(os.Args[1], w)
	// 	w.Flush()
	// } else {
	// 	lines, characters := CountLinesChars(os.Args[1])
	// 	fmt.Printf("%s\t%d\t%d", strings.TrimSuffix(filepath.Base(os.Args[1]), filepath.Ext(os.Args[1])), lines, characters)
	// }
}

func FetchScripts(warId string) {
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
		log.Fatalf(fmt.Sprintf("Could not get data for war with ID %s", warId))
	}

	err = response.UnmarshalJSON(&result)
	if err != nil {
		log.Fatal("Error unmarshaling JSON:", err)
	}

	var scripts []Scripts
	for _, spot := range result.Spots {
		for _, quest := range spot.Quests {
			if quest.Type == "main" {
				for _, phase := range quest.PhaseScripts {
					scripts = append(scripts, phase.Scripts...)
				}
			}
		}
	}

	// s, _ := json.MarshalIndent(scripts, "", "\t")
	// fmt.Println(string(s))

	lines := 0
	characters := 0
	for _, script := range scripts {
		r, err := fetch.Get(script.Script)
		if err != nil {
			log.Fatalf("Error fetching script " + script.ScriptId)
			return
		}
		l, c := CountFromJson(r.String())
		lines += l
		characters += c
	}

	fmt.Println("Name\tLines\tCharacters")
	fmt.Printf("%s\t%d\t%d", result.Name, lines, characters)
}

func TraverseDirectories(path string, w *csv.Writer) {
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
			TraverseDirectories(filepath.Join(path, e.Name()), w)
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
		l, c := CountLinesChars(filepath.Join(path, e.Name()))
		lines += l
		characters += c
	}
	w.Write([]string{filepath.Base(path), fmt.Sprint(lines), fmt.Sprint(characters)})
}

func CountLinesChars(path string) (int, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Can't read file: %s", path))
	}

	r, _ := regexp.Compile(`(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]|(？.+?：.+)`)
	matches := r.FindAllString(string(data), -1)
	lines := len(matches)
	characters := 0
	r, _ = regexp.Compile(`(\[[^#&]+?\]|[\[\]#&:]|？.+?：|^＠.+|\n)`)
	for _, m := range matches {
		characters += len([]rune(r.ReplaceAllString(m, "")))
	}

	return lines, characters
}

func CountFromJson(script string) (int, int) {
	r, _ := regexp.Compile(`(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]|(？.+?：.+)`)
	matches := r.FindAllString(string(script), -1)
	lines := len(matches)
	characters := 0
	r, _ = regexp.Compile(`(\[[^#&]+?\]|[\[\]#&:]|？.+?：|^＠.+|\n)`)
	for _, m := range matches {
		characters += len([]rune(r.ReplaceAllString(m, "")))
	}

	return lines, characters
}
