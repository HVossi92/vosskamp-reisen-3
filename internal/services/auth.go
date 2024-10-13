package services

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/mail"
	"time"
)

type AuthService struct {
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Authorize(email string) error {
	return nil
}

func (s *AuthService) IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (s *AuthService) GenerateCookie(name string, isHttpOnly bool) (*http.Cookie, error) {
	token, err := s.generateToken()
	if err != nil {
		return nil, err
	}
	cookie := http.Cookie{
		Name:     name,
		Value:    token,
		Expires:  time.Now().Add(42 * time.Hour),
		HttpOnly: isHttpOnly,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}
	return &cookie, nil
}

func (s *AuthService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
