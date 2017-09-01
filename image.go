package main

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var imageCache = make(map[string]image.Image)

func findImage(name string) (img image.Image, finalErr error) {
	if img, ok := imageCache[name]; ok {
		return img, nil
	}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if name == strings.TrimSuffix(info.Name(), ext) {
				img, finalErr = loadImage(path)
				if finalErr == nil {
					imageCache[name] = img
					return errors.New("done")
				}
			}

		}
		return nil
	})
	if img == nil && finalErr == nil {
		finalErr = fmt.Errorf("no image with the name '%s' found", name)
	}
	return
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
