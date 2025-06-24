# unp

![Build Status](https://github.com/mpolden/unp/workflows/ci/badge.svg)

`unp` monitors a list of configured directories and automatically verifies and
unpacks multi-volume RAR archives as they are written. A combination of file
system events and SFV files are used to determine when archives should be
unpacked.

## Usage

```
$ unp -h
Usage of unp:
  -f string
    	Path to config file (default "~/.unprc")
  -t	Test and print config
```

## Example config

```json
{
  "BufferSize": 1024,
  "Paths": [
    {
      "Name": "/home/foo/videos",
      "Handler": "rar",
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

`BufferSize` sets the maximum number of file system events to queue in memory.
This should be large enough to store any events that occur while an event is
processed by its handler. The default value is `1024`.

`Paths` is an array of paths to watch.

`Name` is the path that should be watched.

`Handler` sets the handler to use. This can be either `rar` (default if
unspecified) or `script`. The `rar` handler automatically unpacks RAR archives
and uses SFV files to determine completeness. The `script` handler calls the
specified `PostCommand` without any processing or completeness checks.

`MinDepth` sets the minimum path depth allowed to trigger the handler. A
`MinDepth` of `4` would allow the path `/home/foo/videos/bar.mkv` to trigger an
event. Path depth counts all path segments, including the file.

`MaxDepth` sets the maximum depth allowed to trigger the handler. A `MaxDepth`
of `5` would allow files in `/home/foo/videos/bar/baz` to trigger an event.

`SkipHidden` determines whether events for hidden files (files prefix with `.`)
should be ignored.

`Patterns` sets the wildcard patterns that a file needs to match to trigger an
event.

`Remove` determines whether the handler should delete files after processing
them.

`PostCommand` is an optional command to run after the handler processing
completes.

## Command templates

The following template variables are available for use in the `PostCommand`
option:

Variable | Description                                    | Example
-------- | ---------------------------------------------- | -------
`Base`   | Basename of the file triggering the event      | `baz.rar`
`Dir`    | Directory holding the file                     | `/tmp/foo/bar`
`Name`   | Full path to archive file triggering the event | `/tmp/foo/bar/baz.rar`

The template is compiled using the
[text/template](http://golang.org/pkg/text/template/) package. Variables can be
used like this: `{{.Name}}`

The working directory of `PostCommand` will be set to the directory where the
archive is located, equal to `{{.Dir}}`.

## Signals

`unp` reacts to the following signals:

`SIGUSR1` triggers a re-scan which walks all configured paths and triggers its
handler for any matching files that are found.

`SIGUSR2` reloads configuration from disk. This can be used to watch new paths
without restarting the program.
