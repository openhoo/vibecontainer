package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openhoo/vibecontainer/internal/config"
	"github.com/openhoo/vibecontainer/internal/domain"
)

const managedLabel = "com.openhoo.vibecontainer.managed=true"

type Compose struct {
	runner Runner
}

func NewCompose(r Runner) *Compose {
	return &Compose{runner: r}
}

func (c *Compose) Up(ctx context.Context, stack string) error {
	_, stderr, err := c.runner.Run(ctx, "docker", c.args(stack, "up", "-d")...)
	if err != nil {
		return fmt.Errorf("compose up failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	return nil
}

func (c *Compose) Stop(ctx context.Context, stack string) error {
	_, stderr, err := c.runner.Run(ctx, "docker", c.args(stack, "stop")...)
	if err != nil {
		return fmt.Errorf("compose stop failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	return nil
}

func (c *Compose) Restart(ctx context.Context, stack string) error {
	_, stderr, err := c.runner.Run(ctx, "docker", c.args(stack, "restart")...)
	if err != nil {
		return fmt.Errorf("compose restart failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	return nil
}

func (c *Compose) Down(ctx context.Context, stack string) error {
	_, stderr, err := c.runner.Run(ctx, "docker", c.args(stack, "down", "--remove-orphans")...)
	if err != nil {
		return fmt.Errorf("compose down failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	return nil
}

func (c *Compose) Logs(ctx context.Context, stack, service string, follow bool) (string, error) {
	args := c.args(stack, "logs")
	if follow {
		args = append(args, "--follow")
	}
	if service != "" {
		args = append(args, service)
	}
	stdout, stderr, err := c.runner.Run(ctx, "docker", args...)
	if err != nil {
		return "", fmt.Errorf("compose logs failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	return stdout, nil
}

func (c *Compose) Status(ctx context.Context, stack string) ([]domain.ServiceStatus, error) {
	stdout, stderr, err := c.runner.Run(ctx, "docker", c.args(stack, "ps", "--format", "json")...)
	if err != nil {
		return nil, fmt.Errorf("compose status failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []domain.ServiceStatus{}, nil
	}
	lines := strings.Split(stdout, "\n")
	statuses := make([]domain.ServiceStatus, 0, len(lines))
	for _, line := range lines {
		var s domain.ServiceStatus
		if err := json.Unmarshal([]byte(line), &s); err != nil {
			return nil, fmt.Errorf("parse compose ps output: %w", err)
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

type ManagedContainer struct {
	Name   string
	State  string
	Labels map[string]string
}

func (c *Compose) ListManagedContainers(ctx context.Context) ([]ManagedContainer, error) {
	stdout, stderr, err := c.runner.Run(
		ctx,
		"docker",
		"ps",
		"-a",
		"--filter",
		"label="+managedLabel,
		"--format",
		"{{json .}}",
	)
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %w\n%s", err, strings.TrimSpace(stderr))
	}
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return []ManagedContainer{}, nil
	}
	lines := strings.Split(stdout, "\n")
	out := make([]ManagedContainer, 0, len(lines))
	for _, line := range lines {
		var item struct {
			Names  string `json:"Names"`
			State  string `json:"State"`
			Labels string `json:"Labels"`
		}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, fmt.Errorf("parse docker ps output: %w", err)
		}
		out = append(out, ManagedContainer{
			Name:   item.Names,
			State:  item.State,
			Labels: parseLabelString(item.Labels),
		})
	}
	return out, nil
}

func (c *Compose) args(stack string, cmd ...string) []string {
	args := []string{
		"compose",
		"-p",
		stack,
		"-f",
		config.RunComposePath(stack),
		"--env-file",
		config.RunEnvPath(stack),
	}
	args = append(args, cmd...)
	return args
}

func parseLabelString(s string) map[string]string {
	labels := map[string]string{}
	if strings.TrimSpace(s) == "" {
		return labels
	}
	parts := strings.Split(s, ",")
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		labels[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return labels
}
