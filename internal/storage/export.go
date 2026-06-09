package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	appstate "github.com/xiakn/logcat/internal/app"
)

func ExportVisibleLogs(items []appstate.LogViewItem) (string, error) {
	dir, err := exportDir()
	if err != nil {
		return "", err
	}
	return writeVisibleLogsFile(items, time.Now(), dir)
}

func writeVisibleLogsFile(items []appstate.LogViewItem, now time.Time, dir string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no_visible_logs_to_export")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("logcat-viewer-%s.tsv", now.Format("20060102-150405"))
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(renderVisibleLogsTSV(items)), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func exportDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads"), nil
}

func renderVisibleLogsTSV(items []appstate.LogViewItem) string {
	var builder strings.Builder
	builder.WriteString("时间\t级\t标签\t消息\t来源\n")
	for _, item := range items {
		builder.WriteString(tsvCell(item.Entry.TimeText))
		builder.WriteByte('\t')
		builder.WriteString(tsvCell(item.Entry.Level))
		builder.WriteByte('\t')
		builder.WriteString(tsvCell(item.Entry.Tag))
		builder.WriteByte('\t')
		builder.WriteString(tsvCell(item.Entry.Message))
		builder.WriteByte('\t')
		builder.WriteString(tsvCell(item.Entry.Source))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func tsvCell(value string) string {
	replacer := strings.NewReplacer("\t", " ", "\r", " ", "\n", " ")
	return replacer.Replace(value)
}
