package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"vosskamp-reisen-3/internal/models"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HomeHandler)
	mux.HandleFunc("/login", s.LoginHandler)
	mux.Handle("/cms/users", CheckSession(http.HandlerFunc(s.fetchUsersHandler)))
	mux.Handle("GET /cms/users/form", http.HandlerFunc(s.fetchUserFormHandler))
	mux.Handle("POST /cms/users/form", http.HandlerFunc(s.createUserFormHandler))
	// mux.Handle("/cms/users/form", CheckSession(http.HandlerFunc(s.fetchUserFormHandler)))

	mux.HandleFunc("/health", s.healthHandler)

	return mux
}

func (s *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	s.tmpl.ExecuteTemplate(w, "home.html", nil)
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "test",
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		HttpOnly: true,
		Secure:   true,
		MaxAge:   60 * 60 * 24 * 7,
	}
	http.SetCookie(w, cookie)
	s.tmpl.ExecuteTemplate(w, "login.html", nil)
}

func (s *Server) fetchUsersHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessionToken := cookie.Value
	fmt.Println(sessionToken)

	users, err := s.userService.FetchAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *Server) fetchUserFormHandler(w http.ResponseWriter, r *http.Request) {
	err := s.tmpl.ExecuteTemplate(w, "userCreation", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) createUserFormHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := models.Users{
		FirstName: r.FormValue("first-name"),
		LastName:  r.FormValue("last-name"),
		Email:     r.FormValue("email"),
		Password:  r.FormValue("password"),
	}
	fmt.Print(user)

	var errorMessages = map[string]bool{
		"firstName": false,
		"lastName":  false,
		"email":     false,
		"password":  false,
	}
	hasErrors := false
	if user.FirstName == "" {
		errorMessages["firstName"] = true
		hasErrors = true
	}
	if user.LastName == "" {
		errorMessages["lastName"] = true
		hasErrors = true
	}
	if user.Email == "" {
		errorMessages["email"] = true
		hasErrors = true
	}
	if user.Password == "" {
		errorMessages["password"] = true
		hasErrors = true
	}
	fmt.Print(errorMessages)
	if hasErrors {
		// s.tmpl.ExecuteTemplate(w, "authErrors", errorMessages)
		s.tmpl.ExecuteTemplate(w, "userForm", errorMessages)
		return
	}

	_, err = s.userService.CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")
	w.WriteHeader((http.StatusNoContent))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
