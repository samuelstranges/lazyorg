package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/config"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/internal/ui"
	"github.com/samuelstranges/chronos/pkg/views"
	"github.com/jroimartin/gocui"
)

func main() {
	var backupPath string
	var debugMode bool
	var dbPath string
	var nextFlag bool
	var currentFlag bool
	var agendaFlag bool
	flag.StringVar(&backupPath, "backup", "", "Backup database to specified location")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug logging to /tmp/chronos_debug.txt and /tmp/chronos_getevents_debug.txt")
	flag.StringVar(&dbPath, "db", "", "Custom database file path (default: ~/.local/share/chronos/data.db)")
	flag.BoolVar(&nextFlag, "next", false, "Return next event")
	flag.BoolVar(&currentFlag, "current", false, "Return current event (if exists)")
	flag.BoolVar(&agendaFlag, "agenda", false, "Export agenda for today or specified date (provide date as next argument in YYYYMMDD format)")
	flag.Parse()

	// Set up cursor restoration on exit
	setupCursorHandling()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		cfg = config.GetDefaultConfig()
	}

	// Command line flag takes precedence over config file
	var dbFilePath string
	if dbPath != "" {
		dbFilePath = dbPath
	} else {
		dbFilePath = config.GetDatabasePath(cfg)
	}

	// Create directory for database path if it doesn't exist
	dbDir := filepath.Dir(dbFilePath)
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	if backupPath != "" {
		err := backupDatabase(dbFilePath, backupPath)
		if err != nil {
			log.Fatal("Backup failed:", err)
		}
		fmt.Printf("Database backed up to: %s\n", backupPath)
		return
	}

	database := &database.Database{DebugMode: debugMode}
	err = database.InitDatabase(dbFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDatabase()

	// Handle command-line queries
	if nextFlag {
		handleNextEvent(database)
		return
	}
	
	if currentFlag {
		handleCurrentEvent(database)
		return
	}
	
	if agendaFlag {
		// Get the date argument if provided
		dateStr := ""
		if len(flag.Args()) > 0 {
			dateStr = flag.Args()[0]
		}
		handleAgenda(database, dateStr)
		return
	}

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer func() {
		restoreCursor()
		g.Close()
	}()

	// Set cursor to block shape for better visibility in tmux
	setCursorBlock()

	av := views.NewAppView(g, database, cfg)
	g.SetManager(av)

	if err := ui.InitKeybindings(g, av); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func backupDatabase(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source database: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy database: %w", err)
	}

	return nil
}

func setupCursorHandling() {
	// Set up signal handling for graceful cursor restoration
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	
	go func() {
		<-c
		restoreCursor()
		os.Exit(0)
	}()
}

func setCursorBlock() {
	// ESC[2 q = Set cursor to steady block
	fmt.Print("\033[2 q")
}

func restoreCursor() {
	// ESC[0 q = Reset cursor to default shape
	fmt.Print("\033[0 q")
}

// handleNextEvent finds and prints the next upcoming event
func handleNextEvent(db *database.Database) {
	events, err := db.GetAllEvents()
	if err != nil {
		log.Fatal("Error getting events:", err)
	}

	now := time.Now()
	var nextEvent *calendar.Event = nil
	var nextTime time.Time

	for _, event := range events {
		// Convert UTC stored time to local time for comparison
		eventTime := event.Time.Local()
		if eventTime.After(now) {
			if nextEvent == nil || eventTime.Before(nextTime) {
				nextEvent = event
				nextTime = eventTime
			}
		}
	}

	if nextEvent == nil {
		fmt.Println("No upcoming events found")
		return
	}

	// Convert UTC stored time to local time for display
	localTime := nextEvent.Time.Local()
	fmt.Printf("%s at %s\n", nextEvent.Name, localTime.Format("2006-01-02 15:04"))
	if nextEvent.Description != "" {
		fmt.Printf("Description: %s\n", nextEvent.Description)
	}
	if nextEvent.Location != "" {
		fmt.Printf("Location: %s\n", nextEvent.Location)
	}
}

// handleCurrentEvent finds and prints the current event if one exists
func handleCurrentEvent(db *database.Database) {
	events, err := db.GetEventsByDate(time.Now())
	if err != nil {
		log.Fatal("Error getting today's events:", err)
	}

	now := time.Now()
	
	for _, event := range events {
		// Convert UTC stored time to local time for comparison
		eventStart := event.Time.Local()
		eventEnd := eventStart.Add(time.Duration(event.DurationHour * float64(time.Hour)))
		
		if (now.After(eventStart) || now.Equal(eventStart)) && now.Before(eventEnd) {
			fmt.Printf("%s (until %s)\n", event.Name, eventEnd.Format("15:04"))
			if event.Description != "" {
				fmt.Printf("Description: %s\n", event.Description)
			}
			if event.Location != "" {
				fmt.Printf("Location: %s\n", event.Location)
			}
			return
		}
	}
	
	fmt.Println("No current event")
}

// handleAgenda prints agenda for specified date or today if no date provided
func handleAgenda(db *database.Database, dateStr string) {
	var targetDate time.Time
	var err error
	
	if dateStr == "" || dateStr == "today" {
		// If no date specified or "today" specified, use today
		targetDate = time.Now()
	} else {
		// Parse the date string in YYYYMMDD format
		targetDate, err = time.ParseInLocation("20060102", dateStr, time.Local)
		if err != nil {
			log.Fatal("Error parsing date (use YYYYMMDD format):", err)
		}
	}
	
	events, err := db.GetEventsByDate(targetDate)
	if err != nil {
		log.Fatal("Error getting events:", err)
	}

	if len(events) == 0 {
		if dateStr == "" || dateStr == "today" {
			fmt.Println("No events today")
		} else {
			fmt.Printf("No events on %s\n", targetDate.Format("Monday, January 2, 2006"))
		}
		return
	}

	// Sort events by time (comparing UTC times for accurate sorting)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Time.Before(events[j].Time)
	})

	if dateStr == "" || dateStr == "today" {
		fmt.Printf("Today's Agenda - %s\n", targetDate.Format("Monday, January 2, 2006"))
	} else {
		fmt.Printf("Agenda for %s\n", targetDate.Format("Monday, January 2, 2006"))
	}
	fmt.Println(strings.Repeat("=", 50))
	
	for _, event := range events {
		// Convert UTC stored time to local time for display
		localStartTime := event.Time.Local()
		localEndTime := localStartTime.Add(time.Duration(event.DurationHour * float64(time.Hour)))
		
		fmt.Printf("%s - %s: %s\n", localStartTime.Format("15:04"), localEndTime.Format("15:04"), event.Name)
		if event.Location != "" {
			fmt.Printf("  Location: %s\n", event.Location)
		}
		if event.Description != "" {
			fmt.Printf("  Description: %s\n", event.Description)
		}
	}
}
