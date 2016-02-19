# gounpack

[![Build Status](https://travis-ci.org/martinp/gounpack.svg)](https://travis-ci.org/martinp/gounpack)

gounpack is a small server that monitors directories, verifies SFV files and
unpacks archives automatically.

## Usage

```
$ gounpack -h
Usage:
  gounpack [OPTIONS]

Application Options:
  -b, --buffer-size=COUNT    Number of events to buffer (100)
  -f, --config=FILE          Config file (~/.gounpackrc)
  -t, --test                 Test and print config

Help Options:
  -h, --help                 Show this help message
```

## Example config

```json
{
  "Async": false,
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
      "ArchiveExt": ".rar",
      "UnpackCommand": "dtrx --noninteractive --recursive --flat {{.Name}}",
      "PostCommand": "mv {{.Dir}} /tmp/"
    }
  ]
}
```

## Configuration options

`Async` determines whether to do unpacking asynchronously. Set to `false` to
queue unpack events and process them one at a time, this might be faster as
unpacking is often I/O bound.

`Paths` is an array of paths to watch.

`Name` is the path that should be watched.

`MinDepth` sets the minimum depth allowed to trigger `UnpackCommand`. A
`MinDepth` of `4` would allow archive files in `/home/foo/videos/bar` to trigger
an event.

`MaxDepth` sets the maximum depth allowed to trigger `UnpackCommand`. A `MaxDepth`
of `5` would allow archives files in `/home/foo/videos/bar/baz` to trigger an
event.

`SkipHidden` determines whether events for hidden files (files prefix with `.`)
should be ignored.

`Patterns` sets the wildcard patterns that a file needs to match to be able to
trigger an event.

`Remove` determines whether archive files should be deleted after
`UnpackCommand` has run.

`ArchiveExt` sets the archive extension that we expect to unpack.

`UnpackCommand` is the command used for unpacking files.

`PostCommand` is an optional command to be run after `UnpackCommand`.

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
