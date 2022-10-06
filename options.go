package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)



type processOptions struct {
	clean       bool
	Source      string
	Destination string
	Extension   string

	// AutoGeneratedDir - default '_'.  This is where resized images get placed
	AutoGeneratedDir string
	// UsePlaceholderImages - use grey placeholder images when no images are found on disk
	UsePlaceholderImages bool


	GlobalScope GlobalScope
}

func (po *processOptions) ResolvePath(path string) (string, error) {
	d := filepath.Join(po.Destination, path)
	s := filepath.Join(po.Source, path)

	stat, err := os.Stat(d)
	if err == nil {
		if !stat.Mode().IsRegular() {
			return d, fmt.Errorf("%s is not a regular file", d)
		}
		return d, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	stat, err = os.Stat(s)
	if err == nil {
		if !stat.Mode().IsRegular() {
			return s, fmt.Errorf("%s is not a regular file", s)
		}
		return s, nil
	} else {
		return "", err
	}
}
