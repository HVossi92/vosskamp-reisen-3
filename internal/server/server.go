package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"vosskamp-reisen-3/internal/database"
	"vosskamp-reisen-3/internal/services"
)

type Server struct {
	port              int
	db                database.Service
	userService       *services.UserService
	authService       *services.AuthService
	tokenService      *services.TokenService
	middleWareService *services.MiddleWareService
	tmpl              *template.Template
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()
	tmpl, err := template.ParseGlob("internal/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	tokenService := services.NewTokenService(db)
	userService := services.NewUserService(db)
	NewServer := &Server{
		port:              port,
		db:                db,
		tmpl:              tmpl,
		userService:       userService,
		authService:       services.NewAuthService(),
		tokenService:      tokenService,
		middleWareService: services.NewMiddleWareService(tokenService, userService),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
