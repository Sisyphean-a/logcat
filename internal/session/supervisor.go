package session

import (
	"context"

	"github.com/xiakn/logcat/internal/logcat"
)

type Config struct {
	DeviceID    string
	PackageName string
	ProcessName string
	AllowedPIDs []int
}

type Event struct {
	Entry   *logcat.LogEntry
	Problem error
}

type Handle struct {
	events <-chan Event
}

func NewHandle(events <-chan Event) Handle {
	return Handle{events: events}
}

func (h Handle) Events() <-chan Event {
	return h.events
}

type Source interface {
	Stream(ctx context.Context, cfg Config) (<-chan string, <-chan error)
}

type Supervisor struct {
	source Source
}

func NewSupervisor(source Source) Supervisor {
	return Supervisor{source: source}
}

func (s Supervisor) Start(ctx context.Context, cfg Config) (Handle, error) {
	events := make(chan Event, 16)
	lines, errs := s.source.Stream(ctx, cfg)

	go s.forward(ctx, cfg, lines, errs, events)

	return NewHandle(events), nil
}

func (s Supervisor) forward(
	ctx context.Context,
	cfg Config,
	lines <-chan string,
	errs <-chan error,
	events chan<- Event,
) {
	defer close(events)

	for line := range lines {
		entry, err := logcat.ParseThreadtimeLine(cfg.DeviceID, line)
		if err != nil {
			events <- Event{Problem: err}
			continue
		}
		if logcat.MatchesH5Preset(entry) && allowPID(cfg, entry.PID) {
			events <- Event{Entry: &entry}
		}
	}

	select {
	case <-ctx.Done():
		return
	case err, ok := <-errs:
		if ok && err != nil {
			events <- Event{Problem: err}
		}
	default:
	}
}

func allowPID(cfg Config, pid int) bool {
	if len(cfg.AllowedPIDs) == 0 {
		return true
	}

	for _, allowed := range cfg.AllowedPIDs {
		if allowed == pid {
			return true
		}
	}

	return false
}
