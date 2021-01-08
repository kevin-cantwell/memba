package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db = func() *sql.DB {
		databaseFile := filepath.Join(os.Getenv("HOME"), ".memba/sqlite.db")
		mkdir(databaseFile)
		db, err := sql.Open("sqlite3", databaseFile)
		if err != nil {
			log.Panicln(databaseFile, err)
		}
		migrate(db)
		return db
	}()
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalln("usage: memba [dis|dat] [dis or dat]")
	}
	switch os.Args[1] {
	case "dis":
		membadis(os.Args[2])
	case "dat":
		membadat(os.Args[2])
	default:
		log.Fatalln("usage: memba [dis|dat] [dis or dat]")
	}
}

type thing struct {
	ID    int64
	Type  string
	Value string
	Time  time.Time
}

func mkdir(path string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Panicln("mkdir", err)
	}
}

func migrate(db *sql.DB) {
	const q = `
		CREATE TABLE IF NOT EXISTS things (
			id integer PRIMARY KEY,
			type text,
			value text,
			time timestamp
		)
	`
	if _, err := db.Exec(q); err != nil {
		log.Panicln("migrate", err)
	}
}

func membadis(this string) {
	ct := http.DetectContentType([]byte(this))
	_, err := db.Exec("INSERT INTO things (type, value, time) VALUES (?, ?, ?)", ct, this, time.Now())
	if err != nil {
		log.Panicln("Exec", err)
	}
}

func membadat(dats ...string) {
	args := make([]interface{}, len(dats))
	likes := "like '%' || ? || '%'"
	args[0] = dats[0]
	for i, dat := range dats[1:] {
		likes += " OR like '%' || ? || '%'"
		args[i] = dat
	}
	rows, err := db.Query("SELECT id, type, value, time FROM things WHERE value "+likes+" ORDER BY id ASC", args...)
	if err != nil {
		log.Panicln("Query", err)
	}

	for rows.Next() {
		var t thing
		if err := rows.Scan(&t.ID, &t.Type, &t.Value, &t.Time); err != nil {
			log.Panicln("Scan", err)
		}
		fmt.Println(string(t.Value))
	}
	if err := rows.Err(); err != nil {
		log.Panicln("Err", err)
	}
}
