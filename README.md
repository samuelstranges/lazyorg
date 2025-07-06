# Chronos

A sophisticated terminal-based calendar and event management system with
advanced scheduling features. Forked from
[HubertBel/lazyorg](https://github.com/HubertBel/lazyorg), and from there just
basically gave Claude Code the reins...

## Features

- **Multiple View Modes**: Week view, month view and agenda (day) views (toggle
  with `v`)
- **CLI Query Interface**: Get event information from command line without
  launching GUI
- **Intuitive vim-like functionality**
- 📅 **Smart Event Management**: Advanced event creation with automatic overlap
  prevention
- 🔄 **Undo/Redo System**: Full operation history with `u` and `r` keys
- 🎨 **Colored Events**: Automatic color assignment or manual color selection
  with `C`
- 📋 **Yank/Paste Events**: Copy events with `y`, paste with `p`, delete with
  `d`
- 🔍 **Smart Search**: Search events across all dates with `/` (supports text
  and date filtering)
- 🎯 **Jump Navigation**: Quick navigation with `g` and event jumping with
  `w`/`b`
- ⏰ **Current Time Highlighting**: Visual indicators for current time
- 🌤️ **Weather Integration**: 3-day weather forecast in month view with
  configurable location and units
- 🔔 **Desktop Notifications**: Configurable event reminders (0-60 minutes
  before events)
- ⌨️ **Vim-style Keybindings**: Familiar navigation and shortcuts
- 🔒 **Conflict Prevention**: Automatic detection and prevention of overlapping
  events

## Installation

```bash
git clone https://github.com/samuelstranges/chronos.git
cd chronos
go build cmd/chronos/main.go
./main

# Or build with specific output name
go build -o chronos cmd/chronos/main.go
./chronos

# Custom database location
./chronos -db /path/to/custom/database.db

# Backup database
./chronos -backup /path/to/backup.db

# Enable debug logging
./chronos -debug

# Show all available command-line options
./chronos --help
```

## Command-Line Options

### Database and Configuration

- **`-db <path>`** - Specify custom database file location  
  Default: `~/.local/share/chronos/data.db`
- **`-backup <path>`** - Backup database to specified location and exit
- **`-debug`** - Enable debug logging to `/tmp/chronos_debug.txt` and
  `/tmp/chronos_getevents_debug.txt`
- **`--help`** - Show all available command-line options

### Configuration File

Create `~/.config/chronos/config.json` to customize application settings:

```json
{
    "database_path": "/path/to/custom/database.db",
    "default_view": "month"
}
```

**Available Options:**

- **`database_path`** - Custom database file location (overrides default)
- **`default_view`** - Default view mode on startup: `"week"` (default),
  `"month"`, or `"agenda"`

**Example Configurations:**

```json
// Start in month view
{"default_view": "month"}

// Start in agenda view with custom database
{"default_view": "agenda", "database_path": "/home/user/my_calendar.db"}
```

### Event Queries (CLI Mode)

- **`--next`** - Return next upcoming event
- **`--current`** - Return current event (if exists)
- **`--agenda [YYYYMMDD]`** - Export agenda for today or specified date
- **`--test-notification`** - Send a test desktop notification

### CLI Query Examples

```bash
# Get next upcoming event
./chronos --next

# Get current event (if any)
./chronos --current

# Get today's agenda
./chronos --agenda

# Get agenda for specific date (June 17, 2025)
./chronos --agenda 20250617

# Test desktop notifications
./chronos --test-notification
```

## Usage

Press `?` in the application to see all available keybindings, or see the quick
reference below:

### Navigation

- `h/l` or `←/→` - Previous/Next day
- `H/L` - Previous/Next week
- `m/M` - Previous/Next month (in month view)
- `j/k` or `↑/↓` - Move time cursor up/down
- `t` - Jump to today
- `T` - To specific date
- `w/b` - Jump to next/previous event
- `e` - Jump to end of current event, or next event if already at end
- `g/G` - Jump to start/end of day (00:00/23:30)

### View Modes

- `v` - Toggle between Week View, Month View, and Agenda View
    - **Week View**: 7-day calendar with half-hour time slots (default)
    - **Month View**: Monthly calendar grid overview
    - **Agenda View**: Daily event list with detailed information

### Events

- `a` - Add new event
- `c` - Change current event
- `d` - Delete current event (also copies to clipboard)
- `D` - Delete all events with same name
- `y` - Copy/yank current event
- `p` - Paste copied event
- `C` - Change event color

### Search

- `/` - Search events with text and date filters
- `n/N` - Go to next/previous search result
- `Esc` - Clear search

#### Search Filters

The search function supports multiple types of filters:

**Text Search:**

- Search across event names, descriptions, and locations
- Case-insensitive matching
- Partial text matching supported

**Date Filters:**

- **From Date**: Filter events starting from a specific date
- **To Date**: Filter events ending by a specific date
- **Date Format**: Use `YYYYMMDD` format (e.g., `20241225`)
- **Today Shortcut**: Use `t` to represent today's date

**Examples:**

- Text only: `meeting` (finds all events containing "meeting")
- Date range: From `20240101` to `20240131` (January events)
- Today shortcut: From `t` to `t` (today's events only)
- Mixed: `doctor` with From Date `t` (appointments with "doctor" from today
  onwards)

### Operations

- `u` - Undo last operation
- `r` - Redo last undone operation

### View Controls

- `?` - Show/Hide help menu
- `q` - Quit application

### Event Creation

When creating a new event (`a`), you'll be prompted to fill in:

- **Name**: Title of event
- **Date**: Date of the event (YYYYMMDD format)
- **Time**: Time of the event (HH:MM format, 30-minute intervals)
- **Location** (optional): Location of the event
- **Duration**: Duration of the event in hours (0.5 = 30 minutes)
- **Frequency**: Repeat interval in days (default: 7 for weekly)
- **Occurence**: Number of repetitions (default: 1)
- **Description** (optional): Additional notes or details

## Configuration

### Database Location

By default, the database is created at `~/.local/share/chronos/data.db`. You can
specify a custom location:

**Command Line:**

```bash
./chronos -db /path/to/custom/database.db
```

**Config File:** Edit `~/.config/chronos/config.json`:

```json
{
    "database_path": "/path/to/custom/database.db"
}
```

Command line flags take precedence over config file settings.

### Weather Integration

Chronos supports optional weather integration that displays current weather in
the title bar and 3-day forecasts in month view.

**Configuration:** Edit `~/.config/chronos/config.json`:

```json
{
    "weather_location": "Melbourne",
    "weather_unit": "celsius"
}
```

**Weather Configuration Options:**

- `weather_location` (string): Location for weather data (required to enable
  weather)

    - Examples: `"London"`, `"New York"`, `"Tokyo"`, `"Melbourne"`
    - Supports cities, airports (3-letter codes), coordinates
    - Leave empty or omit to disable weather features

- `weather_unit` (string): Temperature unit preference (optional)
    - `"celsius"` or `"c"` - Show temperatures in Celsius (default)
    - `"fahrenheit"` or `"f"` - Show temperatures in Fahrenheit
    - Invalid values default to Celsius

**Weather Display:**

- **Title Bar**: Shows current weather in all views (e.g., "Melbourne: ☁️ 21°C")
- **Month View**: Shows 3-day forecast next to day numbers (e.g., "6• 17°⛅", "7
  16°☀️")

**Features:**

- **Smart Caching**: Weather data cached for 2 hours to minimize API calls
- **Background Loading**: Weather preloads on startup to prevent lag when
  switching views
- **Automatic Updates**: Data refreshes every 2 hours automatically
- **Emoji Support**: Uses weather emojis (☀️, ⛅, ☁️, 🌧️, etc.) for visual
  indicators

**Technical Notes:**

- Weather icons appear last in day display due to emoji width rendering issues
  in terminals
- Supports wttr.in service locations (IP-based, coordinates, city names, airport
  codes)
- No API key required - uses the free wttr.in weather service
- Graceful fallback - weather failures don't affect calendar functionality

**Example Month View with Weather:**

```
│ 6• 17°⛅           │ 7 16°☀️            │ 8 14°⛅            │
│06:00 Morning       │06:00 Morning       │06:00 Morning       │
│08:30 Pump up tyres │18:00 Meeting       │                    │
```

### Desktop Notifications

Chronos supports optional desktop notifications that can remind you of upcoming
events 0-60 minutes before they start.

**Configuration:** Edit `~/.config/chronos/config.json`:

```json
{
    "notifications_enabled": true,
    "notification_minutes": 15
}
```

**Notification Configuration Options:**

- `notifications_enabled` (boolean): Enable or disable desktop notifications

    - `true` - Enable desktop notifications (default: `false`)
    - `false` - Disable desktop notifications

- `notification_minutes` (integer): Minutes before event to show notification
    - Valid range: 0-60 minutes
    - Default: 15 minutes
    - Values outside range default to 15 minutes

**Features:**

- **Cross-Platform**: Works on macOS and Linux desktop environments
- **Smart Timing**: Notifications appear exactly N minutes before events start
- **Duplicate Prevention**: Won't spam multiple notifications for the same event
- **Timezone Aware**: Properly handles timezone conversions between local time
  and UTC storage
- **Event Details**: Shows event name, time, location, and description
  (truncated if long)
- **Test Function**: Use `--test-notification` flag to verify notifications work

**Example Notification:**

```
Title: "Upcoming Event"
Message: "Team Meeting
         2:30 PM - 3:30 PM
         Location: Conference Room A
         Weekly project sync and updates"
```

**Testing Notifications:**

```bash
# Test if notifications work
./chronos --test-notification

# If notifications are disabled, you'll see:
# "Notifications are disabled in config"
# "To enable notifications, add the following to ~/.config/chronos/config.json:"
```

**Technical Notes:**

- Uses the `beeep` library for cross-platform desktop notifications
- Notifications run in background every 30 seconds to check for upcoming events
- Automatically starts when you launch Chronos (if enabled)
- Events stored in UTC, notifications calculated in local timezone
- Won't duplicate notifications for events within 1 hour window

## Future

- major changes:
    - dynamically change view based on available size in week mode (rather than
      cutting things off of bottom, use the number of available seen lines to
      shift down what user sees)
- visual fixes:
    - form colors are ugly... this might be a limitation of gocui
- additional keybinds:
    - visually change duration shortcut (running out of keybinds...)
- export flags:
    - `--ics`
    - `--json`
    - `--csv`

## Not planned

- syncing: using a local database as a single source of truth improves speed and
  flexibility (undo/redo of mass change of events would be hard to sync), and
  fits well within the constraints of a TUI
- import not planned due to limitations of 30 min events
- shift-tab through forms: not supported by gocui
- handle events that wraparound the end of a day into the next day (i cant
  imaging handling multi multi day events... )

## Acknowledgments

- Forked from [HubertBel/lazyorg](https://github.com/HubertBel/lazyorg)
- Built with [gocui](https://github.com/jroimartin/gocui) TUI framework
