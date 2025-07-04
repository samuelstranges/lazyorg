# Chronos

A sophisticated terminal-based calendar and event management system with
advanced scheduling features. Forked from
[HubertBel/lazyorg](https://github.com/HubertBel/lazyorg), and from there just
basically gave Claude Code the reigns...

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
go build
./chronos

# Custom database location
./chronos -db /path/to/custom/database.db

# Backup database
./chronos -backup /path/to/backup.db

# Enable debug logging
./chronos -debug
```

## Usage

Press `?` in the application to see all available keybindings, or see the quick
reference below:

### Navigation

- `h/l` or `←/→` - Previous/Next day
- `H/L` - Previous/Next week
- `j/k` or `↑/↓` - Move time cursor up/down
- `T` - Jump to today
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

- `/` - Search events in current week
- `n/N` - Go to next/previous search result
- `Esc` - Clear search

### Operations

- `u` - Undo last operation
- `r` - Redo last undone operation

### View Controls

- `Ctrl+s` - Show/Hide side view
- `?` - Show/Hide help menu
- `q` - Quit application

### Event Creation

When creating a new event (`a`), you'll be prompted to fill in:

- **Name**: Title of event
- **Time**: Date and time of the event
- **Location** (optional): Location of the event
- **Duration**: Duration of the event in hours (0.5 = 30 minutes)
- **Frequency**: Repeat interval in days (default: 7 for weekly)
- **Occurence**: Number of repetitions (default: 1)
- **Description** (optional): Additional notes or details

## Configuration

### Database Location
By default, the database is created at `~/.local/share/chronos/data.db`. You can specify a custom location:

**Command Line:**
```bash
./chronos -db /path/to/custom/database.db
```

**Config File:**
Edit `~/.config/chronos/config.json`:
```json
{
  "hide_day_on_startup": true,
  "database_path": "/path/to/custom/database.db"
}
```

Command line flags take precedence over config file settings.

## Acknowledgments

- Inspired by [lazygit](https://github.com/jesseduffield/lazygit)
- Built with [gocui](https://github.com/jroimartin/gocui) TUI framework
- Forked from [HubertBel/lazyorg](https://github.com/HubertBel/lazyorg)
