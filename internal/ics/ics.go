package ics

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/samuelstranges/chronos/internal/calendar"
)

const (
	// iCalendar version
	VERSION = "2.0"
	// Product identifier for Chronos
	PRODID = "-//Chronos//Chronos Calendar Application//EN"
	// Calendar method
	METHOD = "PUBLISH"
)

// ICSExporter handles export of events to iCalendar format
type ICSExporter struct{}

// NewICSExporter creates a new ICS exporter
func NewICSExporter() *ICSExporter {
	return &ICSExporter{}
}

// ExportEvents exports a slice of events to iCalendar format
func (e *ICSExporter) ExportEvents(events []*calendar.Event) string {
	var builder strings.Builder

	// Write calendar header
	builder.WriteString("BEGIN:VCALENDAR\r\n")
	builder.WriteString(fmt.Sprintf("VERSION:%s\r\n", VERSION))
	builder.WriteString(fmt.Sprintf("PRODID:%s\r\n", PRODID))
	builder.WriteString("CALSCALE:GREGORIAN\r\n")
	builder.WriteString(fmt.Sprintf("METHOD:%s\r\n", METHOD))

	// Write events
	for _, event := range events {
		builder.WriteString(e.formatEvent(event))
	}

	// Write calendar footer
	builder.WriteString("END:VCALENDAR\r\n")

	return builder.String()
}



// formatEvent formats a single event as a VEVENT component
func (e *ICSExporter) formatEvent(event *calendar.Event) string {
	var builder strings.Builder

	builder.WriteString("BEGIN:VEVENT\r\n")
	
	// UID - Generate unique identifier using event ID
	builder.WriteString(fmt.Sprintf("UID:chronos-event-%d@chronos.local\r\n", event.Id))
	
	// DTSTAMP - Creation/modification timestamp (current time in UTC)
	now := time.Now().UTC()
	builder.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", now.Format("20060102T150405Z")))
	
	// DTCREATED - Event creation timestamp (use current time in UTC)
	builder.WriteString(fmt.Sprintf("DTCREATED:%s\r\n", now.Format("20060102T150405Z")))
	
	// DTSTART - Event start time (UTC format for compatibility)
	builder.WriteString(fmt.Sprintf("DTSTART:%s\r\n", event.Time.Format("20060102T150405Z")))
	
	// DTEND - Event end time (UTC format for compatibility)
	utcEndTime := event.Time.Add(time.Duration(event.DurationHour * float64(time.Hour)))
	builder.WriteString(fmt.Sprintf("DTEND:%s\r\n", utcEndTime.Format("20060102T150405Z")))
	
	// SUMMARY - Event title (required)
	builder.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", e.escapeText(event.Name)))
	
	// DESCRIPTION - Event description (optional)
	if event.Description != "" {
		builder.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", e.escapeText(event.Description)))
	}
	
	// LOCATION - Event location (optional)
	if event.Location != "" {
		builder.WriteString(fmt.Sprintf("LOCATION:%s\r\n", e.escapeText(event.Location)))
	}
	
	

	builder.WriteString("END:VEVENT\r\n")

	return builder.String()
}

// escapeText escapes special characters in text fields according to RFC 5545
func (e *ICSExporter) escapeText(text string) string {
	// Replace special characters according to RFC 5545
	text = strings.ReplaceAll(text, "\\", "\\\\") // Backslash must be escaped first
	text = strings.ReplaceAll(text, ",", "\\,")   // Comma
	text = strings.ReplaceAll(text, ";", "\\;")   // Semicolon
	text = strings.ReplaceAll(text, "\n", "\\n")  // Newline
	text = strings.ReplaceAll(text, "\r", "")     // Remove carriage returns
	return text
}

// ExportToFile writes events to an iCalendar file
func (e *ICSExporter) ExportToFile(events []*calendar.Event, filename string) error {
	icsContent := e.ExportEvents(events)
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	_, err = file.WriteString(icsContent)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	return nil
}