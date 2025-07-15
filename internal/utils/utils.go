package utils

import (
	"os"
	"time"
)

// TimeNow returns current time
func TimeNow() time.Time {
	return time.Now()
}

// ParseDate parses date string in YYYY-MM-DD format
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// CreateDirIfNotExists creates directory if it doesn't exist
func CreateDirIfNotExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// CreateFile creates a new file
func CreateFile(filepath string) (*os.File, error) {
	return os.Create(filepath)
}
