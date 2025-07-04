package views

import (
	"fmt"
	"github.com/jroimartin/gocui"
)

type KeybindsView struct {
	*BaseView
	IsVisible bool
}

func NewKeybindsView() *KeybindsView {
	return &KeybindsView{
		BaseView:  NewBaseView("keybinds"),
		IsVisible: false,
	}
}

// GetRequiredHeight returns the number of lines needed for all keybinding content
func (kbv *KeybindsView) GetRequiredHeight() int {
	// Count all the content lines (including empty lines and borders)
	lines := []string{
		" q           - Quit",
		"",
		" Navigation:",
		" h/l or ←/→  - Previous/Next day",
		" H/L         - Previous/Next week",
		" j/k or ↓/↑  - Move time cursor down/up",
		" t           - Jump to today",
		" g           - Goto date/time form",
		" w           - Jump to next event",
		" b           - Jump to previous event",
		"",
		"Advanced Search",
		" /           - Search events (name/desc/loc)",
		" n           - Next search match",
		" N           - Previous search match",
		" Esc         - Clear search",
		"",
		" Event Management:",
		" a           - Add new event",
		" e           - Edit event",
		" c           - Color picker",
		" y           - Copy event",
		" p           - Paste event",
		" d           - Delete event",
		" D           - Delete all events w/ same name",
		"",
		" Undo Buffer:",
		" u           - Undo last action",
		" r           - Redo last undone action",
		"",
	}
	return len(lines) + 2 // +2 for top and bottom borders
}
func (kbv *KeybindsView) Update(g *gocui.Gui) error {
	if !kbv.IsVisible {
		return nil
	}
	v, err := g.SetView(
		kbv.Name,
		kbv.X,
		kbv.Y,
		kbv.X+kbv.W,
		kbv.Y+kbv.H,
	)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Keybindings "
	}
	v.Clear()
	fmt.Fprintln(v, " q           - Quit")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " Navigation:")
	fmt.Fprintln(v, " h/l or ←/→  - Previous/Next day")
	fmt.Fprintln(v, " H/L         - Previous/Next week")
	fmt.Fprintln(v, " j/k or ↓/↑  - Move time cursor down/up")
	fmt.Fprintln(v, " t           - Jump to today")
	fmt.Fprintln(v, " g           - Goto date/time form")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " Events:")
	fmt.Fprintln(v, " a           - Add new event")
	fmt.Fprintln(v, " e           - Edit event")
	fmt.Fprintln(v, " c           - Color picker")
	fmt.Fprintln(v, " w           - Jump to next event")
	fmt.Fprintln(v, " b           - Jump to previous event")
	fmt.Fprintln(v, " /           - Search events (name/desc/loc)")
	fmt.Fprintln(v, " n           - Next search match")
	fmt.Fprintln(v, " N           - Previous search match")
	fmt.Fprintln(v, " Esc         - Clear search")
	fmt.Fprintln(v, " y           - Copy event")
	fmt.Fprintln(v, " p           - Paste event")
	fmt.Fprintln(v, " d           - Delete event")
	fmt.Fprintln(v, " D           - Delete events with same name")
	fmt.Fprintln(v, " c           - Change event color")
	fmt.Fprintln(v, " u           - Undo last action")
	fmt.Fprintln(v, " r           - Redo last undone action")
	g.SetViewOnTop("keybinds")
	return nil
}
