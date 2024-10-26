package services

import (
	"context"
	"net/http"
)

type MiddleWareService struct {
	tokenService *TokenService
	userService  *UserService
}

func NewMiddleWareService(tokenService *TokenService, userService *UserService) *MiddleWareService {
	return &MiddleWareService{tokenService: tokenService, userService: userService}
}

func (m *MiddleWareService) CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		sessionToken := cookie.Value
		userId, err := m.tokenService.ValidateToken(sessionToken)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		user, err := m.userService.FetchUserById(userId)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=4") // Cache for 1 hour
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
