package tests

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ShoshinNikita/budget-manager/internal/app"
	"github.com/ShoshinNikita/budget-manager/internal/pkg/errors"
)

type Component interface {
	GetName() string
	Cleanup() error
}

type StartComponentFn func(*testing.T, *app.Config) Component

// StartPostgreSQL starts a fresh PostgreSQL instance in a docker container.
// It updates PostgreSQL config with a chosen port
func StartPostgreSQL(t *testing.T, cfg *app.Config) Component {
	require := require.New(t)

	port := getFreePort(t)
	cfg.PostgresDB.Port = port

	t.Logf("use port %d for PostgreSQL container", port)

	c := &DockerComponent{
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

// StartSQLite generates a random path and updates SQLite config
func StartSQLite(t *testing.T, cfg *app.Config) Component {
	require := require.New(t)

	dbPath := func() string {
		dir := os.TempDir()

		b := make([]byte, 4)
		_, err := rand.Read(b)
		require.NoError(err)

		filename := "budget-manager-" + hex.EncodeToString(b) + ".db"

		return filepath.Join(dir, filename)
	}()

	cfg.SQLiteDB.Path = dbPath
	t.Logf("use path %s for SQLite", dbPath)

	return &CustomComponent{
		Name: "SQLite (" + dbPath + ")",
		CleanupFn: func() error {
			return os.Remove(dbPath)
		},
	}
}

type CustomComponent struct {
	Name      string
	CleanupFn func() error
}

func (c *CustomComponent) GetName() string {
	return c.Name
}

func (c *CustomComponent) Cleanup() error {
	return c.CleanupFn()
}

// DockerComponent is a dependency (for example, db) that will be run with Docker
type DockerComponent struct {
	ImageName string
	Ports     [][2]int
	Env       []string

	containerID string
}

// Run runs a component as a docker container:
//
//	docker run --rm -d [-e ...] [-p ...] image
//
func (c *DockerComponent) Run() error {
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

func (c *DockerComponent) GetName() string {
	name := c.ImageName
	if c.containerID != "" {
		name += " (" + c.containerID + ")"
	}
	return name
}

// Cleanup stops a docker container
func (c *DockerComponent) Cleanup() error {
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
