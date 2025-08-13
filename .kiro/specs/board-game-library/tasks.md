
# Implementation Plan

- [x] 1. Set up project structure and dependencies
  - Initialize Go module with proper project structure
  - Add required dependencies (Gin, SQLite driver, HTMX, Tailwind CSS)
  - Create directory structure for models, repositories, services, handlers, and templates
  - _Requirements: 5.1_

- [x] 2. Implement core data models and validation
  - Create User, Game, Borrowing, and Alert struct definitions with JSON and database tags
  - Implement validation functions for each model (email format, required fields, etc.)
  - Write unit tests for model validation logic
  - _Requirements: 1.1, 2.1, 3.1, 6.1_

- [x] 3. Set up database infrastructure
- [x] 3.1 Create database connection and initialization
  - Implement SQLite connection management with proper error handling
  - Create database initialization function that sets up tables and indexes
  - Write migration system for schema updates
  - _Requirements: 5.1, 5.3_

- [x] 3.2 Implement repository interfaces and SQLite implementations
  - Create repository interfaces for User, Game, Borrowing, and Alert
  - Implement SQLite-based repositories with prepared statements
  - Write comprehensive unit tests for all repository methods using in-memory SQLite
  - _Requirements: 5.4, 1.1, 2.1, 3.1_

- [x] 4. Implement business logic services
- [x] 4.1 Create UserService with borrowing eligibility logic
  - Implement user registration, retrieval, and borrowing history methods
  - Add logic to check if user can borrow (no overdue items)
  - Write unit tests with mocked repository dependencies
  - _Requirements: 1.1, 1.2, 1.5, 6.1_

- [x] 4.2 Create GameService with inventory management
  - Implement game creation, retrieval, search, and availability tracking
  - Add methods for updating game status during borrowing/returning
  - Write unit tests for all game management operations
  - _Requirements: 2.1, 2.2, 2.5_

- [x] 4.3 Create BorrowingService with workflow logic
  - Implement borrowing workflow with availability checks and due date calculation
  - Add return processing with automatic status updates
  - Create methods for identifying overdue items
  - Write unit tests for complete borrowing/return cycles
  - _Requirements: 6.1, 6.2, 6.3, 2.3, 2.4_

- [x] 4.4 Create AlertService with notification logic
  - Implement overdue alert generation based on due dates
  - Add reminder alert creation for items due within 2 days
  - Create methods for alert management (mark as read, group by user)
  - Write unit tests for alert generation and management
  - _Requirements: 3.1, 3.2, 3.4, 3.5_

- [x] 5. Create web layer with Gin handlers
- [x] 5.1 Implement user management endpoints
  - Create handlers for user registration, listing, and profile viewing
  - Add endpoints for viewing user borrowing history and current loans
  - Implement proper HTTP status codes and error responses
  - Write HTTP handler tests using httptest package
  - _Requirements: 1.1, 1.2, 4.3_

- [x] 5.2 Implement game management endpoints
  - Create handlers for adding, editing, and listing games
  - Add search endpoint with query parameter support
  - Implement game detail view with borrowing history
  - Write HTTP tests for all game management endpoints
  - _Requirements: 2.1, 2.2, 2.5, 4.3_

- [x] 5.3 Implement borrowing workflow endpoints
  - Create handlers for borrowing and returning games
  - Add endpoints for extending due dates and viewing borrowing details
  - Implement proper validation and error handling for workflow operations
  - Write HTTP tests for complete borrowing workflows
  - _Requirements: 6.1, 6.2, 6.4, 4.3_

- [x] 5.4 Implement alert management endpoints
  - Create handlers for viewing alerts and marking them as read
  - Add dashboard endpoint showing overdue items and upcoming due dates
  - Implement alert filtering and grouping by user
  - Write HTTP tests for alert management functionality
  - _Requirements: 3.1, 3.2, 3.3, 3.5, 4.3_

- [x] 6. Create responsive HTML templates
- [x] 6.1 Design base template with navigation and responsive layout
  - Create base HTML template with Tailwind CSS for responsive design
  - Implement navigation menu for users, games, and alerts sections
  - Add responsive breakpoints for mobile and desktop views
  - _Requirements: 4.1, 4.2, 4.5_

- [x] 6.2 Create user management templates
  - Build templates for user listing, registration form, and user profile pages
  - Add borrowing history display with sortable tables
  - Implement user search and filtering interface
  - _Requirements: 4.1, 4.4, 1.2_

- [x] 6.3 Create game management templates
  - Build templates for game listing, add/edit forms, and game detail pages
  - Add search interface with real-time filtering
  - Implement availability status indicators and borrowing history display
  - _Requirements: 4.1, 4.4, 2.2, 2.5_

- [x] 6.4 Create borrowing workflow templates
  - Build templates for borrowing form, return processing, and borrowing history
  - Add due date management interface with extension capabilities
  - Implement confirmation dialogs for borrowing and return actions
  - _Requirements: 4.1, 4.3, 6.1, 6.2_

- [x] 6.5 Create alert and dashboard templates
  - Build dashboard template showing overdue items and upcoming due dates
  - Add alert notification interface with mark-as-read functionality
  - Implement alert grouping and filtering by type and user
  - _Requirements: 4.1, 4.4, 3.1, 3.2, 3.5_

- [x] 7. Add HTMX interactivity for dynamic features
- [x] 7.1 Implement dynamic search and filtering
  - Add HTMX-powered real-time search for games and users
  - Implement dynamic filtering without page reloads
  - Create partial template updates for search results
  - _Requirements: 4.3, 2.5_

- [x] 7.2 Add dynamic borrowing and return actions
  - Implement HTMX-powered borrowing workflow with instant feedback
  - Add dynamic return processing with status updates
  - Create real-time availability status updates
  - _Requirements: 4.3, 6.1, 6.4_

- [x] 7.3 Implement dynamic alert management
  - Add HTMX-powered alert marking and dismissal
  - Implement real-time alert count updates
  - Create dynamic dashboard updates for overdue items
  - _Requirements: 4.3, 3.2, 3.3_

- [x] 8. Create background job system for alerts
  - Implement scheduled job system for generating overdue and reminder alerts
  - Add daily alert generation process that runs automatically
  - Create job logging and error handling for background processes
  - Write tests for alert generation scheduling and execution
  - _Requirements: 3.1, 3.4_

- [x] 9. Add application configuration and startup
  - Create configuration system for database path, server port, and alert settings
  - Implement graceful application startup with database initialization
  - Add proper logging configuration and structured logging throughout the application
  - Create application shutdown handling with database connection cleanup
  - _Requirements: 5.1, 5.3_

- [x] 10. Write integration tests for complete workflows
  - Create end-to-end tests for user registration and game borrowing workflows
  - Add integration tests for alert generation and overdue item processing
  - Implement database integration tests with real SQLite database
  - Write performance tests for concurrent user access and large datasets
  - _Requirements: 1.1, 2.1, 3.1, 6.1_

- [x] 11. Create build system and deployment preparation
  - Set up Go build configuration with embedded static assets
  - Create Makefile or build scripts for cross-platform compilation
  - Add database migration system for production deployments
  - Implement single-binary deployment with all dependencies included
  - _Requirements: 5.1_