package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/HubertBel/lazyorg/internal/config"
	"github.com/HubertBel/lazyorg/internal/database"
	"github.com/HubertBel/lazyorg/internal/ui"
	"github.com/HubertBel/lazyorg/pkg/views"
	"github.com/jroimartin/gocui"
)

func main() {
	var backupPath string
	var debugMode bool
	flag.StringVar(&backupPath, "backup", "", "Backup database to specified location")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug logging to /tmp/lazyorg_debug.txt and /tmp/lazyorg_getevents_debug.txt")
	flag.Parse()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	dbDirPath := filepath.Join(homeDir, ".local", "share", "lazyorg")
	dbFilePath := filepath.Join(dbDirPath, "data.db")

	err = os.MkdirAll(dbDirPath, 0755)
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

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: Could not load config, using defaults: %v", err)
		cfg = config.GetDefaultConfig()
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
	defer g.Close()

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
