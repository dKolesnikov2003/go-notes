# go-notes
A simple Go CLI utility for creating notes.

## Installing
`go install github.com/dKolesnikov2003/go-notes@latest`

## Ussage
`go-notes --add "Title"`

`go-notes --list`

`go-notes --show 1`

`go-notes --del 1`

type `go-notes --help` for details

## Where are the notes stored?

In the `notes.json` file located at:

- `$XDG_DATA_HOME/go-notes/notes.json` if `XDG_DATA_HOME` is set
- `~/.local/share/go-notes/notes.json` by default

