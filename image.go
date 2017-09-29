package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var imageCache = make(map[string]image.Image)

// findImage walks the "." directory in a breadth-first search to find a file
// with the given name, case-insensitive. It loads the image from the file and
// returns it.
func findImage(name string) (image.Image, error) {
	name = strings.ToLower(name)

	if img, ok := imageCache[name]; ok {
		return img, nil
	}

	var img image.Image
	var finalErr error

	dirQueue := []string{"."}
	for len(dirQueue) > 0 {
		dir := dirQueue[0]
		dirQueue = dirQueue[1:]
		files, err := ioutil.ReadDir(dir)
		if err == nil {
			for _, file := range files {
				if file.IsDir() {
					dirQueue = append(dirQueue, filepath.Join(dir, file.Name()))
				} else if strings.ToLower(file.Name()) == name {
					img, finalErr = loadImage(filepath.Join(dir, file.Name()))
					if finalErr == nil {
						imageCache[name] = img
						return img, nil
					}
				}
			}
		}
	}

	if img == nil && finalErr == nil {
		finalErr = fmt.Errorf("no image with the name '%s' found", name)
	}
	return img, finalErr
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
