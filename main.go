package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	baseDir = func() string {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		dir := filepath.Join(home, ".memba")
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
		return dir
	}()

	editor = func() string {
		if ed, ok := os.LookupEnv("EDITOR"); !ok {
			return "vim"
		} else {
			return ed
		}
	}()
)

func main() {
	start := time.Now()

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	flag.Parse()

	var (
		buf bytes.Buffer
	)

	for i, arg := range flag.Args() {
		if i > 0 {
			fmt.Fprintln(&buf, arg)
		}
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if _, err := io.Copy(&buf, os.Stdin); err != nil {
			log.Fatalln(err)
		}
	}

	// If there's no input, then open an editor and append the contents to the buffer
	if buf.Len() == 0 {
		f, err := ioutil.TempFile("", "")
		if err != nil {
			log.Fatalln(err)
		}
		defer func() {
			_ = os.Remove(f.Name())
		}()

		if err := f.Close(); err != nil {
			log.Fatalln(err)
		}

		executable, err := exec.LookPath(editor)
		if err != nil {
			log.Fatalln(err)
		}

		cmd := exec.Command(executable, f.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
		}

		bytes, err := ioutil.ReadFile(f.Name())
		if err != nil {
			log.Fatalln(err)
		}

		if _, err := buf.Write(bytes); err != nil {
			log.Fatalln(err)
		}
	}

	if len(flag.Args()) == 0 && buf.Len() == 0 {
		os.Exit(0)
	}

	notesPath := filepath.Join(baseDir, "notes.txt")

	// open the notes file for appending
	notes, err := os.OpenFile(notesPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = notes.Close()
	}()

	fmt.Fprintf(notes, "<%s> %s\n\n", start.Format(time.RFC3339), flag.Arg(0))
	if _, err := io.Copy(notes, &buf); err != nil {
		log.Fatalln(err)
	}
	fmt.Fprintf(notes, "\n\n")
}

type memory struct {
	Start   time.Time
	End     time.Time
	Note    string
	WorkDir string
	Tags    []string
}

type context struct {
}
