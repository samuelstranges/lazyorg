package views

import (
	"fmt"
	"time"

	"github.com/samuelstranges/chronos/internal/utils"
	"github.com/jroimartin/gocui"
)

type TimeView struct {
	*BaseView
	Body   string
	Cursor int
	// Viewport management for dynamic scrolling
	ViewportStart int // Starting time slot (0 = 00:00, 1 = 00:30, etc.)
	MaxTimeSlots  int // Maximum number of time slots (48 for 24 hours)
}

func NewTimeView() *TimeView {
	tv := &TimeView{
		BaseView:      NewBaseView("time"),
		Cursor:        0,
		ViewportStart: 0,
		MaxTimeSlots:  48, // 24 hours * 2 slots per hour
	}

	return tv
}

func (tv *TimeView) Update(g *gocui.Gui) error {
	v, err := g.SetView(
		tv.Name,
		tv.X,
		tv.Y,
		tv.X+tv.W,
		tv.Y+tv.H,
	)

	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorGreen
	}

	tv.updateBody(v)

	return nil
}

func (tv *TimeView) updateBody(v *gocui.View) {
	tv.Body = ""

	// Calculate which time slots to show based on viewport
	// Reserve space for the day view border at the bottom
	visibleSlots := tv.H - 1  // Subtract 1 for the bottom border
	if visibleSlots < 1 {
		visibleSlots = 1
	}
	if visibleSlots > tv.MaxTimeSlots {
		visibleSlots = tv.MaxTimeSlots
	}

	// Ensure viewport doesn't go beyond available time slots
	if tv.ViewportStart+visibleSlots > tv.MaxTimeSlots {
		tv.ViewportStart = tv.MaxTimeSlots - visibleSlots
	}
	if tv.ViewportStart < 0 {
		tv.ViewportStart = 0
	}

	for i := 0; i < visibleSlots; i++ {
		slotIndex := tv.ViewportStart + i
		if slotIndex >= tv.MaxTimeSlots {
			break
		}

		hour := slotIndex / 2
		minute := (slotIndex % 2) * 30

		var timeStr string
		if minute == 0 {
			formattedHour := utils.FormatHour(hour, 0)
			timeStr = fmt.Sprintf(" %s - \n", formattedHour)
		} else {
			formattedHour := utils.FormatHour(hour, 30)
			timeStr = fmt.Sprintf(" %s \n", formattedHour)
		}

		// Show cursor if this is the cursor position
		if i == tv.Cursor {
			runes := []rune(timeStr)
			runes[0] = '>'
			timeStr = string(runes)
		}

		tv.Body += timeStr
	}

	v.Clear()
	fmt.Fprintln(v, tv.Body)
}

func (tv *TimeView) SetCursor(y int) {
	tv.Cursor = y
}

// AutoAdjustViewport automatically adjusts the viewport based on cursor position
func (tv *TimeView) AutoAdjustViewport(calendarTime time.Time) {
	visibleSlots := tv.GetVisibleSlots()
	
	// If we can show all time slots, start from the beginning
	if visibleSlots >= tv.MaxTimeSlots {
		tv.ViewportStart = 0
		return
	}
	
	// Calculate the calendar time slot for centering
	currentSlot := calendarTime.Hour()*2
	if calendarTime.Minute() >= 30 {
		currentSlot++
	}
	
	// Special handling for the last time slot (23:30) - do this FIRST
	// Position 23:30 comfortably visible, not at the bottom edge
	if currentSlot == tv.MaxTimeSlots-1 { // Only for 23:30 (slot 47)
		// We want 23:30 to appear 4-6 slots from the bottom for comfortable viewing
		// Work backwards: if we want 23:30 at position (visibleSlots - 5), 
		// then ViewportStart = currentSlot - (visibleSlots - 5)
		slotsFromBottom := 5  // Position 23:30 this many slots from the bottom
		if visibleSlots < 10 { // For very small viewports
			slotsFromBottom = 2
		}
		targetPosition := visibleSlots - slotsFromBottom
		
		tv.ViewportStart = currentSlot - targetPosition
		
		// Apply boundary checks
		if tv.ViewportStart < 0 {
			tv.ViewportStart = 0
		}
		// For 23:30, we know ViewportStart + visibleSlots will be > MaxTimeSlots
		// So we want the maximum ViewportStart that keeps 23:30 visible
		maxViewportStart := tv.MaxTimeSlots - visibleSlots
		if tv.ViewportStart > maxViewportStart {
			tv.ViewportStart = maxViewportStart
		}
	} else {
		// Normal centering logic for all other times
		tv.ViewportStart = currentSlot - visibleSlots/2
		
		// Ensure viewport doesn't go beyond bounds
		if tv.ViewportStart < 0 {
			tv.ViewportStart = 0
		}
		if tv.ViewportStart+visibleSlots > tv.MaxTimeSlots {
			tv.ViewportStart = tv.MaxTimeSlots - visibleSlots
		}
	}
}

// GetViewportStart returns the current viewport start position
func (tv *TimeView) GetViewportStart() int {
	return tv.ViewportStart
}

// GetVisibleSlots returns the number of visible time slots
func (tv *TimeView) GetVisibleSlots() int {
	// Reserve space for the day view border at the bottom
	visibleSlots := tv.H - 1  // Subtract 1 for the bottom border
	if visibleSlots < 1 {
		visibleSlots = 1
	}
	if visibleSlots > tv.MaxTimeSlots {
		visibleSlots = tv.MaxTimeSlots
	}
	return visibleSlots
}

// AdjustViewportForCursor adjusts the viewport to center around the cursor position
func (tv *TimeView) AdjustViewportForCursor() {
	visibleSlots := tv.GetVisibleSlots()
	
	// If we can show all time slots, start from the beginning
	if visibleSlots >= tv.MaxTimeSlots {
		tv.ViewportStart = 0
		return
	}
	
	// Calculate the absolute cursor position in the time grid
	absoluteCursorPosition := tv.ViewportStart + tv.Cursor
	
	// Try to center the viewport around the cursor position
	tv.ViewportStart = absoluteCursorPosition - visibleSlots/2
	
	// Ensure viewport doesn't go beyond bounds
	if tv.ViewportStart < 0 {
		tv.ViewportStart = 0
	}
	if tv.ViewportStart+visibleSlots > tv.MaxTimeSlots {
		tv.ViewportStart = tv.MaxTimeSlots - visibleSlots
	}
	
	// Update cursor position relative to new viewport
	tv.Cursor = absoluteCursorPosition - tv.ViewportStart
	
	// Ensure cursor is within visible range
	if tv.Cursor < 0 {
		tv.Cursor = 0
	}
	if tv.Cursor >= visibleSlots {
		tv.Cursor = visibleSlots - 1
	}
}

