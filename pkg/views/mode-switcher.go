package views

import (
	"github.com/jroimartin/gocui"
)

type KeybindMode string

const (
	ModeNormal     KeybindMode = "normal"
	ModeColorPicker KeybindMode = "colorpicker"
	ModeGoto       KeybindMode = "goto"
)

type ModeKeybind struct {
	Key     interface{}
	Handler func(*gocui.Gui, *gocui.View) error
}

type ModeConfig struct {
	Name        KeybindMode
	PopupViewName string
	Keybinds    []ModeKeybind
	OnEnter     func(*gocui.Gui) error
	OnExit      func(*gocui.Gui) error
}

type ModeSwitcher struct {
	currentMode KeybindMode
	gui         *gocui.Gui
	appView     *AppView
}

func NewModeSwitcher(g *gocui.Gui, av *AppView) *ModeSwitcher {
	return &ModeSwitcher{
		currentMode: ModeNormal,
		gui:         g,
		appView:     av,
	}
}

func (ms *ModeSwitcher) GetCurrentMode() KeybindMode {
	return ms.currentMode
}

func (ms *ModeSwitcher) IsInMode(mode KeybindMode) bool {
	return ms.currentMode == mode
}

func (ms *ModeSwitcher) EnterMode(config ModeConfig) error {
	if ms.currentMode != ModeNormal {
		if err := ms.ExitCurrentMode(); err != nil {
			return err
		}
	}

	ms.currentMode = config.Name

	if err := ms.removeAllMainKeybindings(); err != nil {
		return err
	}

	if err := ms.setModeKeybindings(config.Keybinds); err != nil {
		return err
	}

	if config.OnEnter != nil {
		if err := config.OnEnter(ms.gui); err != nil {
			return err
		}
	}

	return nil
}

func (ms *ModeSwitcher) ExitCurrentMode() error {
	var config ModeConfig
	
	switch ms.currentMode {
	case ModeColorPicker:
		config = ms.getColorPickerConfig()
	case ModeGoto:
		config = ms.getGotoConfig()
	default:
		return nil
	}

	if err := ms.removeModeKeybindings(config.Keybinds); err != nil {
		return err
	}

	if config.OnExit != nil {
		if err := config.OnExit(ms.gui); err != nil {
			return err
		}
	}

	if config.PopupViewName != "" {
		if err := ms.gui.DeleteView(config.PopupViewName); err != nil && err != gocui.ErrUnknownView {
			return err
		}
	}

	if err := ms.restoreAllMainKeybindings(); err != nil {
		return err
	}

	ms.currentMode = ModeNormal
	return nil
}

func (ms *ModeSwitcher) removeAllMainKeybindings() error {
	keysToRemove := []interface{}{
		'a', 'e', 'c', 'h', 'l', 'j', 'k', 'T', 'H', 'L', 'd', 'D', 'y', 'p', 'u', 'r', 'g',
		gocui.KeyCtrlN, gocui.KeyCtrlS, '?', 'q',
		gocui.KeyArrowLeft, gocui.KeyArrowRight, gocui.KeyArrowDown, gocui.KeyArrowUp,
	}

	for _, viewName := range WeekdayNames {
		for _, key := range keysToRemove {
			ms.gui.DeleteKeybinding(viewName, key, gocui.ModNone)
		}
	}

	return nil
}

func (ms *ModeSwitcher) setModeKeybindings(keybinds []ModeKeybind) error {
	for _, kb := range keybinds {
		if err := ms.gui.SetKeybinding("", kb.Key, gocui.ModNone, kb.Handler); err != nil {
			return err
		}
	}
	return nil
}

func (ms *ModeSwitcher) removeModeKeybindings(keybinds []ModeKeybind) error {
	for _, kb := range keybinds {
		if err := ms.gui.DeleteKeybinding("", kb.Key, gocui.ModNone); err != nil && err != gocui.ErrUnknownView {
			// Ignore errors if keybinding doesn't exist
		}
	}
	return nil
}

func (ms *ModeSwitcher) restoreAllMainKeybindings() error {
	av := ms.appView
	
	for _, viewName := range WeekdayNames {
		mainKeybindings := []struct {
			key     interface{}
			handler func(*gocui.Gui, *gocui.View) error
		}{
			{'a', func(g *gocui.Gui, v *gocui.View) error { return av.ShowNewEventPopup(g) }},
			{'e', func(g *gocui.Gui, v *gocui.View) error { return av.ShowEditEventPopup(g) }},
			{'c', func(g *gocui.Gui, v *gocui.View) error { return av.ShowColorPicker(g) }},
			{'g', func(g *gocui.Gui, v *gocui.View) error { return av.ShowGotoMode(g) }},
			{'h', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevDay(g); return nil }},
			{'l', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextDay(g); return nil }},
			{'j', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextTime(g); return nil }},
			{'k', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevTime(g); return nil }},
			{'T', func(g *gocui.Gui, v *gocui.View) error { av.JumpToToday(); return nil }},
			{'H', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevWeek(); return nil }},
			{'L', func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextWeek(); return nil }},
			{'d', func(g *gocui.Gui, v *gocui.View) error { av.DeleteEvent(g); return nil }},
			{'D', func(g *gocui.Gui, v *gocui.View) error { av.DeleteEvents(g); return nil }},
			{'y', func(g *gocui.Gui, v *gocui.View) error { av.CopyEvent(g); return nil }},
			{'p', func(g *gocui.Gui, v *gocui.View) error { return av.PasteEvent(g) }},
			{'u', func(g *gocui.Gui, v *gocui.View) error { return av.Undo(g) }},
			{'r', func(g *gocui.Gui, v *gocui.View) error { return av.Redo(g) }},
			{gocui.KeyCtrlN, func(g *gocui.Gui, v *gocui.View) error { return av.ChangeToNotepadView(g) }},
			{gocui.KeyCtrlS, func(g *gocui.Gui, v *gocui.View) error { return av.ShowOrHideSideView(g) }},
			{'?', func(g *gocui.Gui, v *gocui.View) error { return av.ShowKeybinds(g) }},
			{'q', func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }},
			{gocui.KeyArrowLeft, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevDay(g); return nil }},
			{gocui.KeyArrowRight, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextDay(g); return nil }},
			{gocui.KeyArrowDown, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToNextTime(g); return nil }},
			{gocui.KeyArrowUp, func(g *gocui.Gui, v *gocui.View) error { av.UpdateToPrevTime(g); return nil }},
		}
		
		for _, kb := range mainKeybindings {
			if err := ms.gui.SetKeybinding(viewName, kb.key, gocui.ModNone, kb.handler); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (ms *ModeSwitcher) getColorPickerConfig() ModeConfig {
	av := ms.appView
	
	return ModeConfig{
		Name:        ModeColorPicker,
		PopupViewName: "colorpicker",
		Keybinds: []ModeKeybind{
			{'r', func(g *gocui.Gui, v *gocui.View) error { return av.SelectColor(g, "Red") }},
			{'g', func(g *gocui.Gui, v *gocui.View) error { return av.SelectColor(g, "Green") }},
			{'y', func(g *gocui.Gui, v *gocui.View) error { return av.SelectColor(g, "Yellow") }},
			{'b', func(g *gocui.Gui, v *gocui.View) error { return av.SelectColor(g, "Blue") }},
			{'m', func(g *gocui.Gui, v *gocui.View) error { return av.SelectColor(g, "Magenta") }},
			{'c', func(g *gocui.Gui, v *gocui.View) error { return av.SelectColor(g, "Cyan") }},
			{'w', func(g *gocui.Gui, v *gocui.View) error { return av.SelectColor(g, "White") }},
			{gocui.KeyEsc, func(g *gocui.Gui, v *gocui.View) error { return av.CloseColorPicker(g) }},
		},
		OnExit: func(g *gocui.Gui) error {
			av.colorPickerActive = false
			av.colorPickerEvent = nil
			return nil
		},
	}
}

func (ms *ModeSwitcher) getGotoConfig() ModeConfig {
	av := ms.appView
	
	return ModeConfig{
		Name:        ModeGoto,
		PopupViewName: "gotomode",
		Keybinds: []ModeKeybind{
			{'t', func(g *gocui.Gui, v *gocui.View) error { return av.HandleGotoTime(g, v) }},
			{'d', func(g *gocui.Gui, v *gocui.View) error { return av.HandleGotoDate(g, v) }},
			{gocui.KeyEsc, func(g *gocui.Gui, v *gocui.View) error { return av.CloseGotoMode(g) }},
		},
		OnExit: func(g *gocui.Gui) error {
			av.gotoMode = false
			return nil
		},
	}
}