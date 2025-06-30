// Package main sets up the HTTP routes for the web application.
package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/config"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/handlers"
)

// routes sets up the application's HTTP routes and middleware.
func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	// Global middleware
	mux.Use(middleware.Recoverer) // Recover from panics
	mux.Use(NoSurf)               // CSRF protection
	mux.Use(SessionLoad)          // Load and save session data

	// Public routes
	mux.Get("/register", handlers.Repo.RegisterGet) // Registration page (GET)
	mux.Post("/register", handlers.Repo.RegisterPost) // Registration form submission (POST)
	
	mux.Get("/login", handlers.Repo.LoginGet) // Login page (GET)
	mux.Post("/login", handlers.Repo.LoginPost) // Login form submission (POST)

	mux.Get("/email-verification", handlers.Repo.EmailVerificationGet) // Email verification page (GET)
	mux.Post("/email-verification", handlers.Repo.EmailVerificationPost) // Email verification form (POST)
	
	// Protected routes (require authentication)
	mux.Route("/", func(mux chi.Router) {
		mux.Use(Auth) // Authentication middleware
		mux.Get("/dashboard", handlers.Repo.DashboardGet) // Dashboard page (GET)
	})

	// Serve static files from the ./static directory
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}