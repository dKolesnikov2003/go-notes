package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Note struct {
	Timestamp time.Time
	Title     string
	Text      string
}

var notesPath string

func getNotesPath() (string, error) {
	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		xdgData = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(xdgData, "go-notes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "notes.json"), nil
}

func main() {
	var err error
	notesPath, err = getNotesPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "ERROR: invalid arguments")
		printUsage(os.Stderr)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "-a", "--add":
		var title string
		if len(os.Args) > 2 {
			title = os.Args[2]
		}
		addNote(title)

	case "-l", "--list":
		listNotes()

	case "-s", "--show":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "ERROR: show requires a note number")
			printUsage(os.Stderr)
			os.Exit(1)
		}
		showNote(os.Args[2])

	case "-d", "--del":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "ERROR: del requires a note number")
			printUsage(os.Stderr)
			os.Exit(1)
		}
		deleteNote(os.Args[2])

	case "-h", "--help":
		printUsage(os.Stdout)

	default:
		fmt.Fprintln(os.Stderr, "ERROR: invalid arguments")
		printUsage(os.Stderr)
	}

}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `
USAGE
    go-notes <option> [argument]

OPTIONS
    -a, --add
        Add a new note. The title can be specified as an argument.
		Press Ctrl+D to finish input.

    -l, --list
        Display the list of all saved notes.

    -s, --show
        Show the text of a note selected by number.
        The number can be specified as an argument.

    -d, --del
        Delete a note selected by number.
        The number can be specified as an argument.

    -h, --help
        Display this help.`)
}

func addNote(title string) {
	f, err := os.OpenFile(notesPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	var notes []Note
	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &notes)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var text string
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	text = strings.Join(lines, "\n")

	n := Note{time.Now(), title, text}
	notes = append(notes, n)

	f.Truncate(0)
	f.Seek(0, 0)

	b, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_, err = f.Write(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func listNotes() {
	f, err := os.Open(notesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	var notes []Note
	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(data) == 0 {
		fmt.Fprintln(os.Stdout, "Not a single note has been created yet.")
		os.Exit(0)
	}
	err = json.Unmarshal(data, &notes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	const maxLen = 40
	for i, note := range notes {
		firstLine := strings.SplitN(note.Text, "\n", 2)[0]
		displayText := firstLine
		if len(displayText) > maxLen {
			displayText = displayText[:maxLen] + "..."
		}
		fmt.Printf("%2d. %s  [%s]\n    %s\n\n",
			i+1,
			note.Timestamp.Format("02/01/2006 15:04"),
			note.Title,
			displayText)
	}
}

func showNote(id string) {
	f, err := os.Open(notesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	var notes []Note
	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &notes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	i, err := strconv.Atoi(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	i--
	if i >= len(notes) || i < 0 {
		fmt.Fprintln(os.Stderr, "Invalid note number")
		os.Exit(1)
	}
	note := notes[i]
	fmt.Printf("%2d. %s  [%s]\n\n%s\n",
		i+1,
		note.Timestamp.Format("02/01/2006 15:04"),
		note.Title,
		note.Text)
}

func deleteNote(id string) {
	f, err := os.OpenFile(notesPath, os.O_RDWR, 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	var notes []Note
	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &notes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	i, err := strconv.Atoi(id)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	i--
	if i >= len(notes) || i < 0 {
		fmt.Fprintln(os.Stderr, "Invalid note number")
		os.Exit(1)
	}
	notes = append(notes[:i], notes[i+1:]...)

	f.Truncate(0)
	f.Seek(0, 0)

	b, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_, err = f.Write(b)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Note %s was deleted successfully\n", id)
}
