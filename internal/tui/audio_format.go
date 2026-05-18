package tui

import "fmt"

func audioPosition(position, total int) string {
	if position < 0 {
		position = 0
	}
	if total > 0 {
		return formatAudioClock(position) + "/" + formatAudioClock(total)
	}
	return formatAudioClock(position)
}

func formatAudioClock(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	if hours > 0 {
		return fmt.Sprintf("%dh:%02dm:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
