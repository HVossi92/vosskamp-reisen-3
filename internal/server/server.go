package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	postService       *services.PostService
	tmpl              *template.Template
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()

	tokenService := services.NewTokenService(db)
	userService := services.NewUserService(db)
	NewServer := &Server{
		port:              port,
		db:                db,
		tmpl:              getTemplates(),
		userService:       userService,
		authService:       services.NewAuthService(),
		tokenService:      tokenService,
		middleWareService: services.NewMiddleWareService(tokenService, userService),
		postService:       services.NewPostService(db),
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

func getTemplates() *template.Template {
	tmplDir := "internal/templates"
	tmpl := template.New("")

	// Walk through all directories and find HTML files
	err := filepath.Walk(tmplDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only process .html files
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			var err error
			tmpl, err = tmpl.ParseFiles(path)
			if err != nil {
				return fmt.Errorf("parsing template %s: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking templates: %v", err)
	}

	tmpl = template.Must(tmpl, err)
	return tmpl
}
