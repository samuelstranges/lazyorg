# Chronos

A sophisticated terminal-based calendar and event management system with
advanced scheduling features. Forked from
[HubertBel/lazyorg](https://github.com/HubertBel/lazyorg), and from there just
basically gave Claude Code the reins...

## Features

- 📅 **Smart Event Management**: Advanced event creation with automatic overlap
  prevention
- 🔄 **Undo/Redo System**: Full operation history with `u` and `r` keys
- 🎨 **Colored Events**: Automatic color assignment or manual color selection
  with `c`
- 📋 **Yank/Paste Events**: Copy events with `y`, paste with `p`, delete with
  `d`
- 🔍 **Smart Search**: Search events within current week with `/`
- 🎯 **Jump Navigation**: Quick navigation with `g` and event jumping with
  `w`/`b`
- ⏰ **Current Time Highlighting**: Visual indicators for current time
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

- **`-db <path>`** - Specify custom database file location  
  Default: `~/.local/share/chronos/data.db`
- **`-backup <path>`** - Backup database to specified location and exit
- **`-debug`** - Enable debug logging to `/tmp/chronos_debug.txt` and `/tmp/chronos_getevents_debug.txt`
- **`--help`** - Show all available command-line options

## Usage

Press `?` in the application to see all available keybindings, or see the quick
reference below:

### Navigation

- `h/l` or `←/→` - Previous/Next day
- `H/L` - Previous/Next week
- `j/k` or `↑/↓` - Move time cursor up/down
- `t` - Jump to today
- `g` - Go to specific date
- `w/b` - Jump to next/previous event

### Events

- `a` - Add new event
- `e` - Edit current event
- `d` - Delete current event (also copies to clipboard)
- `D` - Delete all events with same name
- `y` - Copy/yank current event
- `p` - Paste copied event
- `c` - Change event color

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

## Acknowledgments

- Inspired by [lazygit](https://github.com/jesseduffield/lazygit)
- Built with [gocui](https://github.com/jroimartin/gocui) TUI framework
- Forked from [HubertBel/lazyorg](https://github.com/HubertBel/lazyorg)
