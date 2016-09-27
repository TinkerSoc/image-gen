# image-gen
Generates images for the TinkerSoc Website

Traverses a tree of images, and resizes them based on a configuration

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
          "Suffix": "-large",
          "Width": 960,
          "Height": 960,
          "KeepAspectRatio": true
        },
        {
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
