package initdir

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRun_alphabetical_order(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dir := t.TempDir()
	output := filepath.Join(dir, "output.txt")

	writeScript(t, dir, "02_second.sh", `echo "second" >> `+output)
	writeScript(t, dir, "01_first.sh", `echo "first" >> `+output)
	writeScript(t, dir, "03_third.sh", `echo "third" >> `+output)

	if err := Run(context.Background(), dir, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(output) //nolint:gosec // test file path from t.TempDir
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	want := "first\nsecond\nthird\n"
	if string(data) != want {
		t.Errorf("got %q, want %q", string(data), want)
	}
}

func TestRun_skips_non_sh_files(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dir := t.TempDir()
	output := filepath.Join(dir, "output.txt")

	writeScript(t, dir, "run.sh", `echo "executed" >> `+output)
	writeFile(t, dir, "skip.txt", "should not run")
	writeFile(t, dir, "skip.py", "should not run")

	if err := Run(context.Background(), dir, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(output) //nolint:gosec // test file path from t.TempDir
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	want := "executed\n"
	if string(data) != want {
		t.Errorf("got %q, want %q", string(data), want)
	}
}

func TestRun_continues_after_failure(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dir := t.TempDir()
	output := filepath.Join(dir, "output.txt")

	writeScript(t, dir, "01_ok.sh", `echo "first" >> `+output)
	writeScript(t, dir, "02_fail.sh", `exit 1`)
	writeScript(t, dir, "03_ok.sh", `echo "third" >> `+output)

	if err := Run(context.Background(), dir, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(output) //nolint:gosec // test file path from t.TempDir
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	want := "first\nthird\n"
	if string(data) != want {
		t.Errorf("got %q, want %q", string(data), want)
	}
}

func TestRun_empty_directory(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dir := t.TempDir()

	if err := Run(context.Background(), dir, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRun_nonexistent_directory(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := Run(context.Background(), "/nonexistent/path", logger)
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestRun_respects_context_cancellation(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	dir := t.TempDir()

	writeScript(t, dir, "01_slow.sh", `sleep 10`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(ctx, dir, logger)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func writeScript(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+content+"\n"), 0o700); err != nil { //nolint:gosec // test script needs execute permission
		t.Fatalf("failed to write script %s: %v", name, err)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write file %s: %v", name, err)
	}
}
