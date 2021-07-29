package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
)

type StartComponentFn func(*testing.T, *app.Config) *Component

// StartPostgreSQL starts a fresh PostgreSQL instance in a docker container.
// It updates PostgreSQL config with a chosen port
func StartPostgreSQL(t *testing.T, cfg *app.Config) *Component {
	require := require.New(t)

	port := getFreePort(t)
	cfg.PostgresDB.Port = port

	t.Logf("use port %d for PostgreSQL container", port)

	c := &Component{
		ImageName: "postgres:12-alpine",
		Ports: [][2]int{
			{port, 5432},
		},
		Env: []string{
			"POSTGRES_HOST_AUTH_METHOD=trust",
		},
	}

	err := c.Run()
	require.NoError(err)

	return c
}

// Component is a dependency (for example, db) that will be run with Docker
type Component struct {
	ImageName string
	Ports     [][2]int
	Env       []string

	containerID string
}

// Run runs a component as a docker container:
//
//	docker run --rm -d [-e ...] [-p ...] image
//
func (c *Component) Run() error {
	if err := checkDockerImage(c.ImageName); err != nil {
		return err
	}

	args := []string{
		"run", "--rm", "-d",
	}
	for _, env := range c.Env {
		args = append(args, "-e", env)
	}
	for _, p := range c.Ports {
		args = append(args, "-p", fmt.Sprintf("%d:%d", p[0], p[1]))
	}
	args = append(args, c.ImageName)

	cmd := exec.Command("docker", args...)
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "couldn't run component %q", c.ImageName)
	}

	c.containerID = output.String()
	c.containerID = strings.TrimSpace(c.containerID)

	if c.containerID == "" {
		return errors.Errorf("couldn't get container id for component %q", c.ImageName)
	}
	return nil
}

// Stop stops a component
func (c *Component) Stop() error {
	if c.containerID == "" {
		return errors.Errorf("component %q is not run", c.ImageName)
	}
	return exec.Command("docker", "stop", c.containerID).Run() //nolint:gosec
}

func checkDockerImage(image string) error {
	cmd := exec.Command("docker", "images", "-q", image)
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "couldn't check image %q", image)
	}
	if strings.TrimSpace(output.String()) == "" {
		return errors.Errorf("no image %q", image)
	}
	return nil
}
