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
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
