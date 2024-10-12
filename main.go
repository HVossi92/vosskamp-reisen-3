package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

var db *sql.DB
var tmpl *template.Template

func main() {
	setUpDb()

	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandler)

	fmt.Println("Listening on port 8080...")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		panic(err)
	}
}

func setUpDb() {
	fmt.Println("Connecting to database...")
	var err error
	db, err = sql.Open("sqlite3", "./database.db?_journal=WAL&_busy_timeout=5000&_foreign_keys=on&_synchronous=NORMAL")
	if err != nil {
		panic(err)
	}

	setUpTemplate()

	fmt.Println("Initializing database...")
	statement, err := db.Prepare(("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY AUTOINCREMENT, first_name TEXT, last_name TEXT, email TEXT, phone TEXT, address TEXT, created_at TEXT, position INTEGER) strict;"))
	if err != nil {
		panic(err)
	}
	statement.Exec()

	defer db.Close()
}

func setUpTemplate() {
	tmpl = template.Must(template.ParseGlob("./templates/*.html"))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
