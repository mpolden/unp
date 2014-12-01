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
  -f, --config=FILE    Config file (~/.gounpackrc)
  -c, --colors         Use colors in log output

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

# UnpackCommand template

The following template variables are available for use in `UnpackCommand` option:

Variable | Description                                    | Example
-------- | ---------------------------------------------- | -------
`Base`   | Basename of the archive file                   | `baz.rar`
`Dir`    | Directory holding the archive file             | `/tmp/foo/bar`
`Name`   | Full path to archive file triggering the event | `/tmp/foo/bar/baz.rar`

The template is compiled using the
[text/template](http://golang.org/pkg/text/template/) package. Variables can be
used like this: `{{.Name}}`

The working directory of `UnpackCommand` will be set to the directory where the
archive is located, equal to `{{.Dir}}`.
