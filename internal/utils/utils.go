package utils

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func DurationToHeight(d float64) int {
	return int(d * 2)
}

func FormatDate(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
}

func FormatHourFromTime(t time.Time) string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}

func FormatHour(hour, minute int) string {
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

func TimeToPosition(t time.Time, s string) int {
	timeStr := FormatHourFromTime(t)
	lines := strings.Split(s, "\n")

	for i, v := range lines {
		if strings.Contains(v, timeStr) {
			return i
		}
	}

	return -1
}

// TimeToPositionWithViewport calculates the position of a time within the viewport
func TimeToPositionWithViewport(t time.Time, viewportStart int) int {
	// Convert time to slot index (0 = 00:00, 1 = 00:30, etc.)
	hour := t.Hour()
	minute := t.Minute()
	
	// Round to nearest half hour
	if minute >= 30 {
		minute = 30
	} else {
		minute = 0
	}
	
	slotIndex := hour*2 + minute/30
	
	// Calculate position relative to viewport
	position := slotIndex - viewportStart
	
	// Debug logging
	if f, err := os.OpenFile("/tmp/chronos_utils_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "TimeToPositionWithViewport: time=%s, hour=%d, minute=%d, slotIndex=%d, viewportStart=%d, position=%d\n", 
			t.Format("15:04"), hour, minute, slotIndex, viewportStart, position)
		f.Close()
	}
	
	// Return -1 if time is outside the viewport
	if position < 0 {
		return -1
	}
	
	return position
}

func ValidateTime(value string) bool {
	regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$`)

	if !regex.MatchString(value) {
		return false
	}

	parts := strings.Split(value, " ")
	dateParts := strings.Split(parts[0], "-")
	timeParts := strings.Split(parts[1], ":")

	year, err := strconv.Atoi(dateParts[0])
	if err != nil || year <= 0 {
		return false
	}

	month, err := strconv.Atoi(dateParts[1])
	if err != nil || month <= 0 || month > 12 {
		return false
	}

	day, err := strconv.Atoi(dateParts[2])
	if err != nil || day <= 1 || day > 31 {
		return false
	}

	hours, err := strconv.Atoi(timeParts[0])
	if err != nil || hours < 0 || hours > 23 {
		return false
	}

	minutes, err := strconv.Atoi(timeParts[1])
	if err != nil || (minutes != 0 && minutes != 30) {
		return false
	}

	_, err = time.Parse("2006-01-02 15:04", value)
	return err == nil
}

func ValidateName(value string) bool {
	if value == "" {
		return false
	}

	return true
}

func ValidateNumber(value string) bool {
	n, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	if n <= 0 {
		return false
	}

	return true
}

func ValidateDuration(value string) bool {
	duration, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}

	if duration <= 0.0 || duration > 24.0 {
		return false
	}

	if math.Mod(duration, 0.5) != 0 {
		return false
	}

	return true
}

func ValidateHourMinute(value string) bool {
	regex := regexp.MustCompile(`^\d{2}$`)
	if !regex.MatchString(value) {
		return false
	}

	hour, err := strconv.Atoi(value)
	if err != nil {
		return false
	}

	if hour < 0 || hour > 23 {
		return false
	}

	return true
}

func ValidateDate(value string) bool {
	regex := regexp.MustCompile(`^\d{8}$`)
	if !regex.MatchString(value) {
		return false
	}

	if len(value) != 8 {
		return false
	}

	year, err := strconv.Atoi(value[:4])
	if err != nil || year < 1900 || year > 2100 {
		return false
	}

	month, err := strconv.Atoi(value[4:6])
	if err != nil || month < 1 || month > 12 {
		return false
	}

	day, err := strconv.Atoi(value[6:8])
	if err != nil || day < 1 || day > 31 {
		return false
	}

	_, err = time.Parse("20060102", value)
	return err == nil
}

func ValidateEventDate(value string) bool {
	regex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !regex.MatchString(value) {
		return false
	}

	parts := strings.Split(value, "-")
	if len(parts) != 3 {
		return false
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 1900 || year > 2100 {
		return false
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return false
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil || day < 1 || day > 31 {
		return false
	}

	_, err = time.Parse("2006-01-02", value)
	return err == nil
}

func ValidateEventTime(value string) bool {
	regex := regexp.MustCompile(`^\d{2}:\d{2}$`)
	if !regex.MatchString(value) {
		return false
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return false
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil || hours < 0 || hours > 23 {
		return false
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil || (minutes != 0 && minutes != 30) {
		return false
	}

	return true
}

func ValidateOptionalEventDate(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	if value == "t" {
		return true
	}
	return ValidateEventDate(value)
}

func ValidateOptionalDate(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	if value == "t" {
		return true
	}
	return ValidateDate(value)
}

func ValidateOptionalEventTime(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return true
	}
	return ValidateEventTime(value)
}

func ValidateFlexibleHour(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false // Hour is required
	}

	// Try to parse as float (e.g., 14.5 for 14:30)
	if floatHour, err := strconv.ParseFloat(value, 64); err == nil {
		// Check if hour is in valid range
		if floatHour < 0 || floatHour >= 24 {
			return false
		}
		// Check if fractional part is valid (only .0 or .5 allowed)
		fractional := floatHour - math.Floor(floatHour)
		if fractional != 0.0 && fractional != 0.5 {
			return false
		}
		return true
	}

	// If not a float, must be an integer
	if hour, err := strconv.Atoi(value); err == nil {
		return hour >= 0 && hour <= 23
	}

	return false
}

func ValidateFrequency(value string) bool {
	value = strings.TrimSpace(value)
	
	// Accept 'w' for weekdays
	if value == "w" || value == "W" {
		return true
	}
	
	// Otherwise, must be a positive number
	n, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	
	if n <= 0 {
		return false
	}
	
	return true
}

// IsWeekday returns true if the given time is a weekday (Monday-Friday)
func IsWeekday(t time.Time) bool {
	day := t.Weekday()
	return day >= time.Monday && day <= time.Friday
}
