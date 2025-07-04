package views

import (
	"os"
)

// appendDebugLog writes debug information to a log file if debug mode is enabled
func (av *AppView) appendDebugLog(filename, content string) {
	if !av.DebugMode {
		return
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	file.WriteString(content)
}