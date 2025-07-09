package ui

import (
	"fmt"
	"os"
	"github.com/samuelstranges/chronos/pkg/views"
	"github.com/jroimartin/gocui"
)

type Keybind struct {
	key     interface{}
	handler func(*gocui.Gui, *gocui.View) error
}

// Debug logging function for keybindings
func debugLogKeybinding(key interface{}, viewName string, av *views.AppView) {
	f, err := os.OpenFile("/tmp/chronos_keybind_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	
	keyStr := fmt.Sprintf("%v", key)
	if keyStr == "65514" { // Arrow left
		keyStr = "ArrowLeft"
	} else if keyStr == "65515" { // Arrow right
		keyStr = "ArrowRight"
	} else if keyStr == "65516" { // Arrow up
		keyStr = "ArrowUp"
	} else if keyStr == "65517" { // Arrow down
		keyStr = "ArrowDown"
	}
	
	currentView := "unknown"
	if av != nil {
		currentView = av.GetCurrentViewName()
	}
	
	fmt.Fprintf(f, "KEY_PRESS: key=%s, view=%s, currentView=%s, monthMode=%t\n", 
		keyStr, viewName, currentView, av != nil && av.IsMonthMode())
}

func InitKeybindings(g *gocui.Gui, av *views.AppView) error {
	g.InputEsc = true

	if err := initMainKeybindings(g, av); err != nil {
		return err
	}
	if err := initHelpKeybindings(g, av); err != nil {
		return err
	}

	return nil
}

func initMainKeybindings(g *gocui.Gui, av *views.AppView) error {
	mainKeybindings := []Keybind{
		{'a', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('a', v.Name(), av); return av.ShowNewEventPopup(g) }},
		{'c', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('c', v.Name(), av); return av.ShowEditEventPopup(g) }},
		{'C', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('C', v.Name(), av); return av.ShowColorPicker(g) }},
		{'d', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('d', v.Name(), av); return av.ShowDurationPopup(g) }},
		{'h', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('h', v.Name(), av); av.UpdateToPrevDay(g); return nil }},
		{'l', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('l', v.Name(), av); av.UpdateToNextDay(g); return nil }},
		{'j', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('j', v.Name(), av); av.UpdateToNextTime(g); return nil }},
		{'k', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('k', v.Name(), av); av.UpdateToPrevTime(g); return nil }},
		{gocui.KeyArrowLeft, func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding(gocui.KeyArrowLeft, v.Name(), av); av.UpdateToPrevDay(g); return nil }},
		{gocui.KeyArrowRight, func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding(gocui.KeyArrowRight, v.Name(), av); av.UpdateToNextDay(g); return nil }},
		{gocui.KeyArrowDown, func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding(gocui.KeyArrowDown, v.Name(), av); av.UpdateToNextTime(g); return nil }},
		{gocui.KeyArrowUp, func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding(gocui.KeyArrowUp, v.Name(), av); av.UpdateToPrevTime(g); return nil }},
		{'t', func(g *gocui.Gui, v *gocui.View) error { av.JumpToToday(); av.UpdateCurrentView(g); return nil }},
		{'H', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevWeek(); return nil }},
		{'L', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextWeek(); return nil }},
		{'x', func(g *gocui.Gui, v *gocui.View) error { av.DeleteEvent(g); return nil }},
		{'B', func(g *gocui.Gui, v *gocui.View) error { av.DeleteEvents(g); return nil }},
		{'y', func(g *gocui.Gui, v *gocui.View) error { av.CopyEvent(g); return nil }},
		{'p', func(g *gocui.Gui, v *gocui.View) error { return av.PasteEvent(g) }},
		{'u', func(g *gocui.Gui, v *gocui.View) error { return av.Undo(g) }},
		{'r', func(g *gocui.Gui, v *gocui.View) error { return av.Redo(g) }},
		{'T', func(g *gocui.Gui, v *gocui.View) error { return av.ShowGotoPopup(g) }},
		{'D', func(g *gocui.Gui, v *gocui.View) error { return av.ShowDatePopup(g) }},
		{'w', func(g *gocui.Gui, v *gocui.View) error { av.JumpToNextEvent(); av.UpdateCurrentView(g); return nil }},
		{'b', func(g *gocui.Gui, v *gocui.View) error { av.JumpToPrevEvent(); av.UpdateCurrentView(g); return nil }},
		{'e', func(g *gocui.Gui, v *gocui.View) error { av.JumpToEndOfEvent(); av.UpdateCurrentView(g); return nil }},
		{'g', func(g *gocui.Gui, v *gocui.View) error { av.JumpToStartOfDay(); av.UpdateCurrentView(g); return nil }},
		{'G', func(g *gocui.Gui, v *gocui.View) error { av.JumpToEndOfDay(); av.UpdateCurrentView(g); return nil }},
		{'/', func(g *gocui.Gui, v *gocui.View) error { return av.StartSearch(g) }},
		{'n', func(g *gocui.Gui, v *gocui.View) error { av.GoToNextMatch(); av.UpdateCurrentView(g); return nil }},
		{'N', func(g *gocui.Gui, v *gocui.View) error { av.GoToPrevMatch(); av.UpdateCurrentView(g); return nil }},
		{'m', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextMonth(); return nil }},
		{'M', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevMonth(); return nil }},
		{'v', func(g *gocui.Gui, v *gocui.View) error { debugLogKeybinding('v', v.Name(), av); err := av.ToggleView(g); av.UpdateCurrentView(g); return err }},
		{gocui.KeyEsc, func(g *gocui.Gui, v *gocui.View) error { av.ClearSearch(); return nil }},
		{'?', func(g *gocui.Gui, v *gocui.View) error { return av.ShowKeybinds(g) }},
		{'q', func(g *gocui.Gui, v *gocui.View) error { return quit(g, v) }},
	}
	
	// Set keybindings for weekday views (needed for week view)
	for _, viewName := range views.WeekdayNames {
		for _, kb := range mainKeybindings {
			if err := g.SetKeybinding(viewName, kb.key, gocui.ModNone, kb.handler); err != nil {
				return err
			}
		}
	}
	
	// Set keybindings for month day views (needed for month view)
	for i := 0; i < 42; i++ {
		viewName := fmt.Sprintf("monthday_%d", i)
		for _, kb := range mainKeybindings {
			if err := g.SetKeybinding(viewName, kb.key, gocui.ModNone, kb.handler); err != nil {
				return err
			}
		}
	}
	
	// Set keybindings for agenda event views (needed for agenda view)
	// We'll set up a large number to handle all possible events
	for i := 0; i < 100; i++ {
		viewName := fmt.Sprintf("agenda_event_%d", i)
		for _, kb := range mainKeybindings {
			if err := g.SetKeybinding(viewName, kb.key, gocui.ModNone, kb.handler); err != nil {
				return err
			}
		}
	}
	
	// Set keybindings for the main agenda view
	for _, kb := range mainKeybindings {
		if err := g.SetKeybinding("agenda", kb.key, gocui.ModNone, kb.handler); err != nil {
			return err
		}
	}

	return nil
}


func initHelpKeybindings(g *gocui.Gui, av *views.AppView) error {
	helpKeybindings := []Keybind{
		{gocui.KeyEsc, func(g *gocui.Gui, v *gocui.View) error { return av.ShowKeybinds(g) }},
		{'?', func(g *gocui.Gui, v *gocui.View) error { return av.ShowKeybinds(g) }},
		{'q', func(g *gocui.Gui, v *gocui.View) error { return quit(g, v) }},
	}
	for _, kb := range helpKeybindings {
		if err := g.SetKeybinding("keybinds", kb.key, gocui.ModNone, kb.handler); err != nil {
			return err
		}
	}

	return nil
}


func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
