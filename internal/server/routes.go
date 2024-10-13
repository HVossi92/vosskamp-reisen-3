package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"vosskamp-reisen-3/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.homeFormHandler)
	mux.HandleFunc("GET /login", s.fetchLoginPageHandler)
	mux.HandleFunc("POST /login", s.loginHandler)
	mux.Handle("GET /profile", s.middleWareService.CheckSession(http.HandlerFunc(s.profileFormHandler)))
	mux.Handle("GET /edit-profile", s.middleWareService.CheckSession(http.HandlerFunc(s.editProfileFormHandler)))
	mux.Handle("PUT /edit-profile", s.middleWareService.CheckSession(http.HandlerFunc(s.updateProfile)))
	mux.Handle("GET /upload-avatar", s.middleWareService.CheckSession(http.HandlerFunc(s.uploadAvatarForm)))
	mux.Handle("POST /upload-avatar", s.middleWareService.CheckSession(http.HandlerFunc(s.uploadAvatar)))
	mux.Handle("DELETE /logout", s.middleWareService.CheckSession(http.HandlerFunc(s.logoutHandler)))
	mux.Handle("GET /users", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchUsersHandler)))
	mux.Handle("DELETE /users/{id}", s.middleWareService.CheckSession(http.HandlerFunc(s.deleteUserHandler)))
	mux.Handle("GET /users/form", s.middleWareService.CheckSession(http.HandlerFunc(s.fetchUserFormHandler)))
	mux.Handle("POST /users/form", s.middleWareService.CheckSession(http.HandlerFunc(s.createUserFormHandler)))

	fileServer := http.FileServer(http.Dir("./uploads"))
	mux.Handle("GET /uploads/", s.middleWareService.CheckSession(http.StripPrefix("/uploads/", fileServer)))

	mux.HandleFunc("GET /health", s.healthHandler)

	return mux
}

func (s *Server) homeFormHandler(w http.ResponseWriter, r *http.Request) {
	s.tmpl.ExecuteTemplate(w, "home.html", nil)
}

func (s *Server) profileFormHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.Users)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data := struct {
		User *models.Users
		Tab  string
	}{
		User: user,
		Tab:  "one",
	}

	s.tmpl.ExecuteTemplate(w, "profile.html", data)
}

func (s *Server) updateProfile(w http.ResponseWriter, r *http.Request) {
	requestUser, ok := r.Context().Value("user").(*models.Users)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := models.Users{
		FirstName: r.FormValue("first-name"),
		LastName:  r.FormValue("last-name"),
		Email:     r.FormValue("email"),
	}

	var errorMessages = map[string]bool{
		"FirstName":    false,
		"LastName":     false,
		"Email":        false,
		"InvalidEmail": false,
	}
	hasErrors := false
	if user.FirstName == "" {
		errorMessages["FirstName"] = true
		hasErrors = true
	}
	if user.LastName == "" {
		errorMessages["LastName"] = true
		hasErrors = true
	}
	if user.Email == "" {
		errorMessages["Email"] = true
		hasErrors = true
	}
	if s.authService.IsValidEmail(user.Email) == false {
		errorMessages["InvalidEmail"] = true
		hasErrors = true
	}

	if hasErrors {
		errorFormData := map[string]interface{}{
			"ErrorMessages": errorMessages,
			"User":          user,
			"NoPassword":    true,
		}
		s.tmpl.ExecuteTemplate(w, "userForm", errorFormData)
		return
	}

	user.ID = requestUser.ID
	user.Password = requestUser.Password
	_, err = s.userService.UpdateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/")
	w.WriteHeader((http.StatusNoContent))
}

func (s *Server) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	// Initialize error messages slice
	var errorMessages []string

	// Parse the multipart form, 10 MB max upload size
	r.ParseMultipartForm(10 << 20)

	// Retrieve the file from form data
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		if err == http.ErrMissingFile {
			errorMessages = append(errorMessages, "No file submitted")
		} else {
			errorMessages = append(errorMessages, "Error retrieving the file")
		}

		if len(errorMessages) > 0 {
			s.tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
			return
		}

	}
	s.tmpl.ExecuteTemplate(w, "autherrors", errorMessages)

	// Generate a unique filename to prevent overwriting and conflicts
	uuid, err := uuid.NewRandom()
	if err != nil {
		errorMessages = append(errorMessages, "Error generating unique identifier")
		s.tmpl.ExecuteTemplate(w, "autherrors", errorMessages)

		return
	}
	filename := uuid.String() + filepath.Ext(handler.Filename) // Append the file extension

	// Create the full path for saving the file
	filePath := filepath.Join("uploads", filename)

	// Save the file to the server
	dst, err := os.Create(filePath)
	if err != nil {
		errorMessages = append(errorMessages, "Error saving the file")
		s.tmpl.ExecuteTemplate(w, "autherrors", errorMessages)

		return
	}
	defer dst.Close()
	if _, err = io.Copy(dst, file); err != nil {
		errorMessages = append(errorMessages, "Error saving the file")
		s.tmpl.ExecuteTemplate(w, "autherrors", errorMessages)
		return
	}

	user, ok := r.Context().Value("user").(*models.Users)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Update the user's avatar in the database
	//userID := r.FormValue("userID") // Assuming you pass the userID somehow
	if err := s.userService.UpdateUserAvatar(user.ID, filename); err != nil {
		fmt.Println(err)
		errorMessages = append(errorMessages, "Error updating user avatar")
		s.tmpl.ExecuteTemplate(w, "autherrors", errorMessages)

		log.Fatal(err)
		return
	}

	//Delete current image from the initial fetch of the user
	if user.Avatar != "" {
		oldAvatarPath := filepath.Join("uploads", user.Avatar)

		//Check if the oldPath is not the same as the new path
		if oldAvatarPath != filePath {
			if err := os.Remove(oldAvatarPath); err != nil {
				fmt.Printf("Warning: failed to delete old avatar file: %s\n", err)
			}
		}
	}

	//Navigate to the profile page after the update
	w.Header().Set("HX-Location", "/")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) uploadAvatarForm(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.Users)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data := struct {
		User *models.Users
		Tab  string
	}{
		User: user,
		Tab:  "three",
	}

	s.tmpl.ExecuteTemplate(w, "uploadAvatar", data)
}

func (s *Server) editProfileFormHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.Users)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	errorMessages := map[string]bool{
		"missingEmail":    false,
		"missingPassword": false,
		"invalidEmail":    false,
		"invalidPassword": false,
	}

	data := struct {
		User          *models.Users
		ErrorMessages map[string]bool
		NoPassword    bool
		Tab           string
	}{
		User:          user,
		ErrorMessages: errorMessages,
		NoPassword:    true,
		Tab:           "two",
	}

	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	s.tmpl.ExecuteTemplate(w, "editProfile", data)
}

func (s *Server) fetchLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		s.tmpl.ExecuteTemplate(w, "login", nil)
		return
	}
	sessionToken := cookie.Value
	_, err = s.tokenService.ValidateToken(sessionToken)
	if err != nil {
		s.tmpl.ExecuteTemplate(w, "login", nil)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value("user").(*models.Users)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	s.tokenService.RemoveToken(user.ID)
	if r.Header.Get("HX-Request") == "true" {
		// Send HX-Redirect header to force a GET request to /login
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusOK)
	} else {
		// Otherwise, use a normal HTTP redirect
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}
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

	w.Header().Set("HX-Location", "/profile")
	w.WriteHeader((http.StatusNoContent))
}

func (s *Server) fetchUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.userService.FetchAllUsers()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Users *[]models.Users
	}{
		Users: users,
	}

	err = s.tmpl.ExecuteTemplate(w, "userOverview", data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.userService.DeleteUser(userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// users, err := s.userService.FetchAllUsers()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// data := struct {
	// 	Users *[]models.Users
	// }{
	// 	Users: users,
	// }

	// s.tmpl.ExecuteTemplate(w, "userOverview", data)
	w.WriteHeader(http.StatusAccepted)
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
		"FirstName":        false,
		"LastName":         false,
		"Email":            false,
		"Password":         false,
		"PasswordTooShort": false,
		"InvalidEmail":     false,
	}
	hasErrors := false
	if user.FirstName == "" {
		errorMessages["FirstName"] = true
		hasErrors = true
	}
	if user.LastName == "" {
		errorMessages["LastName"] = true
		hasErrors = true
	}
	if user.Email == "" {
		errorMessages["Email"] = true
		hasErrors = true
	}
	if user.Password == "" {
		errorMessages["Password"] = true
		hasErrors = true
	}
	if len(user.Password) < 8 {
		errorMessages["PasswordTooShort"] = true
		hasErrors = true
	}
	if s.authService.IsValidEmail(user.Email) == false {
		errorMessages["InvalidEmail"] = true
		hasErrors = true
	}

	if hasErrors {
		errorFormData := map[string]interface{}{
			"ErrorMessages": errorMessages,
			"User":          user,
		}
		s.tmpl.ExecuteTemplate(w, "userForm", errorFormData)
		return
	}

	_, err = s.userService.CreateUser(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Location", "/users")
	w.WriteHeader((http.StatusNoContent))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
