package main

import (
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
		dir := filepath.Join(home, "Documents", "Obsidian")
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
	// info, err := os.Stat("/Users/kcantwell/Documents/Obsidian/Data Plane.md")
	// if err != nil {
	// 	panic(err)
	// }
	// sysstat, _ := info.Sys().(*syscall.Stat_t)

	// fmt.Printf("Birthtime: %s\n", time.Unix(sysstat.Birthtimespec.Unix()))
	// fmt.Printf("LastModified: %s\n", info.ModTime())
	// os.Exit(0)

	// start := time.Now()

	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)

	flag.Usage = func() {
		fmt.Print(`Usage:
    memba TITLE
    memba TITLE < FILE
    cat FILE | memba TITLE
`)
	}
	flag.Parse()

	if flag.Arg(0) == "" {
		flag.Usage()
		return
	}

	var (
		title = flag.Arg(0)
	)

	tempFile, err := ioutil.TempFile("", title)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = os.Remove(tempFile.Name())
	}()

	stat, _ := os.Stdin.Stat()

	switch {
	case (stat.Mode() & os.ModeCharDevice) == 0:
		if _, err := io.Copy(tempFile, os.Stdin); err != nil {
			log.Fatalln(err)
		}
	case len(flag.Args()) > 1:
		for _, arg := range flag.Args()[1:] {
			fmt.Fprintln(tempFile, arg)
		}
	default:
		executable, err := exec.LookPath(editor)
		if err != nil {
			log.Fatalln(err)
		}

		cmd := exec.Command(executable, tempFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
		}
	}

	if err := tempFile.Close(); err != nil {
		log.Fatalln(err)
	}

	tempFile, err = os.Open(tempFile.Name())
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = tempFile.Close()
	}()

	info, err := tempFile.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	if info.Size() == 0 {
		return
	}

	mdPath := filepath.Join(baseDir, title+".md")

	// open the notes file for appending
	mdFile, err := os.OpenFile(mdPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = mdFile.Close()
	}()

	if _, err := io.Copy(mdFile, tempFile); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%q\n", mdPath)
}

type memory struct {
	Start   time.Time
	End     time.Time
	Note    string
	WorkDir string
	Tags    []string
}
