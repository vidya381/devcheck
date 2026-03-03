package check

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGitHooksCheck_Node_PassAndWarn(t *testing.T) {
	dir := t.TempDir()

	// ensure we are treated as a git repo
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	check := &GitHooksCheck{Dir: dir, Stack: "node"}

	// Pass when .husky exists
	if err := os.Mkdir(filepath.Join(dir, ".husky"), 0o755); err != nil {
		t.Fatalf("failed to create .husky directory: %v", err)
	}
	result := check.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass when .husky exists, got %v: %s", result.Status, result.Message)
	}

	// Warn when .husky is missing
	if err := os.RemoveAll(filepath.Join(dir, ".husky")); err != nil {
		t.Fatalf("failed to remove .husky directory: %v", err)
	}
	result = check.Run(context.Background())
	if result.Status != StatusWarn {
		t.Errorf("expected warn when .husky missing, got %v: %s", result.Status, result.Message)
	}
}

func TestGitHooksCheck_Python_PassAndWarn(t *testing.T) {
	dir := t.TempDir()

	// ensure we are treated as a git repo
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// override lookPath so tests don't depend on environment
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	check := &GitHooksCheck{Dir: dir, Stack: "python"}

	// Pass when config exists and pre-commit is "installed"
	if err := os.WriteFile(filepath.Join(dir, ".pre-commit-config.yaml"), []byte("repos: []\n"), 0o644); err != nil {
		t.Fatalf("failed to write .pre-commit-config.yaml: %v", err)
	}
	lookPath = func(string) (string, error) {
		return "/usr/bin/pre-commit", nil
	}

	result := check.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass when pre-commit and config present, got %v: %s", result.Status, result.Message)
	}

	// Warn when either piece is missing (simulate missing pre-commit)
	lookPath = func(string) (string, error) {
		return "", os.ErrNotExist
	}

	result = check.Run(context.Background())
	if result.Status != StatusWarn {
		t.Errorf("expected warn when pre-commit missing, got %v: %s", result.Status, result.Message)
	}
}

