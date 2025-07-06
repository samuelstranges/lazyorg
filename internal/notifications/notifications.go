package notifications

import (
	"fmt"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/samuelstranges/chronos/internal/calendar"
	"github.com/samuelstranges/chronos/internal/config"
)

// NotificationManager handles desktop notifications for events
type NotificationManager struct {
	config *config.Config
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(cfg *config.Config) *NotificationManager {
	return &NotificationManager{
		config: cfg,
	}
}

// IsEnabled returns true if notifications are enabled
func (nm *NotificationManager) IsEnabled() bool {
	return config.IsNotificationsEnabled(nm.config)
}

// GetNotificationMinutes returns the configured notification minutes
func (nm *NotificationManager) GetNotificationMinutes() int {
	return config.GetNotificationMinutes(nm.config)
}

// SendEventNotification sends a notification for an upcoming event
func (nm *NotificationManager) SendEventNotification(event *calendar.Event) error {
	if !nm.IsEnabled() {
		return nil // Notifications disabled
	}

	title := "Upcoming Event"
	message := fmt.Sprintf("%s\n%s", event.Name, formatEventTime(event))
	
	// Add location if present
	if event.Location != "" {
		message += fmt.Sprintf("\nLocation: %s", event.Location)
	}
	
	// Add description if present (truncated)
	if event.Description != "" {
		desc := event.Description
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		message += fmt.Sprintf("\n%s", desc)
	}

	return beeep.Notify(title, message, "")
}

// SendTestNotification sends a test notification
func (nm *NotificationManager) SendTestNotification() error {
	title := "Chronos Test Notification"
	message := "Desktop notifications are working correctly!"
	return beeep.Notify(title, message, "")
}

// formatEventTime formats the event time for display in notifications
func formatEventTime(event *calendar.Event) string {
	// Convert UTC stored time to local time for display
	eventTimeLocal := event.Time.Local()
	startTime := eventTimeLocal.Format("3:04 PM")
	endTime := eventTimeLocal.Add(time.Duration(event.DurationHour) * time.Hour).Format("3:04 PM")
	return fmt.Sprintf("%s - %s", startTime, endTime)
}

// ShouldNotify checks if an event should trigger a notification
func (nm *NotificationManager) ShouldNotify(event *calendar.Event) bool {
	if !nm.IsEnabled() {
		return false
	}

	now := time.Now()
	notificationTime := event.Time.Add(-time.Duration(nm.GetNotificationMinutes()) * time.Minute)
	
	// Check if notification time is within the next minute
	return notificationTime.After(now) && notificationTime.Before(now.Add(time.Minute))
}

// NotificationScheduler handles background notification scheduling
type NotificationScheduler struct {
	manager       *NotificationManager
	database      EventDatabase
	ticker        *time.Ticker
	stopChan      chan struct{}
	notifiedEvents map[int]time.Time // Track which events we've already notified about
}

// EventDatabase interface for getting events from database
type EventDatabase interface {
	GetEventsByDateRange(startDate, endDate time.Time) ([]*calendar.Event, error)
}

// NewNotificationScheduler creates a new notification scheduler
func NewNotificationScheduler(manager *NotificationManager, db EventDatabase) *NotificationScheduler {
	return &NotificationScheduler{
		manager:        manager,
		database:       db,
		notifiedEvents: make(map[int]time.Time),
	}
}

// Start begins the background notification scheduler
func (ns *NotificationScheduler) Start() {
	if !ns.manager.IsEnabled() {
		return // Don't start if notifications are disabled
	}

	ns.ticker = time.NewTicker(30 * time.Second) // Check every 30 seconds
	ns.stopChan = make(chan struct{})

	go ns.run()
}

// Stop stops the notification scheduler
func (ns *NotificationScheduler) Stop() {
	if ns.ticker != nil {
		ns.ticker.Stop()
	}
	if ns.stopChan != nil {
		close(ns.stopChan)
	}
}

// run is the main loop for the notification scheduler
func (ns *NotificationScheduler) run() {
	for {
		select {
		case <-ns.ticker.C:
			ns.checkUpcomingEvents()
		case <-ns.stopChan:
			return
		}
	}
}

// checkUpcomingEvents checks for upcoming events and sends notifications
func (ns *NotificationScheduler) checkUpcomingEvents() {
	now := time.Now()
	
	// Get events for the next 24 hours
	endTime := now.Add(24 * time.Hour)
	events, err := ns.database.GetEventsByDateRange(now, endTime)
	if err != nil {
		return // Silently fail - don't spam errors
	}

	// Clean up old notified events (older than 24 hours)
	ns.cleanupOldNotifications(now)

	// Check each event for notification
	for _, event := range events {
		if ns.shouldNotifyNow(event, now) {
			err := ns.manager.SendEventNotification(event)
			if err == nil {
				// Mark as notified
				ns.notifiedEvents[event.Id] = now
			}
		}
	}
}

// shouldNotifyNow checks if we should notify about an event right now
func (ns *NotificationScheduler) shouldNotifyNow(event *calendar.Event, now time.Time) bool {
	// Check if we've already notified about this event recently
	if lastNotified, exists := ns.notifiedEvents[event.Id]; exists {
		// Don't notify again if we notified within the last hour
		if now.Sub(lastNotified) < time.Hour {
			return false
		}
	}

	// CRITICAL: Convert event time to local timezone for comparison since event.Time is stored in UTC
	// but 'now' is in local timezone
	eventTimeLocal := event.Time.Local()
	notificationTime := eventTimeLocal.Add(-time.Duration(ns.manager.GetNotificationMinutes()) * time.Minute)
	
	// Notify if the notification time is within the last 30 seconds to next 30 seconds
	return notificationTime.After(now.Add(-30*time.Second)) && notificationTime.Before(now.Add(30*time.Second))
}

// cleanupOldNotifications removes old entries from the notified events map
func (ns *NotificationScheduler) cleanupOldNotifications(now time.Time) {
	for eventID, notifiedTime := range ns.notifiedEvents {
		if now.Sub(notifiedTime) > 24*time.Hour {
			delete(ns.notifiedEvents, eventID)
		}
	}
}