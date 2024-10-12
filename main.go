package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"vosskamp-reisen-3/models"
	"vosskamp-reisen-3/services"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

var db *sql.DB
var tmpl *template.Template

func main() {
	setUpDb()
	setUpTemplates()

	router := http.NewServeMux()
	router.HandleFunc("GET /", homeHandler)
	router.HandleFunc("GET /tasks", fetchTasksHandler)
	router.HandleFunc("GET /add-task-form", fetchAddTaskFormHandler)
	router.HandleFunc("GET /update-task-form/{id}", fetchUpdateTaskFormHandler)
	router.HandleFunc("POST /task", insertTaskHandler)
	router.HandleFunc("GET /task/{id}", fetchTaskHandler)
	router.HandleFunc("DELETE /task/{id}", deleteTaskHandler)
	router.HandleFunc("PUT /task/{id}", updateTaskHandler)

	fmt.Println("Listening on port 8080...")
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	err := server.ListenAndServe()
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
}

func setUpTemplates() {
	tmpl = template.Must(template.ParseGlob("./templates/*.html"))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("home")
	err := tmpl.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fetchTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := services.FetchTasks(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = tmpl.ExecuteTemplate(w, "task-list", tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fetchAddTaskFormHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "add-task-form", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fetchUpdateTaskFormHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("fetchUpdateTaskFormHandler")
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var task *models.Task
	task, err = services.FetchTaskById(db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = tmpl.ExecuteTemplate(w, "update-task-form", task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func insertTaskHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Insert task")
	taskName := r.FormValue("task")
	err := services.InsertTask(db, taskName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fetchTasksHandler(w, r)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	task := models.Task{
		Id:   id,
		Name: r.FormValue("name"),
		Done: r.FormValue("is-done") == "true" || r.FormValue("is-done") == "on",
	}

	_, err = services.UpdateTask(db, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fetchTasksHandler(w, r)
}

func fetchTaskHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fetch task")
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = services.RemoveTask(db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fetchTasksHandler(w, r)
}
