# unp

[![Build Status](https://travis-ci.org/mpolden/unp.svg)](https://travis-ci.org/mpolden/unp)

`unp` is a small server that monitors directories, verifies SFV files and
unpacks archives automatically.

## Usage

```
unp -h
Usage:
  unp [OPTIONS]

Application Options:
  -f, --config=FILE    Config file (default: ~/.unprc)
  -t, --test           Test and print config

Help Options:
  -h, --help           Show this help message
```

## Example config

```json
{
  "BufferSize": 1024,
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
      "PostCommand": "mv {{.Dir}} /tmp/"
    }
  ]
}
```

## Configuration options

`BufferSize` sets the maximum number of file system to queue. This should be
large enough to store events that occur while unpacking files. The default value
is `1024`.

`Paths` is an array of paths to watch.

`Name` is the path that should be watched.

`MinDepth` sets the minimum depth allowed to trigger unpacking. A `MinDepth` of
`4` would allow archive files in `/home/foo/videos/bar` to trigger an event.

`MaxDepth` sets the maximum depth allowed to trigger unpacking. A `MaxDepth` of
`5` would allow archives files in `/home/foo/videos/bar/baz` to trigger an
event.

`SkipHidden` determines whether events for hidden files (files prefix with `.`)
should be ignored.

`Patterns` sets the wildcard patterns that a file needs to match to be able to
trigger an event.

`Remove` determines whether archive files should be deleted after unpacking.

`PostCommand` is an optional command to be run after unpacking completes.

## Command templates

The following template variables are available for use in the `PostCommand`
option:

Variable | Description                                    | Example
-------- | ---------------------------------------------- | -------
`Base`   | Basename of the archive file                   | `baz.rar`
`Dir`    | Directory holding the archive file             | `/tmp/foo/bar`
`Name`   | Full path to archive file triggering the event | `/tmp/foo/bar/baz.rar`

The template is compiled using the
[text/template](http://golang.org/pkg/text/template/) package. Variables can be
used like this: `{{.Name}}`

The working directory of `PostCommand` will be set to the directory where the
archive is located, equal to `{{.Dir}}`.

## Signals

`unp` reacts to the following signals:

`SIGUSR1` triggers a re-scan which walks all configured paths and triggers
unpacking for any archives that are found.

`SIGUSR2` reloads configuration from disk. This can be used to watch new paths
without restarting the program.
