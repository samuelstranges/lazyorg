# LazyOrg Test Suite

This directory contains unit tests for the LazyOrg calendar application.

## Running Tests

To run all tests:
```bash
go test ./tests/...
```

To run tests with verbose output:
```bash
go test -v ./tests/...
```

To run a specific test:
```bash
go test -v ./tests/ -run TestBulkDeleteUndoRedo
```

## Test Coverage

To run tests with coverage:
```bash
go test -cover ./tests/...
```

To generate detailed coverage report:
```bash
go test -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out
```

## Test Structure

### `eventmanager_test.go`
Contains unit tests for the EventManager functionality including:
- **TestBulkDeleteUndoRedo**: Tests bulk delete operations and their undo/redo functionality
- **TestAddEventUndoRedo**: Tests adding events and undo/redo operations
- **TestDeleteEventUndoRedo**: Tests deleting individual events and undo/redo operations  
- **TestUndoRedoStackLimits**: Tests undo/redo stack behavior and limits

### Helper Functions
- `setupTestDB()`: Creates an in-memory SQLite database for testing
- `setupTestEventManager()`: Creates an EventManager with test database
- `createTestEvent()`: Helper to create test events with specified parameters

## Adding New Tests

When adding new tests:

1. **Follow Go testing conventions**: Test functions should start with `Test`
2. **Use table-driven tests** when testing multiple scenarios
3. **Clean up resources**: Always defer database cleanup with `defer db.CloseDatabase()`
4. **Use descriptive names**: Test names should clearly indicate what they're testing
5. **Test edge cases**: Include tests for error conditions and boundary cases

## Test Database

All tests use an in-memory SQLite database (`:memory:`) which:
- Provides isolation between tests
- Runs fast without file I/O
- Gets cleaned up automatically when the test completes
- Matches the production database schema exactly

## Future Test Areas

Consider adding tests for:
- Calendar navigation functionality
- Event validation and input handling
- Database migration and schema changes
- UI component behavior (when possible)
- Performance testing for large datasets
- Concurrent access scenarios