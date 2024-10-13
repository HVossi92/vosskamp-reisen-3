package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HomeHandler)
	mux.HandleFunc("/login", s.LoginHandler)
	mux.Handle("/cms/users", CheckSession(http.HandlerFunc(s.fetchUsersHandler)))

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

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
