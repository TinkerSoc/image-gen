# image-gen
Generates images for the TinkerSoc Website

Traverses a tree of images, and resizes them based on a configuration.

There are a handful of algorithms for resizing images. They are detailed
[here](https://godoc.org/github.com/disintegration/imaging#ResampleFilter).
`Box` is the fastest, `Lanczos` is the highest quality.

 - Box
 - BSpline
 - CatmullRom
 - Lanczos
 - Linear
 - MitchellNetravali
 - NearestNeighbor

```json
{
  "Paths": [
    {
      "Path": "/path/to/source",
      "Destination": "/path/to/destination",
      "Ignore": "/path/to/source/ignore",
      "Recursive": true,
      "Resize": [
        {
          "Algorithm": "Box",
          "Suffix": "-large",
          "Width": 960,
          "Height": 960,
          "KeepAspectRatio": true
        },
        {
          "Algorithm": "Lanczos",
          "Suffix": "-medium",
          "Width": 480,
          "Height": 480,
          "KeepAspectRatio": true
        }
      ]
    }
  ]
}

```
