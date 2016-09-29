package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/disintegration/imaging"
	"github.com/stretchr/powerwalk"
	"github.com/urfave/cli"
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

func disassemblePaths(p PathDirective, path string) (string, string, string) {
	extension := filepath.Ext(path)
	_, name := filepath.Split(path)
	name = name[0 : len(name)-len(extension)]
	relativePath, _ := filepath.Rel(p.Path, path)
	destPath := filepath.Dir(p.Destination + string(filepath.Separator) + relativePath)

	return destPath, name, extension
}

func calculateSize(r ResizeDirective) (int, int) {
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

	return width, height
}

func resizeDir(p PathDirective) {
	var resizeWalk = func(path string, fileInfo os.FileInfo, _ error) error {
		if fileInfo != nil && fileInfo.Mode().IsRegular() {

			img, err := imaging.Open(path)
			destPath, name, extension := disassemblePaths(p, path)
			if err != nil {
				return err
			}

			err = os.MkdirAll(destPath, 0777)
			if err != nil {
				return err
			}

			for _, r := range p.Resize {
				width, height := calculateSize(r)
				dstImage := imaging.Resize(img, width, height, resampleFilterLookup(r.Algorithm))
				err := imaging.Save(dstImage, (destPath + string(filepath.Separator) + name + r.Suffix + extension))
				if verbose {
					fmt.Printf("%s --> %s\n", path, destPath+string(filepath.Separator)+name+r.Suffix+extension)
				}

				if err != nil {
					log.Println(err)
				}

				// m, _ := metadata.ReadTags(path)
				// for k, v := range m {
				// 	fmt.Printf("key[%s] value[%s]\n", k, v)
				// }

			}
		}

		return nil
	}

	if p.Recursive {
		err := error(nil)
		if useConcurreny {
			err = powerwalk.WalkLimit(p.Path, resizeWalk, runtime.NumCPU()*2)
		} else {
			err = filepath.Walk(p.Path, resizeWalk)
		}

		if err != nil {
			log.Println(err)
		}
	}
}

var verbose = false
var useConcurreny = true

func main() {
	concurrencyLevel := 1

	configPath := ""
	app := cli.NewApp()
	app.Name = "image-gen"
	app.Version = "0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Usage:       "Load configuration from `FILE`",
			Destination: &configPath,
		},
		cli.IntFlag{
			Name:        "concurrency-level",
			Usage:       "Set the number of threads for image-gen to use",
			Value:       runtime.NumCPU() * 2,
			Destination: &concurrencyLevel,
		},
		cli.BoolTFlag{
			Name:        "no-concurrency",
			Usage:       "Disable concurrent workers",
			Destination: &useConcurreny,
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "Run verbosely",
			Destination: &verbose,
		},
	}

	runtime.GOMAXPROCS(concurrencyLevel * 2)

	app.Action = func(c *cli.Context) error {
		file, _ := os.Open(configPath)
		defer file.Close()
		decoder := json.NewDecoder(file)
		configuration := Config{}

		err := decoder.Decode(&configuration)
		if err != nil {
			log.Println(err)
		}

		for _, path := range configuration.Paths {
			path.Destination = filepath.Clean(path.Destination)
			path.Path = filepath.Clean(path.Path)
			resizeDir(path)
		}

		return nil
	}

	app.Run(os.Args)
}
