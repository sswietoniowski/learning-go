package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var a = flag.Bool("a", false, "show all files")
	flag.Parse()
	files := listFiles("testdata", *a)
	for _, f := range files {
		fmt.Println(f)
	}
}

func listFiles(dirname string, showAll bool) []string {
	var dirs []string

	files, err := os.ReadDir(dirname)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if !showAll && f.Name()[0] == '.' {
			continue
		}
		dirs = append(dirs, f.Name())
	}

	return dirs
}
