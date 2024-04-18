package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Missing file name parameter")
		return
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Can't read file:", os.Args[1])
		panic(err)
	}

	r, _ := regexp.Compile(`＠.*\n(.|\n)+?\n\[k\]|(？.+?：)`)
	matches := r.FindAllStringIndex(string(data), -1)
	fmt.Println(len(matches))
}
