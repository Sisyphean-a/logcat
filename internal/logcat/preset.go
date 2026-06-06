package logcat

import "strings"

func MatchesH5Preset(entry LogEntry) bool {
	return entry.Tag == "chromium" && strings.Contains(entry.Message, "[H5]")
}
