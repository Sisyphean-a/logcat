package logcat

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LogEntry struct {
	DeviceID  string
	Timestamp time.Time
	TimeText  string
	PID       int
	TID       int
	Level     string
	Tag       string
	Source    string
	Message   string
	Raw       string
}

var threadtimePattern = regexp.MustCompile(
	`^(\d\d-\d\d \d\d:\d\d:\d\d\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEF])\s+([^:]+):\s(.*)$`,
)

func ParseThreadtimeLine(deviceID, line string) (LogEntry, error) {
	match := threadtimePattern.FindStringSubmatch(line)
	if match == nil {
		return LogEntry{}, fmt.Errorf("invalid threadtime line: %s", line)
	}

	message, source := parseChromiumConsoleMessage(match[6])
	return LogEntry{
		DeviceID:  deviceID,
		Timestamp: parseTimestamp(match[1]),
		TimeText:  match[1],
		PID:       mustAtoi(match[2]),
		TID:       mustAtoi(match[3]),
		Level:     match[4],
		Tag:       strings.TrimSpace(match[5]),
		Source:    source,
		Message:   message,
		Raw:       line,
	}, nil
}

func mustAtoi(value string) int {
	number, _ := strconv.Atoi(value)
	return number
}

func parseTimestamp(value string) time.Time {
	year := time.Now().Year()
	parsed, err := time.Parse("2006-01-02 15:04:05.000", fmt.Sprintf("%d-%s", year, value))
	if err != nil {
		return time.Time{}
	}

	return parsed
}

func parseChromiumConsoleMessage(value string) (string, string) {
	const sourcePrefix = `", source: `
	if !strings.Contains(value, sourcePrefix) {
		return value, ""
	}

	start := strings.Index(value, `"`)
	end := strings.LastIndex(value, sourcePrefix)
	if start == -1 || end == -1 || end <= start {
		return value, ""
	}

	message := value[start+1 : end]
	source := strings.TrimSuffix(value[end+len(sourcePrefix):], ")")
	if index := strings.LastIndex(source, " ("); index != -1 {
		source = source[:index]
	}
	source = strings.TrimPrefix(source, "http://127.0.0.1/")
	source = strings.TrimPrefix(source, "https://127.0.0.1/")
	source = strings.TrimPrefix(source, "http://localhost/")
	source = strings.TrimPrefix(source, "https://localhost/")
	return message, source
}
