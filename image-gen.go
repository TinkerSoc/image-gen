package main

import (
	"encoding/json"
	"fmt"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/stretchr/powerwalk"
	//"github.com/urfave/cli"
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
		if p.Ignore != "" && strings.HasPrefix(path, p.Ignore) {
			log.Printf("skip this directory: %s\n", path)
			return nil
		}

		if fileInfo != nil && fileInfo.Mode().IsRegular() {
			destPath, name, extension := disassemblePaths(p, path)

			// ignore dot files
			if name == "" {
				return nil
			}

			img, err := imaging.Open(path)
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


	//app := cli.NewApp()
	//app.Name = "image-gen"
	//app.Usage = "build multiple resolution images for a static website"
	//app.Version = "0.1.1"
	//app.Author = [1]string{"Simon Cooksey"}
	//app.EnableBashCompletion = true
	flag.String("What", "0.1.1", "0.1.1")
	flag.String("Version", "image-gen", "image-gen")
	flag.String("Purpose", "build multiple resolution images for a static website", "build multiple resolution images for a static website")
	concurrencyLevelPtr := flag.Int("concurrency-level", 1,
			 "Set the number of threads for image-gen to use")
	configPathPtr := flag.String("config", "", 
			 "Load configuration from `FILE`")
	verbosePtr := flag.Bool("verbose", false, 
			 "Run verbosely")
	useConcurrenyPtr := flag.Bool("no-concurrency", false,
			 "Disable concurrent workers")//might be bad
	flag.Parse()

	concurrencyLevel := *concurrencyLevelPtr
	configPath := *configPathPtr
	verbose = *verbosePtr
	useConcurreny = *useConcurrenyPtr

	if(configPath=="") {
		fmt.Printf("config path parameter required\n")
	}

	file, _ := os.Open(configPath)
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Config{}

	// Set the runtime concurrency level
	runtime.GOMAXPROCS(concurrencyLevel)

	err := decoder.Decode(&configuration)
	if err != nil {
		log.Println(err)
	}

	for _, path := range configuration.Paths {
		path.Path, _ = filepath.Abs(filepath.Clean(path.Path))
		path.Destination, _ = filepath.Abs(filepath.Clean(path.Destination))
		if path.Ignore != "" {
			path.Ignore, _ = filepath.Abs(filepath.Clean(path.Ignore))
		}

		fmt.Printf("%s --> %s\nIgnoring %s\n\n", path.Path, path.Destination, path.Ignore)
		resizeDir(path)
	}

		//return nil
	//}

	//app.Run(os.Args)
}
