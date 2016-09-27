package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

// PathDirective contains instruction for traversing a path of images
type PathDirective struct {
	Path        string
	Destination string
	Ignore      string
	Recursive   bool
	Resize      []ResizeDirective
}

// ResizeDirective contains instructions to resize an images within a path
type ResizeDirective struct {
	Algorithm       string
	Suffix          string
	Width           int
	Height          int
	KeepAspectRatio bool
}

// Config contains a list of PathDirectives to run the program over
type Config struct {
	Paths []PathDirective
}

func resampleFilterLookup(name string) imaging.ResampleFilter {
	switch name {
	default:
	case "Box":
		return imaging.Box
	case "BSpline":
		return imaging.BSpline
	case "CatmullRom":
		return imaging.CatmullRom
	case "Lanczos":
		return imaging.Lanczos
	case "Linear":
		return imaging.Linear
	case "MitchellNetravali":
		return imaging.MitchellNetravali
	case "NearestNeighbor":
		return imaging.NearestNeighbor
	}
	// Unreacable, but Go doesn't care...
	return imaging.Box
}

func resizeDir(p PathDirective) {
	var resizeWalk = func(path string, fileInfo os.FileInfo, _ error) error {
		if fileInfo.Mode().IsRegular() {
			img, err := imaging.Open(path)
			if err != nil {
				return err
			}

			// FIXME: This is ugly, and appears to be broken
			extension := filepath.Ext(path)
			_, name := filepath.Split(path)
			name = name[0 : len(name)-len(extension)]
			relativePath, _ := filepath.Rel(p.Path, path)
			destPath := filepath.Dir(p.Destination + relativePath)

			fmt.Printf("Path: %s\n  Name: %s\n  Extension: %s\n  BasePath: %s\n",
				path, name, extension, relativePath)

			err = os.MkdirAll(destPath, 0777)
			if err != nil {
				return err
			}

			for _, r := range p.Resize {
				width := 0
				height := 0

				if r.KeepAspectRatio {
					if r.Width > r.Height {
						width = r.Width
					} else {
						height = r.Height
					}
				} else {
					width = r.Width
					height = r.Height
				}

				dstImage := imaging.Resize(img, width, height, resampleFilterLookup(r.Algorithm))
				err := imaging.Save(dstImage, (destPath + "/" + name + r.Suffix + extension))
				if err != nil {
					log.Println(err)
				}
			}
		}

		if fileInfo.IsDir() {
			fmt.Printf("%s\n", path)
		}

		return nil
	}

	if p.Recursive {
		err := filepath.Walk(p.Path, resizeWalk)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	configPath := flag.String("config", "", "The configuration file")

	flag.Parse()

	file, _ := os.Open(*configPath)
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Config{}

	err := decoder.Decode(&configuration)
	if err != nil {
		log.Println(err)
	}

	for _, path := range configuration.Paths {
		resizeDir(path)
	}
}
