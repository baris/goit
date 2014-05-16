package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Filter(s []int, fn func(int) bool) (p []int) {
	for _, i := range s {
		if fn(i) {
			p = append(p, i)
		}
	}
	return p
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func has(path, fileOrDir string) bool {
	return exists(filepath.Join(path, fileOrDir))
}

func removeExt(path string) string {
	ext := filepath.Ext(path)
	return path[:len(path)-len(ext)]
}

func toCSSName(path string) string {
	return strings.Replace(
		strings.Replace(path, "/", "_", -1),
		".",
		"_",
		-1)
}

// This is a generator that reads lines from a file
func readLines(path string) chan string {
	ch := make(chan string)
	go func() {
		inputFile, err := os.Open(config.Git_projects_file)
		if err != nil {
			log.Println("Failed to read the projects file")
			return
		}
		scanner := bufio.NewScanner(inputFile)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
		close(ch)
	}()
	return ch
}
