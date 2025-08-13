package main

import (
	"log"
	"os"

	"board-game-library/internal/app"
	_ "board-game-library/docs" // This line is needed for go-swagger to find your docs!
)

// @title Board Game Library API
// @version 1.0
// @description API pour la gestion d'une bibliothèque de jeux de société
// @termsOfService http://swagger.io/terms/

// @contact.name Support API
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

func main() {
	// Create and initialize application
	application, err := app.New()
	if err != nil {
		log.Printf("Failed to create application: %v", err)
		os.Exit(1)
	}

	// Initialize all components
	if err := application.Initialize(); err != nil {
		log.Printf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// Run application (blocks until shutdown)
	if err := application.Run(); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}
}