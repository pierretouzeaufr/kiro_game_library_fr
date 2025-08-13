# Board Game Library Management System

A Go-based web application for managing a board game library with user borrowing capabilities.

## Features

- User management and registration
- Game inventory management
- Borrowing and return workflow
- Overdue alerts and notifications
- Responsive web interface with HTMX
- SQLite database for local storage

## Project Structure

```
board-game-library/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── models/          # Data models
│   ├── repositories/    # Data access layer
│   ├── services/        # Business logic
│   └── handlers/        # HTTP handlers
├── pkg/
│   └── database/        # Database utilities
├── web/
│   ├── templates/       # HTML templates
│   └── static/          # CSS, JS, images
└── go.mod
```

## Dependencies

- **Gin**: Web framework for HTTP routing and middleware
- **SQLite3**: Database driver for local data storage
- **HTMX**: Frontend interactivity without complex JavaScript
- **Tailwind CSS**: Utility-first CSS framework

## Getting Started

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Run the application: `go run cmd/server/main.go`
4. Open http://localhost:8080 in your browser

## Development

The application follows a clean architecture pattern with clear separation of concerns:

- **Models**: Define data structures and validation
- **Repositories**: Handle data persistence and retrieval
- **Services**: Implement business logic and workflows
- **Handlers**: Process HTTP requests and responses

## Requirements

- Go 1.21 or higher
- SQLite3 (included with go-sqlite3 driver)