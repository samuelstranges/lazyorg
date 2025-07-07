# Chronos

<div align="center">

**A sophisticated terminal-based calendar and event management system**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-lightgrey)](https://github.com/samuelstranges/chronos)

_Vim-inspired • Terminal-native • Offline-first_

</div>

## 🚀 Quick Start

```bash
# Clone and build
git clone https://github.com/samuelstranges/chronos.git
cd chronos
go build -o chronos cmd/chronos/main.go

# Run in local directory
./chronos

# Press ? for help, q to quit
```

## 📖 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Usage](#-usage)
- [Configuration](#️-configuration)
- [CLI Interface](#-cli-interface)
- [Examples](#-examples)
- [Architecture](#️-architecture)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features

### 🎯 Core Features

- **📅 Smart Event Management** - Create, edit, and delete events with automatic
  overlap prevention
- **🔄 Undo/Redo System** - Full operation history with vim-style `u` and `r`
  keys
- **🎨 Colored Events** - Automatic color assignment or manual selection with
  `C`
- **📋 Yank/Paste Events** - Copy events with `y`, paste with `p`, delete with
  `d`
- **🔍 Smart Search** - Search events across all dates with `/` (supports text
  and date filtering)

### 🖥️ Interface

- **Multiple View Modes** - Week view, month view, and agenda view (toggle with
  `v`)
- **📱 Responsive Design** - Dynamic viewport adjustment for different terminal
  sizes
- **⌨️ Vim-style Keybindings** - Familiar navigation and shortcuts
- **⏰ Current Time Highlighting** - Visual indicators for current time

### 🌍 Integrations

- **🌤️ Weather Integration** - 3-day weather forecast in month view with
  configurable location
- **🔔 Desktop Notifications** - Configurable event reminders (0-60 minutes
  before events)
- **🖥️ CLI Query Interface** - Get event information from command line without
  launching GUI
- **📄 iCalendar Export** - Export events to `.ics` format for other calendar
  apps

### 🔒 Data Management

- **🗃️ SQLite Database** - Lightweight, fast, and reliable local storage
- **💾 Backup Support** - Easy database backup and restore
- **🔄 Conflict Prevention** - Automatic detection and prevention of overlapping
  events
- **🚀 Offline-First** - No internet connection required for core functionality

## 🛠️ Installation

### Prerequisites

- Go 1.21 or higher
- Linux or macOS (Windows support not tested)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/samuelstranges/chronos.git
cd chronos

# Build the application
go build -o chronos cmd/chronos/main.go

# Optional: Install to system PATH
sudo mv chronos /usr/local/bin/
```

### Quick Test

```bash
# Run with default settings
chronos

# Run with custom database
chronos -db ~/my-calendar.db

# Get help
chronos --help
```

## 📚 Usage

### Interface Overview

Chronos provides three main view modes, which can be cycled through with `v`:

| View            | Description                       |
| --------------- | --------------------------------- |
| **Week View**   | 7-day layout with half-hour slots |
| **Month View**  | Monthly calendar grid             |
| **Agenda View** | Detailed daily event list         |

### Keybindings

| Category       | Key            | Action                           |
| -------------- | -------------- | -------------------------------- |
| **View**       | `q`            | Quit                             |
|                | `?`            | Show/Hide help                   |
|                | `v`            | Toggle view mode                 |
| **Navigation** | `h/l` or `←/→` | Previous/Next day                |
|                | `H/L`          | Previous/Next week               |
|                | `m/M`          | Previous/Next month              |
|                | `j/k` or `↑/↓` | Move time cursor                 |
|                | `t`            | Jump to today                    |
|                | `T`            | Jump to specific date            |
|                | `w/b/e`        | Next/Previous/End event          |
|                | `g/G`          | Start/End of day                 |
| **Events**     | `a`            | Add new event                    |
|                | `c`            | Change/Edit event                |
|                | `C`            | Change event color               |
|                | `y`            | Yank/Copy event                  |
|                | `p`            | Paste event                      |
|                | `d`            | Delete event                     |
|                | `D`            | Delete all events with same name |
| **Search**     | `/`            | Search events                    |
|                | `n/N`          | Next/Previous search result      |
|                | `Esc`          | Clear search                     |
| **Operations** | `u`            | Undo last operation              |
|                | `r`            | Redo last operation              |

### Creating Events

When adding a new event (`a` key):

1. **Name** - Event title
2. **Date** - YYYYMMDD format (e.g., 20250707)
3. **Time** - HH:MM format (30-minute intervals)
4. **Location** - Optional location
5. **Duration** - In hours (0.5 = 30 minutes)
6. **Frequency** - Repeat interval in days (1 = daily, 7 = weekly)
7. **Occurrences** - Number of repetitions
8. **Color** - leave blank for default
9. **Description** - Optional details

### Search System

Press `/` to open the search dialog with powerful filtering:

- **Text Search** - Search names, descriptions, locations
- **Date Range** - Filter text search by date range (YYYYMMDD format)
- **Today Shortcut** - Use `t` for today's date (works on start and end dates)

**Examples:**

- `meeting` - Find all meetings
- `doctor` + From: `t` - Doctor appointments from today

## ⚙️ Configuration

### Database Location

**Default:** `~/.local/share/chronos/data.db`

**Custom location:**

```bash
# Command line
./chronos -db /path/to/custom.db

# Config file: ~/.config/chronos/config.json
{
    "database_path": "/path/to/custom.db"
}
```

### Basic Configuration

Create `~/.config/chronos/config.json`:

```json
{
    "default_view": "week",
    "database_path": "~/.local/share/chronos/data.db"
}
```

### Weather Integration

Add weather to your calendar (using `wttr.in`). Weather data is polled
asynchronously on restart, then every two hours:

```json
{
    "weather_location": "Melbourne",
    "weather_unit": "celsius"
}
```

**Options:**

- `weather_location` - City name, airport code, or coordinates
- `weather_unit` - "celsius" or "fahrenheit"

### Desktop Notifications

Set up event reminders:

```json
{
    "notifications_enabled": true,
    "notification_minutes": 15
}
```

**Options:**

- `notifications_enabled` - true/false
- `notification_minutes` - 0-60 minutes before event

### Default Event Settings

Configure default values for new events:

```json
{
    "default_color": "Blue",
    "default_event_length": 2.0
}
```

**Options:**

- `default_color` - Default color for new events ("Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White", or empty for auto-generation)
- `default_event_length` - Default duration in hours (0.1-24.0 hours)

### Complete Configuration Example

```json
{
    "default_view": "month",
    "database_path": "/home/user/calendar.db",
    "weather_location": "Melbourne",
    "weather_unit": "celsius",
    "notifications_enabled": true,
    "notification_minutes": 30,
    "default_color": "Blue",
    "default_event_length": 1.5
}
```

## 🖥️ CLI Interface

Query events without opening the GUI:

```bash
# Get next upcoming event
chronos --next

# Get current event
chronos --current

# Get today's agenda
chronos --agenda

# Get specific date agenda
chronos --agenda 20250707

# Export to iCalendar
chronos --ics ~/calendar.ics

# Test notifications
chronos --test-notification

# Backup database
chronos -backup ~/backup.db

# Enable debug mode
chronos -debug
```

## 🏗️ Architecture

### Core Components

- **Database Layer** - SQLite for event storage
- **Event Manager** - Handles CRUD operations with undo/redo
- **UI Layer** - Multiple views with responsive design
- **Weather Service** - Optional weather integration using `wttr.in`
- **Notification Service** - Desktop notification system using
  [beeep](https://github.com/gen2brain/beeep)

### Key Design Principles

- **Offline-First** - No internet required for core functionality
- **Vim-Inspired** - Familiar keybindings for power users
- **Responsive** - Adapts to any terminal size
- **Fast** - Optimized for quick navigation and editing

### Technical Features

- **Viewport System** - Dynamic scrolling for different terminal sizes
- **UTC Storage** - Timezone-aware event storage
- **Conflict Detection** - Automatic overlap prevention

### Known limitations

- **Online sync** - Unlikely due to mass processing of events (think `D`).
  Having a local database as single source of truth improves speed & flexibility
- **.ics imports** - due to limitations of non overlapping events of increments
  of 30 mins
- **Shift-tab through forms** - not supported by gocui
- **Wraparound events past 12am** - things get wonky fast...

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file
for details.

## 🙏 Acknowledgments

- Originally forked from
  [HubertBel/lazyorg](https://github.com/HubertBel/lazyorg), from there, gave
  Claude Code the reins...
- Built with [gocui](https://github.com/jroimartin/gocui) TUI framework
- Weather data provided by [wttr.in](https://wttr.in)
- Notifications powered by [beeep](https://github.com/gen2brain/beeep)
