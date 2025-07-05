package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/samuelstranges/chronos/internal/config"
	"github.com/samuelstranges/chronos/internal/database"
	"github.com/samuelstranges/chronos/pkg/views"
	"github.com/jroimartin/gocui"
)

func main() {
	// Create debug log file
	debugFile, err := os.Create("/tmp/chronos_view_debug.txt")
	if err != nil {
		log.Fatal("Failed to create debug file:", err)
	}
	defer debugFile.Close()

	debug := func(msg string, args ...interface{}) {
		logMsg := fmt.Sprintf(msg, args...)
		fmt.Fprintf(debugFile, "%s: %s\n", time.Now().Format("15:04:05.000"), logMsg)
		debugFile.Sync() // Force write
		fmt.Printf("DEBUG: %s\n", logMsg)
	}

	debug("Starting chronos debug session")

	// Initialize database
	database := &database.Database{DebugMode: true}
	err = database.InitDatabase(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseDatabase()

	debug("Database initialized")

	// Create GUI
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Fatal(err)
	}
	defer g.Close()

	debug("GUI created")

	// Create app view
	cfg := config.GetDefaultConfig()
	av := views.NewAppView(g, database, cfg)
	
	debug("AppView created")

	// We'll check dimensions in the keybindings instead

	g.SetManager(av)

	debug("Layout manager set")

	// Set up keybindings with debug wrapper
	keybindings := map[rune]func(*gocui.Gui, *gocui.View) error{
		'm': func(g *gocui.Gui, v *gocui.View) error {
			debug("Switching to month view")
			av.SwitchToMonthView()
			av.UpdateCurrentView(g)
			debug("Month view switch completed")
			return nil
		},
		'W': func(g *gocui.Gui, v *gocui.View) error {
			debug("Switching to week view")
			av.SwitchToWeekView()
			av.UpdateCurrentView(g)
			debug("Week view switch completed")
			return nil
		},
		'q': func(g *gocui.Gui, v *gocui.View) error {
			debug("Quit requested")
			return gocui.ErrQuit
		},
		'd': func(g *gocui.Gui, v *gocui.View) error {
			debug("=== DIMENSION DEBUG ===")
			maxX, maxY := g.Size()
			debug("Terminal size: %dx%d", maxX, maxY)
			
			if mainView, ok := av.GetChild("main"); ok {
				if mv, ok := mainView.(*views.MainView); ok {
					x, y, w, h := mv.GetProperties()
					debug("MainView: x=%d, y=%d, w=%d, h=%d, mode=%s", x, y, w, h, mv.ViewMode)
					
					if monthView, ok := mv.GetChild("month"); ok {
						mx, my, mw, mh := monthView.GetProperties()
						debug("MonthView: x=%d, y=%d, w=%d, h=%d", mx, my, mw, mh)
						
						if mv, ok := monthView.(*views.MonthView); ok {
							debug("MonthView grid: %dx%d, cellW=%d, cellH=%d", mv.GridCols, mv.GridRows, mv.CellWidth, mv.CellHeight)
						}
					}
					
					if timeView, ok := mv.GetChild("time"); ok {
						tx, ty, tw, th := timeView.GetProperties()
						debug("TimeView: x=%d, y=%d, w=%d, h=%d", tx, ty, tw, th)
					}
					
					if weekView, ok := mv.GetChild("week"); ok {
						wx, wy, ww, wh := weekView.GetProperties()
						debug("WeekView: x=%d, y=%d, w=%d, h=%d", wx, wy, ww, wh)
					}
				}
			}
			debug("=== END DIMENSION DEBUG ===")
			return nil
		},
	}

	for key, handler := range keybindings {
		if err := g.SetKeybinding("", key, gocui.ModNone, handler); err != nil {
			log.Fatal(err)
		}
	}

	debug("Keybindings set up")

	fmt.Println("Debug chronos started!")
	fmt.Println("Controls:")
	fmt.Println("  m - Switch to month view")
	fmt.Println("  W - Switch to week view") 
	fmt.Println("  d - Print dimension debug info")
	fmt.Println("  q - Quit")
	fmt.Println()
	fmt.Printf("Debug log: /tmp/chronos_view_debug.txt\n")
	fmt.Println()

	// Run the main loop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		debug("MainLoop error: %v", err)
		log.Fatal(err)
	}

	debug("Chronos debug session ended")
}