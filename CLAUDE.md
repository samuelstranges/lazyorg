# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## Project Overview

Chronos is a sophisticated terminal-based calendar and event management system
written in Go. It provides a TUI interface for managing events with advanced
features like overlap prevention, undo/redo operations, and vim-style
keybindings.

## Development Commands

### Building

```bash
go build cmd/chronos/main.go       # Build the application (outputs ./main)
go build -o chronos cmd/chronos/main.go  # Build with specific output name
```

### Running

```bash
./main                             # Run the built binary (default name)
./chronos                          # Run the built binary (custom name)
go run cmd/chronos/main.go         # Run directly from source
./chronos -backup /path/to/backup  # Backup database to specified location
./chronos -debug                   # Run with debug logging enabled
./chronos -db /path/to/custom.db   # Use custom database location
./chronos --help                   # Show all available command-line options
```

### Command-Line Options

- **`-db <path>`** - Specify custom database file location  
  Default: `~/.local/share/chronos/data.db`
- **`-backup <path>`** - Backup database to specified location and exit
- **`-debug`** - Enable debug logging to `/tmp/chronos_debug.txt` and `/tmp/chronos_getevents_debug.txt`
- **`--help`** - Show all available command-line options

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

- `cmd/chronos/main.go` - Entry point, handles CLI flags, database
  initialization, and GUI setup
- `pkg/views/app-view.go` - Main application view that orchestrates all UI
  components
- `internal/ui/keybindings.go` - Global keybinding configuration

**Data Layer:**

- `internal/database/database.go` - SQLite database operations for events and
  notes
- `internal/eventmanager/eventmanager.go` - Event management with undo/redo
  functionality
- `internal/calendar/` - Calendar domain models (Event, Day, Week, Calendar)

**UI Layer:**

- `pkg/views/` - All UI view components (day-view, week-view, event-view, etc.)
- Uses `gocui` TUI framework for terminal interface
- Vim-style keybindings throughout

### Key Features

- **Undo/Redo System**: Full undo/redo support for event operations (add,
  delete, edit, bulk delete)
- **Event Management**: Create, edit, delete events with recurrence support
- **Search**: Search events within current week with `/`
- **Yank/Paste**: Copy events with `y` and paste with `p`
- **Color Coding**: Events are automatically colored or manually assigned colors
- **Notepad**: Integrated note-taking functionality

### Database Schema

- **events table**: id, name, description, location, time, duration, frequency,
  occurence, color
- **notes table**: id, content, updated_at
- Database location: `~/.local/share/chronos/data.db` (default) or custom path
  via `-db` flag or config file
- Configuration: `~/.config/chronos/config.json`

### Database Configuration

- **Command line**: `./chronos -db /path/to/custom.db`
- **Config file**: Set `database_path` in `~/.config/chronos/config.json`
- **Priority**: Command line flag > config file > default location

### Testing Strategy

- Unit tests focus on EventManager functionality
- Tests use in-memory SQLite database (`:memory:`)
- Helper functions: `setupTestDB()`, `setupTestEventManager()`,
  `createTestEvent()`
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

- Only use `database.Database` for READ operations (GetEventsByDate,
  GetEventById, etc.)
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
- Dynamically adapts to cursor state (black when cursor active, grey when
  inactive)

### UI Architecture

- View components extend BaseView for common functionality
- Views are managed by AppView which handles state coordination
- Keybindings are defined per view but coordinated globally

**CRITICAL: Avoid Sub-Keybindings at ALL COSTS**

Sub-keybindings (keybindings that exist outside of the globally active
keybindings) should be avoided at ALL COSTS as they almost always break
something in the gocui framework. This includes:

- Custom keybindings set directly on views using `g.SetKeybinding()`
- Modal or popup-specific keybindings that override global ones
- Context-sensitive keybinding changes

**Recommended Approach:**

- Use the Form component type (as used by add events, color picker, etc.) for
  all submenus and modal interactions
- Forms handle their own internal keybinding logic safely within the
  gocui-component framework
- This ensures consistent behavior and prevents keybinding conflicts

**Exception:**

- Only the error popup uses direct keybindings due to its simple nature, but
  even this should be minimized

### Error Handling

- Database operations return errors that should be properly handled
- UI operations gracefully handle and display errors to users
- Test setup includes proper cleanup with `defer` statements

### Enhanced Search System

The search functionality (`/` key) provides powerful filtering capabilities for finding events across the entire database.

**Search Form Fields:**

- **Query**: Text search across event names, descriptions, and locations (case-insensitive, partial matching)
- **From Date**: Optional start date filter (YYYYMMDD format or 't' for today)
- **To Date**: Optional end date filter (YYYYMMDD format or 't' for today)

**Date Shortcut:**

- Use **'t'** in either date field as a shortcut for today's date
- Examples: `From Date: t` finds events from today onwards, `To Date: t` finds events up to today

**Usage Examples:**

- `Query: "meeting"` → Find all events containing "meeting"
- `From Date: t` → Find all events from today onwards  
- `To Date: 20241231` → Find all events up to Dec 31st
- `From Date: 20241225, To Date: 20241231` → Find events in that date range
- `Query: "lunch", From Date: t` → Find lunch events from today onwards
- `From Date: t, To Date: t` → Find only today's events

**Implementation Details:**

- Database function: `SearchEventsWithFilters(criteria SearchCriteria)`
- Text search uses SQL LIKE with case-insensitive matching
- Date filters automatically use 00:00-23:59 time ranges for full-day coverage
- 't' shortcut resolved at query time using `time.Now().Format("20060102")`
- Navigation works with search results (`w`/`b` keys, search result counter)

**Form Component Note:**

- Field name "Query" (not "Search") to avoid naming conflict with "Search" button
- Validation messages document the 't' shortcut in tooltips

### Known Issues and Technical Debt

#### Mixed Timezone Storage in Database

**Problem**: Historical events in the database have inconsistent timezone
information:

- Events created via **Add Event popup** (before fix): Stored in UTC timezone
  due to `time.Parse()`
- Events created via **paste operations**: Stored in local timezone due to
  `time.Date()` with location
- This causes overlap detection failures when comparing events from different
  creation methods

**Current Status**:

- **Fixed for new events**: All creation methods now use
  `time.ParseInLocation()` with local timezone
- **Comparison logic**: `CheckEventOverlap()` normalizes all times to same
  timezone before comparison
- **Database**: Still contains mixed UTC/Local timezone events (not migrated)

**Potential Future Issues**:

- Time-based queries may return unexpected results across different weeks/months
- Backup/restore operations might introduce timezone inconsistencies
- Third-party integrations could be confused by mixed timezone data

**Mitigation**: Always normalize timezones during time comparisons, never assume
consistent timezone storage

**Future Work**: Consider implementing a one-time migration script to normalize
all historical events to local timezone

#### Navigation Bug Due to Mixed Timezone Storage

**Problem**: The `w` and `b` navigation keys (JumpToNextEvent/JumpToPrevEvent)
were navigating to incorrect events due to mixed timezone storage causing
incorrect chronological sorting:

- UTC-stored events were being converted to Local time (+10 hours in Australia)
  during sorting
- This caused events like "Morning at 6:00 AM" (stored as UTC) to appear as
  "4:00 PM" in sorting comparisons
- Navigation would jump to events seemingly out of chronological order

**Root Cause**:

- Historical events stored inconsistently (UTC vs Local timezone) as described
  in "Mixed Timezone Storage" above
- `getAllEventsFromWeek()` function was converting UTC events to Local time
  during sorting
- This caused chronologically incorrect event ordering for navigation algorithms

**Technical Debt Impact**:

- **Event Navigation**: `w`/`b` keys jumped to wrong events, breaking user
  workflow
- **Sorting Logic**: Event chronological ordering was corrupted across the
  entire week view
- **Data Integrity**: Mixed timezone storage affects multiple system components
  beyond just navigation

**Fix Implementation** (`pkg/views/app-view.go`):

```go
// TEMPORARY FIX: ALL events stored in UTC are wrong - use original time instead of converting
if event.Time.Location().String() == "UTC" {
    eventTime = time.Date(event.Time.Year(), event.Time.Month(), event.Time.Day(),
                         event.Time.Hour(), event.Time.Minute(), event.Time.Second(),
                         event.Time.Nanosecond(), time.Local)
} else {
    eventTime = event.Time.In(time.Local)
}
```

**Current Status**:

- **Navigation Fixed**: `w` and `b` keys now navigate correctly in chronological
  order
- **Temporary Solution**: UTC events treated as Local time rather than converted
- **Debug Logging Added**: Comprehensive navigation debugging in
  `/tmp/lazyorg_nav_debug.txt`
- **Database Still Corrupt**: Historical UTC events remain incorrectly stored

**Future Work**: This is the SAME underlying issue as overlap detection -
requires database migration to properly fix

#### Form Component Field/Button Name Conflicts

**Problem**: The gocui-component form library has naming conflicts when an input field and button share the same name.

**Issue**: Originally, the search form had an input field named "Search" and a button named "Search", which caused the form to render incorrectly:
- The "Search" button would not appear
- The input field text would render in the button area (next to "Cancel")
- Form layout would be completely broken

**Root Cause**: The form component library internally confuses fields and buttons with identical names, causing rendering conflicts.

**Solution**: Renamed the search input field from "Search" to "Query" to avoid the naming conflict.

**Prevention**: When creating forms, ensure input field names and button names are always different:
- ✅ Good: Field "Query" + Button "Search"
- ✅ Good: Field "Name" + Button "Add"  
- ❌ Bad: Field "Search" + Button "Search"
- ❌ Bad: Field "Edit" + Button "Edit"

**Current Status**: 
- **Fixed**: Search form now works correctly with "Query" field and "Search" button
- **Enhanced**: Added date filtering fields ("From Date", "To Date") for advanced search
- **Tested**: All form functionality verified with comprehensive unit tests

### Debug Logging

Chronos includes comprehensive debug logging for troubleshooting time bounds and
event overlap issues:

**Primary Debug Files:**

- **`/tmp/chronos_debug.txt`** - Main debug output from paste operations and
  overlap checking
- **`/tmp/chronos_getevents_debug.txt`** - Database query debugging from
  GetEventsByDate function
- **`/tmp/chronos_nav_debug.txt`** - Navigation debugging from
  JumpToNextEvent/JumpToPrevEvent functions

**Debug Sources:**

- **Paste Operations** (`pkg/views/app-view.go:608-680`): Calendar vs view date
  synchronization
- **Overlap Detection** (`internal/database/database.go:275-341`): Detailed time
  range comparisons
- **Database Queries** (`internal/database/database.go:130-185`): Event
  retrieval and date filtering
- **Navigation Functions**
  (`pkg/views/app-view.go:JumpToNextEvent/JumpToPrevEvent`): Event sorting and
  chronological navigation

**Debug Contents:**

- Current view name vs Calendar.CurrentDay.Date synchronization
- Unix timestamps and timezone information for all time comparisons
- Step-by-step overlap detection logic with before/after comparisons
- Database query parameters and retrieved event timestamps
- Duration calculations and floating-point precision details
- Event sorting algorithms with before/after chronological ordering
- Navigation logic with wrap-around behavior and event selection

**Activation:**

- Enable with `./chronos -debug` command line flag
- When disabled, no debug files are created (normal operation)
- When enabled, creates debug files automatically during paste operations and
  navigation

**Common Use Cases:**

- Debugging event overlap detection failures
- Investigating calendar date synchronization issues
- Troubleshooting timezone-related time comparisons
- Verifying database query date ranges
- Analyzing floating-point duration calculation precision
- Debugging navigation issues with `w` and `b` keys jumping to wrong events
- Investigating event chronological sorting problems due to mixed timezone
  storage

## Recent Changes

The project recently added:

- **Unified Date Format**: All date fields now use YYYYMMDD format (no dashes) for consistency across goto, add/edit, and search forms
- **Consistent 't' Usage**: Changed "Jump to today" keybinding from 'T' to 't' for consistency with 't' shortcut in date fields
- **Symmetric Day Boundary Navigation**: Fixed time navigation to work symmetrically - going up from 00:00 now goes to 23:30 of previous day, matching the behavior of going down from 23:30 to 00:00 of next day
- **Event Overlap Prevention**: All event operations now prevent scheduling
  conflicts
- **Current Time Highlighting**: Purple indicators show current half-hour in
  today's column
- **Centralized Event Management**: All modifications route through EventManager
- **Enhanced Color System**: Custom purple for time highlighting, improved color
  management
- **Automatic View Refresh**: Current time highlighting updates automatically
  after all operations
- **Enhanced Search with Date Filtering**: Search form now supports text queries plus optional date range filtering with 't' shortcut for today

Previous features:

- Colored events with color picker (`c` key)
- Undo/redo functionality (`u` and `r` keys)
- Yank/paste system for events (`y`, `p`, `d` keys)
- Jump navigation (`g` key)
- Enhanced search with date filtering (`/` key) - supports text queries and date ranges with 't' shortcut
- Previous/next event navigation within week (`w` and `b` keys)

## notes

- shift-tab functionality cant be implemented... see:
  <https://github.com/jroimartin/gocui/issues/111>

