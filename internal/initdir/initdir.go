// Package initdir executes shell scripts from a directory after kumo starts.
// This provides functionality similar to LocalStack's init/ready.d/ mechanism.
package initdir

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Run executes all .sh files in the given directory sequentially in alphabetical order.
// Each script runs independently — a failure in one script does not prevent subsequent scripts from running.
func Run(ctx context.Context, dir string, logger *slog.Logger) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read init directory %s: %w", dir, err)
	}

	var scripts []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".sh") {
			scripts = append(scripts, entry.Name())
		}
	}

	sort.Strings(scripts)

	if len(scripts) == 0 {
		logger.Info("no init scripts found", "dir", dir)

		return nil
	}

	logger.Info("executing init scripts", "dir", dir, "count", len(scripts))

	var failed int

	for _, name := range scripts {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("init script execution cancelled: %w", err)
		}

		path := filepath.Join(dir, name)
		logger.Info("executing init script", "script", name)

		if err := runScript(ctx, path); err != nil {
			logger.Error("init script failed", "script", name, "error", err)

			failed++

			continue
		}

		logger.Info("init script completed", "script", name)
	}

	if failed > 0 {
		logger.Warn("some init scripts failed", "failed", failed, "total", len(scripts))
	}

	return nil
}

func runScript(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "sh", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Inherit environment so scripts can use AWS_ENDPOINT_URL, AWS_DEFAULT_REGION, etc.
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script %s exited with error: %w", filepath.Base(path), err)
	}

	return nil
}
