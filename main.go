package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	// _ "github.com/mattn/go-sqlite3"
)

var (
	baseDir = func() string {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		dir := filepath.Join(home, "Documents", "memba")
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

	mdPath := filepath.Join(baseDir, fmt.Sprintf("%s %s.md", time.Now().Format("2006-01-02"), title))
	mdFile, err := os.OpenFile(mdPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		_ = mdFile.Close()
	}()

	stat, _ := os.Stdin.Stat()

	switch {
	case (stat.Mode() & os.ModeCharDevice) == 0:
		if _, err := io.Copy(mdFile, os.Stdin); err != nil {
			log.Fatalln(err)
		}
	case len(flag.Args()) > 1:
		for _, arg := range flag.Args()[1:] {
			fmt.Fprintln(mdFile, arg)
		}
	default:
		cmd := exec.Command("open", "-a", "Obsidian")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalln(err)
		}
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
