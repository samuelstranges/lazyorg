# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository.

## IMPORTANT

- you dont have access to a TTY so you cant do testing yourself... put yourself
  in the best position possible by writing debug code for the user so he can run
  it for you... then ask him to do it
- write debug stuff to /tmp/ and then request access to those files so you don't
  have to keep asking for output from user with `cat`

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
- **`-debug`** - Enable debug logging to `/tmp/chronos_debug.txt` and
  `/tmp/chronos_getevents_debug.txt`
- **`--next`** - Return next upcoming event and exit
- **`--current`** - Return current event (if exists) and exit
- **`--agenda [YYYYMMDD]`** - Export agenda for today or specified date and exit
- **`--ics <path>`** - Export all events to iCalendar (.ics) file and exit
- **`--test-notification`** - Send a test desktop notification and exit
- **`--help`** - Show all available command-line options

### CLI Query Examples

```bash
# Get next upcoming event
./chronos --next

# Get current event (if any)
./chronos --current

# Get today's agenda
./chronos --agenda

# Get agenda for specific date
./chronos --agenda 20250617

# Export all events to iCalendar file
./chronos --ics ~/my_calendar.ics
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
- `internal/weather/weather.go` - Weather data fetching, caching, and forecast
  management

**UI Layer:**

- `pkg/views/` - All UI view components (day-view, week-view, month-view,
  agenda-view, event-view, etc.)
- Uses `gocui` TUI framework for terminal interface
- Vim-style keybindings throughout
- Multiple view modes: Week View (default), Month View, and Agenda View

### Key Features

- **Multiple View Modes**: Week View, Month View, and Agenda View (toggle with
  `v`)
- **Weather Integration**: Optional 3-day weather forecast display with
  configurable location and temperature units
- **Undo/Redo System**: Full undo/redo support for event operations (add,
  delete, edit, bulk delete)
- **Event Management**: Create, edit, delete events with recurrence support
  - **Weekday Recurrence**: Use 'w' in frequency field for weekday-only events
  - **Regular Recurrence**: Use numbers for every N days (e.g., 7 for weekly)
- **Search**: Search events across all dates with `/` (supports text and date
  filtering)
- **Yank/Paste**: Copy events with `y` and paste with `p`
- **Color Coding**: Events are automatically colored or manually assigned colors
- **Notepad**: Integrated note-taking functionality
- **CLI Query Interface**: Query events from command line without launching GUI

### View Modes

**Week View (Default):**

- Shows 7-day week layout with half-hour time slots
- Current day highlighted with distinct formatting
- Current time indicator (purple highlighting)
- Primary view for event management and navigation

**Month View (`v` key to switch):**

- Calendar grid showing entire month
- Events displayed as abbreviated text within date cells
- Navigate between months with `m`/`M` keys
- Useful for overview and long-term planning

**Agenda View (`v` key to cycle):**

- Day-focused vertical list of events
- Shows events for current day in chronological order
- Detailed event information display
- Ideal for focused daily planning

### Event Creation and Recurrence

**Creating Events** (Press `a` to add new event):

- **Name**: Event title
- **Date**: YYYYMMDD format (e.g., 20241225)
- **Time**: HH:MM format in 30-minute increments (e.g., 14:30)
- **Location**: Event location (optional)
- **Duration**: Hours in decimal format (e.g., 1.5 for 90 minutes)
- **Frequency**: Recurrence pattern
  - **Numbers**: Every N days (e.g., `7` for weekly, `1` for daily)
  - **'w' or 'W'**: Weekdays only (Monday-Friday)
- **Occurrence**: Number of times to repeat
- **Color**: Event color (optional, auto-generated if empty)
- **Description**: Event details (optional)

**Weekday Recurrence Examples**:
- Event on Monday with frequency `w` and occurrence `5` creates events Mon-Fri
- Event on Saturday with frequency `w` and occurrence `8` starts on next Monday and creates 8 weekday events
- Event on Wednesday with frequency `w` and occurrence `3` creates events Wed-Fri of the same week

**Regular Recurrence Examples**:
- Event with frequency `7` and occurrence `4` creates 4 weekly events
- Event with frequency `1` and occurrence `10` creates 10 daily events

**Important Notes**:
- Recurring events will stop creating new instances if an existing event is found at the same time slot (overlap prevention)
- This applies to both weekday and regular recurrence patterns
- The occurrence count may not be fully reached if overlaps are detected

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

### Default View Configuration

- **Config file**: Set `default_view` in `~/.config/chronos/config.json`
- **Options**: `"week"` (default), `"month"`, or `"agenda"`
- **Example**: `{"default_view": "month"}` - Application will start in month
  view
- **Validation**: Invalid values automatically fallback to week view

### Weather Configuration

- **Config file**: Set `weather_location` and `weather_unit` in
  `~/.config/chronos/config.json`
- **Weather Location**: Required to enable weather features
    - Examples: `"Melbourne"`, `"London"`, `"NYC"`, `"LAX"` (airport codes),
      coordinates
    - Empty or omitted disables weather entirely
- **Weather Unit**: Optional temperature unit preference
    - `"celsius"` or `"c"` - Celsius temperatures (default)
    - `"fahrenheit"` or `"f"` - Fahrenheit temperatures
    - Invalid values default to Celsius
- **Example**: `{"weather_location": "Melbourne", "weather_unit": "fahrenheit"}`

### Notification Configuration

- **Config file**: Set `notifications_enabled` and `notification_minutes` in
  `~/.config/chronos/config.json`
- **Notifications Enabled**: Boolean flag to enable/disable desktop
  notifications
    - `true` - Enable desktop notifications
    - `false` - Disable desktop notifications (default)
- **Notification Minutes**: Integer value for minutes before event to notify
  (0-60)
    - Valid range: 0-60 minutes
    - Default: 15 minutes
    - Values outside range default to 15 minutes
- **Example**: `{"notifications_enabled": true, "notification_minutes": 30}`
- **Test**: Use `./chronos --test-notification` to verify notifications work

### Default Event Configuration

- **Config file**: Set `default_color` and `default_event_length` in
  `~/.config/chronos/config.json`
- **Default Color**: Sets the default color for new events
    - Valid values: `"Red"`, `"Green"`, `"Yellow"`, `"Blue"`, `"Magenta"`,
      `"Cyan"`, `"White"`
    - Empty string or invalid values: Auto-generate color from event name hash
    - Default: `""` (auto-generation)
- **Default Event Length**: Sets the default duration for new events in hours
    - Valid range: 0.1-24.0 hours (0.1 = 6 minutes, 24.0 = 24 hours)
    - Values outside range default to 1.0 hour
    - Default: 1.0 hour
- **Example**: `{"default_color": "Blue", "default_event_length": 2.0}`

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
- **Transparent UTC conversion**: Local time ↔ UTC conversion handled
  automatically

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

**UTC Conversion Implementation:**

- **Storage Layer**: All events stored in UTC timezone in database
- **UI Layer**: All events displayed and manipulated in local timezone
- **Conversion Layer**: EventManager handles automatic conversion between
  timezones
- **Write Operations**: Local time → UTC conversion before database storage
- **Read Operations**: UTC → Local time conversion after database retrieval
- **Query Operations**: Local time boundaries → UTC conversion for database
  queries
- **Date Queries**: Local day/month boundaries converted to UTC before database
  comparison
- **Undo/Redo**: Events stored in local time in undo stacks for UI consistency
- **Transparent**: No changes required to UI code, forms, or display logic

### Color System Architecture

**Available User Colors:**

- Red, Green, Yellow, Blue, Magenta, Cyan, White
- Accessible via color picker (`c` key) with shortcuts: r/g/y/b/m/c/w
- Events auto-generate colors from name hash if not manually set

**Color Implementation:**

- Colors stored as `gocui.Attribute` in database as integers
- `GetAvailableColors()` returns user-selectable colors only
- `GenerateColorFromName()` creates consistent auto-colors
- Color picker supports both single-letter shortcuts and full names

**Current Time Indicators:**

- **Time Sidebar Dot**: Minimal bullet (●) appears next to current time in
  sidebar
- **Title Bar Status**: Shows "Current Event: [Name]" or "Current Event: None"
- **Responsive Design**: Works with viewport scrolling and terminal resizing
- **Elegant Implementation**: No overlays or hash characters that interfere with
  events

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

The search functionality (`/` key) provides powerful filtering capabilities for
finding events across the entire database.

**Search Form Fields:**

- **Query**: Text search across event names, descriptions, and locations
  (case-insensitive, partial matching)
- **From Date**: Optional start date filter (YYYYMMDD format or 't' for today)
- **To Date**: Optional end date filter (YYYYMMDD format or 't' for today)

**Date Shortcut:**

- Use **'t'** in either date field as a shortcut for today's date
- Examples: `From Date: t` finds events from today onwards, `To Date: t` finds
  events up to today

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

- Field name "Query" (not "Search") to avoid naming conflict with "Search"
  button
- Validation messages document the 't' shortcut in tooltips

### Responsive Viewport System

The week view implements a responsive viewport system that automatically adjusts
to terminal size, providing a web-app-like responsive experience.

**Core Features:**

- **Dynamic View Adjustment**: Automatically calculates visible time slots based
  on terminal height
- **Intelligent Scrolling**: Centers viewport around cursor position to prevent
  content cutoff
- **Border-Aware Calculations**: Accounts for view borders to ensure all time
  slots are properly visible
- **23:30 Special Handling**: Ensures the last time slot (23:30) is comfortably
  visible, not hidden at bottom edge

**Implementation Architecture:**

- **TimeView Viewport Management**:

    - `ViewportStart`: Starting time slot for visible area (0-47, representing
      00:00-23:30)
    - `MaxTimeSlots`: Total available time slots (48 for 24-hour day)
    - `AutoAdjustViewport()`: Automatically centers viewport based on cursor
      position
    - `GetVisibleSlots()`: Calculates visible slots as `terminal_height - 1`
      (reserves border space)

- **Event Positioning**:
    - `TimeToPositionWithViewport()`: Calculates viewport-relative positions for
      events
    - Events outside viewport are automatically skipped during rendering
    - Event heights are truncated if they extend beyond visible area

**Key Functions:**

- `AutoAdjustViewport(calendarTime)` - Centers viewport around specified time
- `GetViewportStart()` - Returns current viewport starting position
- `GetVisibleSlots()` - Returns number of visible time slots
- `AdjustViewportForCursor()` - Centers viewport around cursor position

**Responsive Behavior:**

- **Small Terminals**: Shows fewer time slots, automatically scrolls to keep
  cursor visible
- **Large Terminals**: Shows more/all time slots, may show entire day if
  terminal is tall enough
- **Window Resize**: Viewport automatically readjusts on terminal size changes
- **Navigation**: Viewport follows cursor movement to maintain visibility

**Technical Details:**

- Viewport calculations use `tv.H - 1` to reserve space for view borders
- Special logic for 23:30 positioning prevents it from being hidden at bottom
  edge
- Events are positioned using `utils.TimeToPositionWithViewport()` for viewport
  awareness
- MainView coordinates viewport adjustment during UI updates

### Known Issues and Technical Debt

#### Mixed Timezone Storage in Database - RESOLVED

**Problem**: Historical events in the database had inconsistent timezone
information:

- Events created via **Add Event popup** (before fix): Stored in UTC (not UTC)
  due to `time.Parse()`
- Events created via **paste operations**: Stored in local timezone due to
  `time.Date()` with location
- This caused overlap detection failures when comparing events from different
  creation methods

**Resolution**:

- **UTC Storage Implementation**: EventManager now handles transparent UTC
  conversion
- **Workaround Code Removed**: All temporary timezone workaround code has been
  cleaned up
- **Consistent Storage**: All events now stored consistently in UTC timezone via
  EventManager
- **Transparent Conversion**: UI works in local time, database stores in UTC,
  conversion handled automatically
- **Simplified Code**: Event comparison logic no longer needs timezone
  normalization workarounds

#### Navigation Bug Due to Mixed Timezone Storage - RESOLVED

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

**Resolution**:

- **UTC Storage Implementation**: EventManager now handles transparent UTC
  conversion
- **Workaround Code Removed**: All temporary timezone workaround code has been
  cleaned up from navigation functions
- **Simplified Navigation**: Navigation functions now use direct time
  comparisons without timezone workarounds
- **Transparent Conversion**: UI works in local time, database stores in UTC,
  conversion handled automatically
- **Consistent Sorting**: Event chronological ordering is now reliable across
  all views

#### Form Component Field/Button Name Conflicts

**Problem**: The gocui-component form library has naming conflicts when an input
field and button share the same name.

**Issue**: Originally, the search form had an input field named "Search" and a
button named "Search", which caused the form to render incorrectly:

- The "Search" button would not appear
- The input field text would render in the button area (next to "Cancel")
- Form layout would be completely broken

**Root Cause**: The form component library internally confuses fields and
buttons with identical names, causing rendering conflicts.

**Solution**: Renamed the search input field from "Search" to "Query" to avoid
the naming conflict.

**Prevention**: When creating forms, ensure input field names and button names
are always different:

- ✅ Good: Field "Query" + Button "Search"
- ✅ Good: Field "Name" + Button "Add"
- ❌ Bad: Field "Search" + Button "Search"
- ❌ Bad: Field "Edit" + Button "Edit"

**Current Status**:

- **Fixed**: Search form now works correctly with "Query" field and "Search"
  button
- **Enhanced**: Added date filtering fields ("From Date", "To Date") for
  advanced search
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

- **CLI Query Interface**: New command-line flags for event queries:
    - `--next` - Get next upcoming event
    - `--current` - Get current event (if active)
    - `--agenda [YYYYMMDD]` - Get agenda for today or specified date
- **Enhanced View System**: Improved view toggling with `v` key cycling through
  Week → Month → Agenda views
- **Vim-like Navigation**: Added `e` key for end-of-event navigation (moves to
  end of current event, or next event if already at end)
- **Keybinding Reorganization**: Moved goto functionality from `g` to `T` (To
  specific date/time)
- **Start/End of Day Navigation**: Added `g`/`G` keys for vim-like start (00:00)
  and end (23:30) of day navigation
- **Unified Date Format**: All date fields now use YYYYMMDD format (no dashes)
  for consistency across goto, add/edit, and search forms
- **Enhanced 't' Key**: Jump to today's date AND current time (rounded to
  nearest 30 minutes) for consistency with 't' shortcut in date fields and
  improved time awareness
- **Symmetric Day Boundary Navigation**: Fixed time navigation to work
  symmetrically - going up from 00:00 now goes to 23:30 of previous day,
  matching the behavior of going down from 23:30 to 00:00 of next day
- **Event Overlap Prevention**: All event operations now prevent scheduling
  conflicts
- **Elegant Current Time Indicators**:
    - Minimal bullet (●) in time sidebar shows current half-hour
    - Title bar displays "Current Event: [Name]" or "Current Event: None"
    - No intrusive overlays or hash characters that interfere with events
    - Responsive design works with viewport scrolling
- **Centralized Event Management**: All modifications route through EventManager
- **Enhanced Color System**: Improved color management and auto-generation
- **Enhanced Search with Date Filtering**: Search form now supports text queries
  plus optional date range filtering with 't' shortcut for today

Previous features:

- Colored events with color picker (`C` key)
- Undo/redo functionality (`u` and `r` keys)
- Yank/paste system for events (`y`, `p`, `d` keys)
- Jump navigation (`T` key for 'To' specific time, `g`/`G` for start/end of day)
- Enhanced search with date filtering (`/` key) - supports text queries and date
  ranges with 't' shortcut
- Previous/next event navigation within week (`w` and `b` keys)
- End of event navigation (`e` key) - vim-like movement to end of current event,
  or next event if already at end
- CLI query flags (`--next`, `--current`, `--agenda`) for command-line event
  access
- **Weather Integration**: Optional 3-day forecast display with background
  preloading and smart caching

## Weather Integration Architecture

### Overview

Weather integration is optional and only enabled when `weather_location` is
configured. The system provides:

- Current weather in title bar (all views)
- 3-day forecast in month view day cells
- Background preloading to prevent UI lag
- Smart caching to minimize API calls

### Core Components

**Weather Package** (`internal/weather/weather.go`):

- `WeatherCache`: 2-hour TTL caching for both current weather and forecasts
- `WeatherData`: Current weather information (temp, condition, icon, etc.)
- `WeatherForecast`: Multi-day forecast with `DayForecast` entries
- `fetchWeatherData()`: Gets current weather from wttr.in JSON API
- `fetchWeatherForecast()`: Gets 3-day forecast from wttr.in JSON API

**Configuration** (`internal/config/config.go`):

- `GetWeatherLocation()`: Returns configured location or empty string
- `GetWeatherUnit()`: Returns "celsius" or "fahrenheit" with validation
- `IsWeatherEnabled()`: Checks if weather_location is set

**UI Integration**:

- `AppView.preloadWeatherData()`: Background goroutine preloads on startup
- `AppView.updateWeatherData()`: Updates title bar weather (called on every UI
  update)
- `AppView.updateMonthViewWeather()`: Updates month view forecast (when in month
  mode)
- `MonthView.UpdateWeatherData()`: Sets weather icons/temps on day views
- `MonthDayView.SetWeatherData()`: Sets icon and temperature for display

### Data Flow

1. **Startup**: `preloadWeatherData()` runs in background goroutine
2. **UI Updates**: `updateWeatherData()` uses cached data for title bar
3. **Month View**: `updateMonthViewWeather()` updates day cells when in month
   mode
4. **Caching**: 2-hour TTL prevents API calls on every UI update
5. **API Source**: wttr.in provides free weather data (no API key required)

### Display Logic

**Month Day Format**: `[day][today_indicator] [temp]°[weather_icon]`

- Example: `6• 17°⛅` (day 6, today, 17°C, partly cloudy)
- Example: `7 16°☀️` (day 7, 16°C, sunny)

**Title Bar Format**: `[location]: [icon] [temp]`

- Example: `Melbourne: ☁️ 21°C`

### Technical Implementation Details

**Weather Icon Positioning**:

- Icons must appear LAST in display string due to emoji width rendering issues
- Some weather emojis (☀️, ☁️) are 2 Unicode runes, others (⛅) are 1 rune
- Terminal display width varies by emoji and font, causing text truncation
- Solution: Place temperature before icon, icon gets truncated if necessary

**Temperature Units**:

- API returns both Celsius and Fahrenheit
- `DayForecast` stores both `MaxTempC` and `MaxTempF`
- Display chooses appropriate unit based on config
- Graceful fallback to Celsius for invalid config values

**Caching Strategy**:

- Separate caches for current weather (`data`) and forecasts (`forecasts`)
- Separate TTL tracking (`lastFetch`, `lastForecast`)
- 2-hour cache prevents API abuse while keeping data reasonably fresh
- Cache key normalization (lowercase location names)

**Error Handling**:

- Weather failures never crash the application
- Missing weather data results in empty display (no weather shown)
- Network timeouts set to 10 seconds
- Graceful degradation when weather service unavailable

**Performance Optimizations**:

- Background preloading prevents view switch lag
- Cached responses for 2 hours (720 API calls/month maximum per location)
- Month view only updates weather when actually in month mode
- UI updates use cached data, never trigger API calls directly

### Known Issues and Workarounds

**Emoji Width Problems**:

- Different terminals render emojis with different display widths
- Some emojis appear as 1 character, others as 2 characters wide
- Text truncation occurs when cell width calculations don't account for emoji
  width
- Workaround: Always place weather icons at the end of display strings

**Month View Cell Width**:

- Month view cells have limited width (typically 17-19 characters)
- Weather string `6• 17°⛅` can exceed cell width with wide emoji rendering
- Solution: Compact format with temperature before emoji for better truncation
  behavior

**API Rate Limiting**:

- wttr.in has no official rate limits but good practice suggests reasonable
  usage
- 2-hour caching provides good balance of freshness vs API usage
- Approximately 12 API calls per day per location (very reasonable)

## notes

- shift-tab functionality cant be implemented... see:
  <https://github.com/jroimartin/gocui/issues/111>
