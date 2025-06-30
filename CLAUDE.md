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
./lazyorg -debug                   # Run with debug logging enabled
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

### Event Management Architecture

**CRITICAL: Always Use EventManager for Event Operations**

All event modifications MUST go through the EventManager to ensure:
- Undo/redo functionality works correctly
- Event overlap prevention is applied
- Consistent error handling and validation
- Proper state management

**EventManager Methods:**
- `AddEvent(event)` - Add new event with overlap checking
- `UpdateEvent(eventId, newEvent)` - Update existing event with overlap checking
- `DeleteEvent(eventId)` - Delete single event
- `DeleteEventsByName(name)` - Bulk delete events by name
- `Undo()` - Undo last operation
- `Redo()` - Redo last undone operation

**Database Direct Access:**
- Only use `database.Database` for READ operations (GetEventsByDate, GetEventById, etc.)
- NEVER use database directly for CREATE, UPDATE, DELETE operations
- EventManager handles all write operations internally

**Event Overlap Prevention:**
- EventManager automatically prevents overlapping events
- `CheckEventOverlap(event, excludeEventId...)` validates time conflicts
- Returns errors when operations would create overlaps
- Applies to add, edit, and paste operations

### Color System Architecture

**Available User Colors:**
- Red, Green, Yellow, Blue, Magenta, Cyan, White
- Accessible via color picker (`c` key) with shortcuts: r/g/y/b/m/c/w
- Events auto-generate colors from name hash if not manually set

**Special System Colors:**
- `calendar.ColorCustomPurple` - Reserved for current time highlighting
- Uses 256-color palette (color 93) for bright, distinct purple
- NOT available to users - system-only for time indication

**Color Implementation:**
- Colors stored as `gocui.Attribute` in database as integers
- `GetAvailableColors()` returns user-selectable colors only
- `GenerateColorFromName()` creates consistent auto-colors
- Color picker supports both single-letter shortcuts and full names

**Current Time Highlighting:**
- Purple hash characters (`###########`) when no event at current time
- Purple text color when event exists at current time
- Background color matches underlying content (event color or day background)
- Dynamically adapts to cursor state (black when cursor active, grey when inactive)

### UI Architecture
- View components extend BaseView for common functionality
- Views are managed by AppView which handles state coordination
- Keybindings are defined per view but coordinated globally

**CRITICAL: Avoid Sub-Keybindings at ALL COSTS**

Sub-keybindings (keybindings that exist outside of the globally active keybindings) should be avoided at ALL COSTS as they almost always break something in the gocui framework. This includes:
- Custom keybindings set directly on views using `g.SetKeybinding()`
- Modal or popup-specific keybindings that override global ones
- Context-sensitive keybinding changes

**Recommended Approach:**
- Use the Form component type (as used by add events, color picker, etc.) for all submenus and modal interactions
- Forms handle their own internal keybinding logic safely within the gocui-component framework
- This ensures consistent behavior and prevents keybinding conflicts

**Exception:**
- Only the error popup uses direct keybindings due to its simple nature, but even this should be minimized

### Error Handling
- Database operations return errors that should be properly handled
- UI operations gracefully handle and display errors to users
- Test setup includes proper cleanup with `defer` statements

### Debug Logging
LazyOrg includes comprehensive debug logging for troubleshooting time bounds and event overlap issues:

**Primary Debug Files:**
- **`/tmp/lazyorg_debug.txt`** - Main debug output from paste operations and overlap checking
- **`/tmp/lazyorg_getevents_debug.txt`** - Database query debugging from GetEventsByDate function

**Debug Sources:**
- **Paste Operations** (`pkg/views/app-view.go:608-680`): Calendar vs view date synchronization
- **Overlap Detection** (`internal/database/database.go:275-341`): Detailed time range comparisons
- **Database Queries** (`internal/database/database.go:130-185`): Event retrieval and date filtering

**Debug Contents:**
- Current view name vs Calendar.CurrentDay.Date synchronization
- Unix timestamps and timezone information for all time comparisons
- Step-by-step overlap detection logic with before/after comparisons
- Database query parameters and retrieved event timestamps
- Duration calculations and floating-point precision details

**Activation:**
- Enable with `./lazyorg -debug` command line flag
- When disabled, no debug files are created (normal operation)
- When enabled, creates debug files automatically during paste operations

**Common Use Cases:**
- Debugging event overlap detection failures
- Investigating calendar date synchronization issues
- Troubleshooting timezone-related time comparisons
- Verifying database query date ranges
- Analyzing floating-point duration calculation precision

## Recent Changes

The project recently added:
- **Event Overlap Prevention**: All event operations now prevent scheduling conflicts
- **Current Time Highlighting**: Purple indicators show current half-hour in today's column
- **Centralized Event Management**: All modifications route through EventManager
- **Enhanced Color System**: Custom purple for time highlighting, improved color management
- **Automatic View Refresh**: Current time highlighting updates automatically after all operations

Previous features:
- Colored events with color picker (`c` key)
- Undo/redo functionality (`u` and `r` keys) 
- Yank/paste system for events (`y`, `p`, `d` keys)
- Jump navigation (`g` key)
- Search within current week (`/` key)
- Previous/next event navigation within week (`w` and `b` keys)