{
  "ImageProperties": null,
  "LabelModelVersion": "3.0",
  "Labels": [
    {
      "Aliases": null,
      "Categories": [
        {
          "Name": "Person Description"
        }
      ],
      "Confidence": 99.5,
      "Instances": [
        {
          "BoundingBox": {
            "Height": 0.8,
            "Left": 0.1,
            "Top": 0.1,
            "Width": 0.4
          },
          "Confidence": 98.5,
          "DominantColors": null
        }
      ],
      "Name": "Person",
      "Parents": null
    },
    {
      "Aliases": null,
      "Categories": [
        {
          "Name": "Person Description"
        }
      ],
      "Confidence": 99.5,
      "Instances": null,
      "Name": "Human",
      "Parents": null
    },
    {
      "Aliases": null,
      "Categories": [
        {
          "Name": "Person Description"
        }
      ],
      "Confidence": 98.8,
      "Instances": null,
      "Name": "Face",
      "Parents": [
        {
          "Name": "Person"
        },
        {
          "Name": "Human"
        }
      ]
    },
    {
      "Aliases": null,
      "Categories": [
        {
          "Name": "Places and Locations"
        }
      ],
      "Confidence": 85.3,
      "Instances": null,
      "Name": "Outdoors",
      "Parents": null
    },
    {
      "Aliases": null,
      "Categories": [
        {
          "Name": "Places and Locations"
        }
      ],
      "Confidence": 82.1,
      "Instances": null,
      "Name": "Nature",
      "Parents": [
        {
          "Name": "Outdoors"
        }
      ]
    }
  ],
  "OrientationCorrection": "",
  "ResultMetadata": {}
}