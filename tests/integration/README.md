# Integration Tests

This directory contains comprehensive integration tests for the Board Game Library Management System.

## Test Categories

### 1. Workflow Tests (`workflows_test.go`)
- **End-to-end user registration and game borrowing workflows**
- **Multiple users borrowing different games**
- **Overdue item processing and alert generation**
- **Due date extension functionality**
- **Error handling scenarios**

Tests complete user journeys from registration through borrowing, returning, and alert processing.

### 2. Database Integration Tests (`database_test.go`)
- **Database schema integrity verification**
- **Foreign key constraint testing**
- **Transaction integrity and rollback testing**
- **Database connection pool testing**
- **Data persistence across connections**
- **Migration system testing**
- **Performance testing with large datasets**

Tests the database layer with real SQLite databases to ensure data integrity and performance.

### 3. Performance Tests (`performance_test.go`)
- **Concurrent user access testing**
- **Large dataset performance testing**
- **Memory usage monitoring**
- **System stress testing**
- **Race condition testing**

Tests system performance under various load conditions and concurrent access scenarios.

### 4. Alert Integration Tests (`alert_integration_test.go`)
- **Overdue alert generation and processing**
- **Reminder alert generation**
- **Multiple user alert scenarios**
- **Alert performance with large datasets**
- **Alert cleanup and management**

Tests the complete alert system workflow including generation, processing, and management.

## Running the Tests

### Run All Integration Tests
```bash
go test ./tests/integration/... -v
```

### Run Specific Test Categories
```bash
# Workflow tests only
go test ./tests/integration/ -run TestUserRegistrationAndBorrowingWorkflow -v

# Database tests only
go test ./tests/integration/ -run TestDatabaseIntegration -v

# Performance tests only (may take longer)
go test ./tests/integration/ -run TestConcurrentUserAccess -v

# Alert tests only
go test ./tests/integration/ -run TestAlertGenerationIntegration -v
```

### Skip Performance Tests (for faster runs)
```bash
go test ./tests/integration/... -short -v
```

### Run with Race Detection
```bash
go test ./tests/integration/... -race -v
```

## Test Requirements

These integration tests require:
- Go 1.21+
- SQLite3 support
- testify/assert and testify/require packages
- All application dependencies

## Test Data

Tests use:
- In-memory SQLite databases for most tests
- Temporary file-based databases for persistence tests
- Generated test data (users, games, borrowings)
- Concurrent access scenarios
- Large dataset scenarios (1000+ records)

## Performance Benchmarks

The tests include performance assertions:
- User registration: < 30 seconds for 1000 users
- Game creation: < 60 seconds for 2000 games
- Query operations: < 2 seconds for large datasets
- Search operations: < 500ms
- Alert generation: < 5 seconds for 100+ overdue items
- Memory usage: < 100MB for large datasets

## Coverage

These integration tests cover:
- All major user workflows (Requirements 1.1, 6.1)
- Game management operations (Requirements 2.1)
- Alert system functionality (Requirements 3.1)
- Database integrity and performance (Requirements 5.1-5.4)
- Error handling and edge cases
- Concurrent access scenarios
- Performance under load

## Troubleshooting

If tests fail:
1. Ensure all dependencies are installed: `go mod tidy`
2. Check that SQLite3 is properly installed
3. Verify database permissions in test directories
4. For performance test failures, check system resources
5. For race condition failures, run with `-race` flag for details