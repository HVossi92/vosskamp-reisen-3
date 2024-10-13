package server

import (
	"fmt"
	"net/http"
)

func CheckSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		sessionToken := cookie.Value
		fmt.Println(sessionToken)
		next.ServeHTTP(w, r)
	})
}
