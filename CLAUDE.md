# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LazyOrg is a terminal-based calendar and note-taking application written in Go. It provides a TUI interface for managing events, taking notes, and organizing your schedule with vim-style keybindings.

## Development Commands

### Building
```bash
go build                           # Build the application
go build -o lazyorg cmd/lazyorg/   # Build with specific output name
```

### Running
```bash
./lazyorg                          # Run the built binary
go run cmd/lazyorg/main.go         # Run directly from source
./lazyorg -backup /path/to/backup  # Backup database to specified location
```

### Testing
```bash
go test ./tests/...                # Run all tests
go test -v ./tests/...             # Run tests with verbose output
go test -v ./tests/ -run TestName  # Run specific test
go test -cover ./tests/...         # Run tests with coverage
go test -coverprofile=coverage.out ./tests/... && go tool cover -html=coverage.out  # Generate HTML coverage report
```

### Dependencies
```bash
go mod tidy                        # Clean up dependencies
go mod download                    # Download dependencies
```

## Architecture

### Core Components

**Main Application Flow:**
- `cmd/lazyorg/main.go` - Entry point, handles CLI flags, database initialization, and GUI setup
- `pkg/views/app-view.go` - Main application view that orchestrates all UI components
- `internal/ui/keybindings.go` - Global keybinding configuration

**Data Layer:**
- `internal/database/database.go` - SQLite database operations for events and notes
- `internal/eventmanager/eventmanager.go` - Event management with undo/redo functionality
- `internal/calendar/` - Calendar domain models (Event, Day, Week, Calendar)

**UI Layer:**
- `pkg/views/` - All UI view components (day-view, week-view, event-view, etc.)
- Uses `gocui` TUI framework for terminal interface
- Vim-style keybindings throughout

### Key Features
- **Undo/Redo System**: Full undo/redo support for event operations (add, delete, edit, bulk delete)
- **Event Management**: Create, edit, delete events with recurrence support
- **Search**: Search events within current week with `/`
- **Yank/Paste**: Copy events with `y` and paste with `p`
- **Color Coding**: Events are automatically colored or manually assigned colors
- **Notepad**: Integrated note-taking functionality

### Database Schema
- **events table**: id, name, description, location, time, duration, frequency, occurence, color
- **notes table**: id, content, updated_at
- Database location: `~/.local/share/lazyorg/data.db`
- Configuration: `~/.config/lazyorg/config.json`

### Testing Strategy
- Unit tests focus on EventManager functionality
- Tests use in-memory SQLite database (`:memory:`)
- Helper functions: `setupTestDB()`, `setupTestEventManager()`, `createTestEvent()`
- Test coverage includes undo/redo operations, bulk operations, and edge cases

## Key Patterns

### Event Management
- All event operations go through EventManager to ensure undo/redo tracking
- Direct database operations should be avoided in favor of EventManager methods
- Event colors are generated from name hash or manually set

### UI Architecture
- View components extend BaseView for common functionality
- Views are managed by AppView which handles state coordination
- Keybindings are defined per view but coordinated globally

### Error Handling
- Database operations return errors that should be properly handled
- UI operations gracefully handle and display errors to users
- Test setup includes proper cleanup with `defer` statements

## Recent Changes

The project recently added:
- Colored events with color picker
- Undo/redo functionality (`u` and `r` keys)
- Yank/paste system for events (`y`, `p`, `d` keys)
- Jump navigation (`g` key)
- Search within current week (`/` key)
- Previous/next event navigation within week (`w` and `b` keys)