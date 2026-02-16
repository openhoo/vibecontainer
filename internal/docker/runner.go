package docker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, cmd string, args ...string) (string, string, error)
}

type ExecRunner struct{}

func NewExecRunner() *ExecRunner { return &ExecRunner{} }

func (r *ExecRunner) Run(ctx context.Context, cmd string, args ...string) (string, string, error) {
	c := exec.CommandContext(ctx, cmd, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		sub := ""
		if len(args) > 0 {
			sub = " " + args[0]
		}
		return stdout.String(), stderr.String(), fmt.Errorf("%s%s: %w", cmd, sub, err)
	}
	return stdout.String(), stderr.String(), nil
}
