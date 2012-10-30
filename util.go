package main

import (
	"os"
	"path/filepath"
)

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
