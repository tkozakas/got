package telegram

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func ParseDuration(s string) (time.Duration, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	re := regexp.MustCompile(`(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?`)
	matches := re.FindStringSubmatch(s)

	if matches == nil || matches[0] == "" {
		return 0, fmt.Errorf("invalid duration format")
	}

	var duration time.Duration

	if matches[1] != "" {
		days, _ := strconv.Atoi(matches[1])
		duration += time.Duration(days) * 24 * time.Hour
	}
	if matches[2] != "" {
		hours, _ := strconv.Atoi(matches[2])
		duration += time.Duration(hours) * time.Hour
	}
	if matches[3] != "" {
		mins, _ := strconv.Atoi(matches[3])
		duration += time.Duration(mins) * time.Minute
	}
	if matches[4] != "" {
		secs, _ := strconv.Atoi(matches[4])
		duration += time.Duration(secs) * time.Second
	}

	if duration == 0 {
		return 0, fmt.Errorf("invalid duration format")
	}

	return duration, nil
}
