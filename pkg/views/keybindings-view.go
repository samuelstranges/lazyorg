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
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " Navigation:")
	fmt.Fprintln(v, " h/l or ←/→   - Previous/Next day")
	fmt.Fprintln(v, " H/L          - Previous/Next week")
	fmt.Fprintln(v, " j/k or ↓/↑   - Move time cursor down/up")
	fmt.Fprintln(v, " T            - Jump to today")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " Goto:")
	fmt.Fprintln(v, " g            - Goto date/time")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " Events:")
	fmt.Fprintln(v, " a            - Add new event")
	fmt.Fprintln(v, " e            - Edit event")
	fmt.Fprintln(v, " c            - Color picker")
	fmt.Fprintln(v, " w            - Jump to next event")
	fmt.Fprintln(v, " b            - Jump to previous event")
	fmt.Fprintln(v, " /            - Search events (name/desc/loc)")
	fmt.Fprintln(v, " n            - Next search match")
	fmt.Fprintln(v, " N            - Previous search match")
	fmt.Fprintln(v, " Esc          - Clear search")
	fmt.Fprintln(v, " y            - Copy event")
	fmt.Fprintln(v, " p            - Paste event")
	fmt.Fprintln(v, " d            - Delete event")
	fmt.Fprintln(v, " D            - Delete events with same name")
	fmt.Fprintln(v, " u            - Undo last action")
	fmt.Fprintln(v, " r            - Redo last undone action")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " Color Picker:")
	fmt.Fprintln(v, " c            - Color picker (r/g/y/b/m/c/w)")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " View Controls:")
	fmt.Fprintln(v, " Ctrl+s       - Show/Hide side view")
	fmt.Fprintln(v, " ?            - Toggle this help")
	fmt.Fprintln(v, "")
	fmt.Fprintln(v, " Global:")
	fmt.Fprintln(v, " q            - Quit")
	g.SetViewOnTop("keybinds")
	return nil
}
