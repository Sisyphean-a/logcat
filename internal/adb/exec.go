package adb

import (
	"context"
	"io"
	"os/exec"
)

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func (ExecRunner) Start(
	ctx context.Context,
	name string,
	args ...string,
) (io.ReadCloser, <-chan error, error) {
	command := exec.CommandContext(ctx, name, args...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := command.Start(); err != nil {
		return nil, nil, err
	}

	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
		close(done)
	}()

	return stdout, done, nil
}
