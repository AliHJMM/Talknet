package utils

import (
	"fmt"
	"time"
)

// TimeAgo returns a "time ago" style string (e.g., "5 minutes ago").
func TimeAgo(t time.Time) string {
    diff := time.Since(t)
    switch {
    case diff < time.Minute:
        return "just now"
    case diff < time.Hour:
        return fmt.Sprintf("%d minute ago", int(diff.Minutes()))
    case diff < 24*time.Hour:
        return fmt.Sprintf("%d hour ago", int(diff.Hours()))
    default:
        // Or return a short date
        return t.Format("2006-01-02 15:04:05")
    }
}
