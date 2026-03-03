package check

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDepsCheck_Node_PassAndFail(t *testing.T) {
	dir := t.TempDir()
	check := &DepsCheck{Dir: dir, Stack: "node"}

	// Pass when node_modules exists
	if err := os.Mkdir(filepath.Join(dir, "node_modules"), 0o755); err != nil {
		t.Fatalf("failed to create node_modules: %v", err)
	}
	result := check.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass when node_modules exists, got %v: %s", result.Status, result.Message)
	}

	// Fail when node_modules is missing
	if err := os.RemoveAll(filepath.Join(dir, "node_modules")); err != nil {
		t.Fatalf("failed to remove node_modules: %v", err)
	}
	result = check.Run(context.Background())
	if result.Status != StatusFail {
		t.Errorf("expected fail when node_modules missing, got %v: %s", result.Status, result.Message)
	}
}

func TestDepsCheck_Python_PassAndFail(t *testing.T) {
	dir := t.TempDir()
	check := &DepsCheck{Dir: dir, Stack: "python"}

	// Pass when .venv exists
	if err := os.Mkdir(filepath.Join(dir, ".venv"), 0o755); err != nil {
		t.Fatalf("failed to create .venv: %v", err)
	}
	result := check.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass when .venv exists, got %v: %s", result.Status, result.Message)
	}

	// Fail when no venv directories exist
	if err := os.RemoveAll(filepath.Join(dir, ".venv")); err != nil {
		t.Fatalf("failed to remove .venv: %v", err)
	}
	result = check.Run(context.Background())
	if result.Status != StatusFail {
		t.Errorf("expected fail when no venv directories, got %v: %s", result.Status, result.Message)
	}
}

func TestDepsCheck_Go_PassAndFail(t *testing.T) {
	dir := t.TempDir()

	// Pass when vendor directory exists
	check := &DepsCheck{Dir: dir, Stack: "go"}
	if err := os.Mkdir(filepath.Join(dir, "vendor"), 0o755); err != nil {
		t.Fatalf("failed to create vendor: %v", err)
	}
	result := check.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass when vendor exists, got %v: %s", result.Status, result.Message)
	}

	// When vendor is missing, pass if goCheck succeeds
	if err := os.RemoveAll(filepath.Join(dir, "vendor")); err != nil {
		t.Fatalf("failed to remove vendor: %v", err)
	}
	check = &DepsCheck{
		Dir:   dir,
		Stack: "go",
		goCheck: func(string) error {
			return nil
		},
	}
	result = check.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass when goCheck succeeds, got %v: %s", result.Status, result.Message)
	}

	// Fail when goCheck reports an error
	check = &DepsCheck{
		Dir:   dir,
		Stack: "go",
		goCheck: func(string) error {
			return errors.New("download failed")
		},
	}
	result = check.Run(context.Background())
	if result.Status != StatusFail {
		t.Errorf("expected fail when goCheck fails, got %v: %s", result.Status, result.Message)
	}
}

