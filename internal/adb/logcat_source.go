package adb

import (
	"bufio"
	"context"
	"io"

	"github.com/xiakn/logcat/internal/session"
)

type PipeRunner interface {
	Start(ctx context.Context, name string, args ...string) (io.ReadCloser, <-chan error, error)
}

type LogcatSource struct {
	runner  PipeRunner
	adbPath string
}

func NewLogcatSource(runner PipeRunner, adbPath string) LogcatSource {
	if adbPath == "" {
		adbPath = "adb"
	}

	return LogcatSource{
		runner:  runner,
		adbPath: adbPath,
	}
}

func (s LogcatSource) Stream(
	ctx context.Context,
	cfg session.Config,
) (<-chan string, <-chan error) {
	lines := make(chan string, 32)
	errs := make(chan error, 1)

	reader, done, err := s.runner.Start(ctx, s.adbPath, buildLogcatArgs(cfg.DeviceID)...)
	if err != nil {
		errs <- err
		close(lines)
		close(errs)
		return lines, errs
	}

	go scanLines(reader, done, lines, errs)

	return lines, errs
}

func buildLogcatArgs(deviceID string) []string {
	return []string{
		"-s",
		deviceID,
		"logcat",
		"-v",
		"threadtime",
	}
}

func scanLines(
	reader io.ReadCloser,
	done <-chan error,
	lines chan<- string,
	errs chan<- error,
) {
	defer close(lines)
	defer close(errs)
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		errs <- err
		return
	}

	if err := <-done; err != nil {
		errs <- err
	}
}
