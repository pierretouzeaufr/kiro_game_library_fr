package handler

import (
	"net/http"
	"os"
	"sync"

	"board-game-library/internal/app"
)

var (
	application *app.App
	once        sync.Once
	initErr     error
)

// Handler is the main entry point for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize the application once
	once.Do(func() {
		// Set environment variables for Vercel
		os.Setenv("SERVER_HOST", "0.0.0.0")
		os.Setenv("SERVER_PORT", "8080")
		os.Setenv("DATABASE_PATH", "/tmp/library.db")
		
		var err error
		application, err = app.New()
		if err != nil {
			initErr = err
			return
		}
		
		if err := application.Initialize(); err != nil {
			initErr = err
			return
		}
	})

	if initErr != nil {
		http.Error(w, "Failed to initialize application: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	// Get the router from the application and serve the request
	router := application.GetRouter()
	if router != nil {
		router.ServeHTTP(w, r)
	} else {
		http.Error(w, "Router not initialized", http.StatusInternalServerError)
	}
}