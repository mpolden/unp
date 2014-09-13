# gounpack

[![Build Status](https://travis-ci.org/martinp/gounpack.png)](https://travis-ci.org/martinp/gounpack)

gounpack is a small server that monitors directories, verifies SFV files and
unpacks archives automatically.

## Usage

```
$ gounpack -h
Usage:
  gounpack [OPTIONS]

Application Options:
  -c, --config=FILE    Config file (config.json)
  -p, --colors         Use colors in output

Help Options:
  -h, --help           Show this help message
```

## Example config

```json
{
  "Paths": [
    {
      "Name": "/home/foo/videos",
      "MinDepth": 4,
      "MaxDepth": 5,
      "SkipHidden": true,
      "Patterns": [
        "*.r??",
        "*.sfv"
      ],
      "Remove": false,
      "ArchiveExt": "rar",
      "UnpackCommand": "dtrx --noninteractive --recursive --flat {{.Name}}"
    }
  ]
}
```
