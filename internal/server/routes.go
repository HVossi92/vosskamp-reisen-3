package server

import (
	"math"
	"net/http"
	"strconv"
	"vosskamp-reisen-3/internal/helpers"
	"vosskamp-reisen-3/internal/models"
	"vosskamp-reisen-3/internal/structs"
	"vosskamp-reisen-3/internal/templates/user"

	"github.com/a-h/templ"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.homeFormHandler)
	mux.HandleFunc("GET /posts", s.homePostsHandler)
	mux.HandleFunc("GET /login", s.fetchLoginPageHandler)
	mux.HandleFunc("GET /post/{id}", s.postHandler)
	mux.HandleFunc("POST /login", s.loginHandler)
	mux.Handle("DELETE /logout", s.middleWareService.CheckSession(http.HandlerFunc(s.logoutHandler)))
	mux.HandleFunc("GET /affiliate", s.affiliateHandler)
	mux.HandleFunc("GET /about", s.aboutHandler)
	mux.HandleFunc("GET /contact", s.contactHandler)

	s.RegisterPostRoutes(mux)
	s.RegisterEmailRoutes(mux)
	s.RegisterUserRoutes(mux)

	staticServer := http.FileServer(http.Dir("./internal/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", staticServer))

	return addCacheControlMiddleware(mux)
}

func addCacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=60")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) homeFormHandler(w http.ResponseWriter, r *http.Request) {
	page, limit := helpers.GetPagination(r)
	posts, total, err := s.postService.FetchPaginatedPosts(page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	data := structs.HomePostsData{
		Posts: posts,
		PaginatedData: structs.PaginatedData{
			CurrentPage:      page,
			TotalPages:       totalPages,
			Limit:            limit,
			PreviousPage:     page - 1,
			NextPage:         page + 1,
			PageButtonsRange: makeRange(1, totalPages),
		},
	}
	handler := templ.Handler(user.Home(data))
	handler.ServeHTTP(w, r)
}

func (s *Server) homePostsHandler(w http.ResponseWriter, r *http.Request) {
	page, limit := helpers.GetPagination(r)
	posts, total, err := s.postService.FetchPaginatedPosts(page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	data := structs.HomePostsData{
		Posts: posts,
		PaginatedData: structs.PaginatedData{
			CurrentPage:      page,
			TotalPages:       totalPages,
			Limit:            limit,
			PreviousPage:     page - 1,
			NextPage:         page + 1,
			PageButtonsRange: makeRange(1, totalPages),
		},
	}
	handler := templ.Handler(user.Blog(data))
	handler.ServeHTTP(w, r)
}

func (s *Server) postHandler(w http.ResponseWriter, r *http.Request) {
	postIdString := r.PathValue(("id"))
	postId, err := strconv.Atoi(postIdString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post, err := s.postService.FetchPostById(postId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// html, err := helpers.ConvertQuillToHtml(post.Body)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Cache-Control", "public, max-age=3600")
	handler := templ.Handler(user.Post(*post))
	handler.ServeHTTP(w, r)
}

func (s *Server) affiliateHandler(w http.ResponseWriter, r *http.Request) {
	handler := templ.Handler(user.Affiliate())
	handler.ServeHTTP(w, r)
}

func (s *Server) aboutHandler(w http.ResponseWriter, r *http.Request) {
	handler := templ.Handler(user.About())
	handler.ServeHTTP(w, r)
}

func (s *Server) contactHandler(w http.ResponseWriter, r *http.Request) {
	handler := templ.Handler(user.Contact())
	handler.ServeHTTP(w, r)
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

	w.Header().Set("HX-Location", "/admin/posts")
	w.WriteHeader(http.StatusNoContent)

}
