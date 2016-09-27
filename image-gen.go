package main

import (
  "fmt"
  "log"
  "flag"
  "os"
  "encoding/json"
  "path/filepath"

  // "image"
  // "image/color"
  "github.com/disintegration/imaging"
)

type PathDirective struct {
  Path              string
  Destination       string
  Ignore            string
  Recursive         bool
  Resize            []ResizeDirective
}

type ResizeDirective struct {
   Suffix           string
   Width            int
   Height           int
   KeepAspectRatio  bool
}

type Config struct {
  Paths             []PathDirective
}


func resizeDir(p PathDirective) {
  var resizeWalk = func(path string, fileInfo os.FileInfo, _ error) error {
    if fileInfo.Mode().IsRegular() {
      img, err := imaging.Open(path)
      if err != nil {
        return err
      }


      // FIXME: This is ugly, and appears to be broken
      var extension = filepath.Ext(path)
      var name = path[len(filepath.Dir(path)):len(path)-len(extension)]
      var basePath = p.Destination + filepath.Dir(path)[len(p.Path):]
      fmt.Printf("Path: %s\n  Name: %s\n  Extension: %s\n  BasePath: %s\n",
        path, name, extension, basePath)

      err = os.MkdirAll(basePath, 0777)
      if err != nil {
        return err
      }

      for _,r := range p.Resize {
        var width = 0
        var height = 0

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

        dstImage := imaging.Resize(img, width, height, imaging.Lanczos)
        err := imaging.Save(dstImage, basePath + name + r.Suffix + extension)
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

  for _,path := range configuration.Paths {
    resizeDir(path)
  }
}
