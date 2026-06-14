package app

import (
	"strings"

	"github.com/xiakn/logcat/internal/logcat"
)

func searchLowerText(entry logcat.LogEntry) string {
	return strings.ToLower(entry.Tag + "\n" + entry.Message)
}
