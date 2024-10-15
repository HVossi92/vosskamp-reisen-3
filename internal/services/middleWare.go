package services

import (
	"net/http"
)

type MiddleWareService struct {
	tokenService TokenService
}

func NewMiddleWareService(tokenService TokenService) *MiddleWareService {
	return &MiddleWareService{tokenService: tokenService}
}

func (m *MiddleWareService) CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		sessionToken := cookie.Value
		err = m.tokenService.ValidateToken(sessionToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
