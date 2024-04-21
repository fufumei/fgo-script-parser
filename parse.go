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

	"golang.org/x/exp/slices"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing file or directory name")
		return
	}

	argInfo, err := os.Stat(os.Args[1])
	if err != nil {
		log.Fatalf(fmt.Sprintf("Could not get file info for %s", os.Args[1]))
	}

	// If given path is a file, just open and count it, else traverse the directory
	if argInfo.IsDir() {
		file, err := os.Create("script-length.csv")
		if err != nil {
			log.Fatalf("Could not create output file...")
		}
		w := csv.NewWriter(file)
		w.Comma = '\t'
		w.Write([]string{"Nane", "Lines", "Characters"})

		TraverseDirectories(os.Args[1], w)
		w.Flush()
	} else {
		lines, characters := CountLinesChars(os.Args[1])
		fmt.Printf("%s\t%d\t%d", strings.TrimSuffix(filepath.Base(os.Args[1]), filepath.Ext(os.Args[1])), lines, characters)
	}

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
