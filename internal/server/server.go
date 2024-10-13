package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"vosskamp-reisen-3/internal/database"
	"vosskamp-reisen-3/internal/services"
)

type Server struct {
	port        int
	db          database.Service
	userService *services.UserService
	tmpl        *template.Template
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()
	tmpl, err := template.ParseGlob("internal/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}

	NewServer := &Server{
		port:        port,
		db:          db,
		tmpl:        tmpl,
		userService: services.NewUserService(db),
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
