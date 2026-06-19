package logcat

import (
	"fmt"
	"strconv"
	"strings"
)

type LogEntry struct {
	DeviceID string
	TimeText string
	PID      int
	TID      int
	Level    string
	Tag      string
	Source   string
	Message  string
	Raw      string
}

func ParseThreadtimeLine(deviceID, line string) (LogEntry, error) {
	timeText, pidText, tidText, level, rest, ok := splitThreadtimeFields(line)
	if !ok {
		return LogEntry{}, fmt.Errorf("invalid threadtime line: %s", line)
	}

	tag, messageText, ok := splitTagAndMessage(rest)
	if !ok {
		return LogEntry{}, fmt.Errorf("invalid threadtime line: %s", line)
	}

	message, source := parseChromiumConsoleMessage(messageText)
	return LogEntry{
		DeviceID: deviceID,
		TimeText: timeText,
		PID:      mustAtoi(pidText),
		TID:      mustAtoi(tidText),
		Level:    level,
		Tag:      tag,
		Source:   source,
		Message:  message,
		Raw:      line,
	}, nil
}

func splitThreadtimeFields(line string) (string, string, string, string, string, bool) {
	dateText, next, ok := nextField(line, 0)
	if !ok {
		return "", "", "", "", "", false
	}
	clockText, next, ok := nextField(line, next)
	if !ok {
		return "", "", "", "", "", false
	}
	pidText, next, ok := nextField(line, next)
	if !ok {
		return "", "", "", "", "", false
	}
	tidText, next, ok := nextField(line, next)
	if !ok {
		return "", "", "", "", "", false
	}
	level, next, ok := nextField(line, next)
	if !ok {
		return "", "", "", "", "", false
	}
	rest := normalizeThreadtimeRest(line[next:])
	if rest == "" {
		return "", "", "", "", "", false
	}
	return dateText + " " + clockText, pidText, tidText, level, rest, true
}

func splitTagAndMessage(value string) (string, string, bool) {
	index := strings.Index(value, ": ")
	if index == -1 {
		return "", "", false
	}

	tag := strings.TrimSpace(value[:index])
	message := strings.TrimSpace(value[index+2:])
	if tag == "" {
		return "", "", false
	}
	return tag, message, true
}

func mustAtoi(value string) int {
	number, _ := strconv.Atoi(value)
	return number
}

func parseChromiumConsoleMessage(value string) (string, string) {
	const sourcePrefix = `", source: `
	end := strings.LastIndex(value, sourcePrefix)
	if end == -1 {
		return value, ""
	}

	start := strings.IndexByte(value, '"')
	if start == -1 || end <= start {
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

func nextField(line string, start int) (string, int, bool) {
	start = skipThreadtimeSpaces(line, start)
	if start >= len(line) {
		return "", start, false
	}

	end := start
	for end < len(line) && !isThreadtimeSpace(line[end]) {
		end++
	}
	return line[start:end], end, true
}

func skipThreadtimeSpaces(line string, start int) int {
	for start < len(line) && isThreadtimeSpace(line[start]) {
		start++
	}
	return start
}

func isThreadtimeSpace(char byte) bool {
	switch char {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

func normalizeThreadtimeRest(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if !needsSpaceNormalization(trimmed) {
		return trimmed
	}

	var builder strings.Builder
	builder.Grow(len(trimmed))
	spacePending := false
	for index := 0; index < len(trimmed); index++ {
		if isThreadtimeSpace(trimmed[index]) {
			spacePending = true
			continue
		}
		if spacePending && builder.Len() > 0 {
			builder.WriteByte(' ')
		}
		spacePending = false
		builder.WriteByte(trimmed[index])
	}
	return builder.String()
}

func needsSpaceNormalization(value string) bool {
	spaceSeen := false
	for index := 0; index < len(value); index++ {
		if !isThreadtimeSpace(value[index]) {
			spaceSeen = false
			continue
		}
		if value[index] != ' ' || spaceSeen {
			return true
		}
		spaceSeen = true
	}
	return false
}
