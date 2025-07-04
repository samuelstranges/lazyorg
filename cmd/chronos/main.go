package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	flag.StringVar(&backupPath, "backup", "", "Backup database to specified location")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug logging to /tmp/chronos_debug.txt and /tmp/chronos_getevents_debug.txt")
	flag.StringVar(&dbPath, "db", "", "Custom database file path (default: ~/.local/share/chronos/data.db)")
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
