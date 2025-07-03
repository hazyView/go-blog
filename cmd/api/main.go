package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blog-api/internal/config"
	"blog-api/internal/database"
	"blog-api/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure structured logging
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"})

	// Load configuration
	cfg := config.Load()

	// Set log level
	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Msg("Starting Blog API server...")

	// Initialize database connection
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database connection")
		}
	}()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	postHandler := handlers.NewPostHandler(db)
	healthHandler := handlers.NewHealthHandler(db)
	webHandler := handlers.NewWebHandler()

	// Setup router
	router := setupRouter(userHandler, postHandler, healthHandler, webHandler)

	// Configure HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("port", cfg.Port).Msg("Server starting on port " + cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Server shutting down...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		log.Info().Msg("Server gracefully stopped")
	}
}

// setupRouter configures and returns the HTTP router with all routes and middleware
func setupRouter(userHandler *handlers.UserHandler, postHandler *handlers.PostHandler, healthHandler *handlers.HealthHandler, webHandler *handlers.WebHandler) *mux.Router {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(handlers.LoggingMiddleware)
	router.Use(handlers.PanicRecoveryMiddleware)
	router.Use(handlers.CORSMiddleware)
	router.Use(handlers.SecurityHeadersMiddleware)
	router.Use(handlers.TimeoutMiddleware(30 * time.Second))

	// Serve static files
	staticDir := http.Dir("web/static/")
	staticHandler := http.StripPrefix("/static/", http.FileServer(staticDir))
	router.PathPrefix("/static/").Handler(staticHandler)

	// Web interface routes
	router.HandleFunc("/", webHandler.Index).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// User routes
	api.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	api.HandleFunc("/users/{id:[0-9]+}", userHandler.GetUser).Methods("GET")
	api.HandleFunc("/users/{id:[0-9]+}", userHandler.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id:[0-9]+}", userHandler.DeleteUser).Methods("DELETE")

	// Post routes
	api.HandleFunc("/posts", postHandler.CreatePost).Methods("POST")
	api.HandleFunc("/posts", postHandler.GetAllPosts).Methods("GET")
	api.HandleFunc("/posts/{id:[0-9]+}", postHandler.GetPost).Methods("GET")
	api.HandleFunc("/posts/{id:[0-9]+}", postHandler.UpdatePost).Methods("PUT")
	api.HandleFunc("/posts/{id:[0-9]+}", postHandler.DeletePost).Methods("DELETE")

	// API Health check
	api.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")

	// 404 handler
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not Found","message":"The requested resource was not found","code":404}`))
	})

	// 405 handler
	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error":"Method Not Allowed","message":"The request method is not allowed for this resource","code":405}`))
	})

	return router
}
