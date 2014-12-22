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
      "UnpackCommand": "dtrx --noninteractive --recursive --flat {{.Name}}",
      "PostCommand": "mv {{.Dir}} /tmp/"
    }
  ]
}
```

## Configuration options

`Paths` is an array of paths to watch.

`Name` is the path that should be watched.

`MinDepth` sets the minimum depth allowed to trigger `UnpackCommand`. A
`MinDepth` of `4` would allow archive files in `/home/foo/videos/bar` to trigger
an event.

`MaxDepth` sets the maximum depth allowed to trigger `UnpackCommand`. A `MaxDepth`
of `4` would allow archives files in `/home/foo/videos/bar/baz` to trigger an
event.

`SkipHidden` determines whether events for hidden files (files prefix with `.`)
should be ignored.

`Patterns` sets the wildcard patterns that a file needs to match to be able to
trigger an event.

`Remove` determines whether archive files should be deleted after
`UnpackCommand` has run.

`ArchiveExt` sets the archive extension that we expect to unpack.

`UnpackCommand` is the command used for unpacking files.

`PostCommand` is the command is an optional command to be run after
`UnpackCommand`.

## Command templates

The following template variables are available for use in the `UnpackCommand`
and `PostCommand` options:

Variable | Description                                    | Example
-------- | ---------------------------------------------- | -------
`Base`   | Basename of the archive file                   | `baz.rar`
`Dir`    | Directory holding the archive file             | `/tmp/foo/bar`
`Name`   | Full path to archive file triggering the event | `/tmp/foo/bar/baz.rar`

The template is compiled using the
[text/template](http://golang.org/pkg/text/template/) package. Variables can be
used like this: `{{.Name}}`

The working directory of `UnpackCommand` and `PostCommand` will be set to the
directory where the archive is located, equal to `{{.Dir}}`.
