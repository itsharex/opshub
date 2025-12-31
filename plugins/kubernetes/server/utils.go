package server

import (
	"strconv"
	"time"
)

// calculateAge 计算资源年龄
func calculateAge(creationTime time.Time) string {
	duration := time.Since(creationTime)

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes())

	if days > 0 {
		if days == 1 {
			return "1d"
		}
		return strconv.Itoa(days) + "d"
	}

	if hours > 0 {
		if hours == 1 {
			return "1h"
		}
		return strconv.Itoa(hours) + "h"
	}

	if minutes > 0 {
		if minutes == 1 {
			return "1m"
		}
		return strconv.Itoa(minutes) + "m"
	}

	return "<1m"
}
