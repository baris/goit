package main

import (
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
