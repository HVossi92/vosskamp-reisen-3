package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"vosskamp-reisen-3/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HomeHandler)
	mux.HandleFunc("GET /login", s.fetchLoginPageHandler)
	mux.HandleFunc("POST /login", s.loginHandler)
	mux.Handle("/cms/users", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchUsersHandler)))
	mux.Handle("GET /cms/users/form", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchUserFormHandler)))
	mux.Handle("POST /cms/users/form", s.middleWareService.CheckSession(http.HandlerFunc(s.createUserFormHandler)))

	mux.HandleFunc("/health", s.healthHandler)

	return mux
}

func (s *Server) HomeHandler(w http.ResponseWriter, r *http.Request) {
	s.tmpl.ExecuteTemplate(w, "home.html", nil)
}

func (s *Server) fetchLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	s.tmpl.ExecuteTemplate(w, "login", nil)
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var errorMessages = map[string]bool{
		"missingEmail":    false,
		"missingPassword": false,
		"invalidEmail":    false,
		"invalidPassword": false,
	}
	hasErrors := false
	if email == "" {
		errorMessages["missingEmail"] = true
		hasErrors = true
	}
	if password == "" {
		errorMessages["missingPassword"] = true
		hasErrors = true
	}

	errorFormData := map[string]interface{}{
		"errorMessages": errorMessages,
		"email":         email,
	}

	if hasErrors {
		s.tmpl.ExecuteTemplate(w, "loginForm", errorFormData)
		return
	}

	user, err := s.userService.FetchUserByEmail(email)
	if err != nil {
		errorMessages["invalidEmail"] = true
		s.tmpl.ExecuteTemplate(w, "loginForm", errorFormData)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		errorMessages["invalidPassword"] = true
		s.tmpl.ExecuteTemplate(w, "loginForm", errorFormData)
		return
	}

	sessionCookie, err := s.authService.GenerateCookie("session_token", true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.tokenService.InsertToken(sessionCookie.Value, sessionCookie.Expires, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, sessionCookie)

	w.Header().Set("HX-Location", "/")
	w.WriteHeader((http.StatusNoContent))
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

	var errorMessages = map[string]bool{
		"firstName":        false,
		"lastName":         false,
		"email":            false,
		"password":         false,
		"passwordTooShort": false,
		"invalidEmail":     false,
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
	if len(user.Password) < 8 {
		errorMessages["passwordTooShort"] = true
		hasErrors = true
	}
	if s.authService.IsValidEmail(user.Email) == false {
		errorMessages["invalidEmail"] = true
		hasErrors = true
	}

	if hasErrors {
		errorFormData := map[string]interface{}{
			"errorMessages": errorMessages,
			"user":          user,
		}
		s.tmpl.ExecuteTemplate(w, "userForm", errorFormData)
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
