package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type Data struct {
	Person Person
}

type Person struct {
	FirstName string
	LastName  string
	Email     string
	Infos     map[string]string
	Interests []string
	Created   time.Time
	Position  int
}

func upperCase(s string) string {
	return strings.ToUpper(s)
}

func getYearMonthDayFrom(date time.Time) string {
	return date.Format("02-01-2006")
}

func main() {
	fmt.Println("Starting server...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		funcMap := template.FuncMap{
			"upperCase":           upperCase,
			"getYearMonthDayFrom": getYearMonthDayFrom,
		}
		template := template.Must(template.New("templates/*.html").Funcs(funcMap).ParseGlob("templates/*.html"))

		data := Data{
			Person{
				FirstName: "Guillem",
				LastName:  "Bonet", Email: "hello@bunetz.dev",
				Infos:     map[string]string{"Address": "Barcelona", "Phone": "123456789", "Date": "2023-01-01"},
				Interests: []string{"Golang", "Web Development", "Design"},
				Created:   time.Now(),
				Position:  40,
			},
		}

		err := template.ExecuteTemplate(w, "home.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.ListenAndServe(":8080", nil)
}
