package check

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"strings"
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

func TestDepsCheck_Node_FixMessage_DefaultsToNpm(t *testing.T) {
	dir := t.TempDir()
	check := &DepsCheck{Dir: dir, Stack: "node"} // no PackageManager set
	result := check.Run(context.Background())
	if result.Status != StatusFail {
		t.Fatalf("expected fail, got %v", result.Status)
	}
	if !strings.Contains(result.Fix, "npm install") {
		t.Errorf("expected fix to mention 'npm install', got: %s", result.Fix)
	}
}

func TestDepsCheck_Node_FixMessage_Pnpm(t *testing.T) {
	dir := t.TempDir()
	check := &DepsCheck{Dir: dir, Stack: "node", PackageManager: "pnpm"}
	result := check.Run(context.Background())
	if result.Status != StatusFail {
		t.Fatalf("expected fail, got %v", result.Status)
	}
	if !strings.Contains(result.Fix, "pnpm install") {
		t.Errorf("expected fix to mention 'pnpm install', got: %s", result.Fix)
	}
}

func TestDepsCheck_Node_FixMessage_Yarn(t *testing.T) {
	dir := t.TempDir()
	check := &DepsCheck{Dir: dir, Stack: "node", PackageManager: "yarn"}
	result := check.Run(context.Background())
	if result.Status != StatusFail {
		t.Fatalf("expected fail, got %v", result.Status)
	}
	if !strings.Contains(result.Fix, "yarn install") {
		t.Errorf("expected fix to mention 'yarn install', got: %s", result.Fix)
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

func TestDepsCheck_Python_VenvNoRequirements_StillPass(t *testing.T) {
	// Existing behaviour must be preserved: venv present, no requirements.txt → pass.
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".venv"), 0o755); err != nil {
		t.Fatalf("mkdir .venv: %v", err)
	}
	c := &DepsCheck{Dir: dir, Stack: "python"}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected Pass when no requirements.txt, got %v: %s", r.Status, r.Message)
	}
}

func TestDepsCheck_Python_AllPackagesPresent(t *testing.T) {
	dir, pipBin := setupPythonDir(t, "requests==2.31.0\nflask>=2.0\n# comment\n")
	c := &DepsCheck{
		Dir:   dir,
		Stack: "python",
		pipCheck: func(_ string) error { return nil },
		pipFreeze: func(_ string) ([]byte, error) {
			return []byte("requests==2.31.0\nFlask==2.3.0\n"), nil
		},
	}
	_ = pipBin
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected Pass when all packages present, got %v: %s", r.Status, r.Message)
	}
}

func TestDepsCheck_Python_MissingPackage(t *testing.T) {
	dir, _ := setupPythonDir(t, "requests==2.31.0\ncelery>=5.0\n")
	c := &DepsCheck{
		Dir:   dir,
		Stack: "python",
		pipCheck: func(_ string) error { return nil },
		pipFreeze: func(_ string) ([]byte, error) {
			return []byte("requests==2.31.0\n"), nil // celery absent
		},
	}
	r := c.Run(context.Background())
	if r.Status != StatusFail {
		t.Fatalf("expected Fail for missing package, got %v: %s", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "celery") {
		t.Errorf("expected 'celery' in message, got: %s", r.Message)
	}
}

func TestDepsCheck_Python_PipCheckConflict(t *testing.T) {
	dir, _ := setupPythonDir(t, "requests\n")
	c := &DepsCheck{
		Dir:      dir,
		Stack:    "python",
		pipCheck: func(_ string) error { return errors.New("conflict") },
	}
	r := c.Run(context.Background())
	if r.Status != StatusFail {
		t.Errorf("expected Fail when pip check reports conflict, got %v: %s", r.Status, r.Message)
	}
}

func TestDepsCheck_Python_CaseInsensitiveMatch(t *testing.T) {
	// requirements.txt uses "Requests"; freeze returns "requests" — should still pass.
	dir, _ := setupPythonDir(t, "Requests>=2.0\n")
	c := &DepsCheck{
		Dir:   dir,
		Stack: "python",
		pipCheck: func(_ string) error { return nil },
		pipFreeze: func(_ string) ([]byte, error) {
			return []byte("requests==2.31.0\n"), nil
		},
	}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected Pass for case-insensitive match, got %v: %s", r.Status, r.Message)
	}
}

func TestDepsCheck_Python_EditableInstall(t *testing.T) {
	dir, _ := setupPythonDir(t, "mylib\n")
	c := &DepsCheck{
		Dir:   dir,
		Stack: "python",
		pipCheck: func(_ string) error { return nil },
		pipFreeze: func(_ string) ([]byte, error) {
			return []byte("-e git+https://github.com/org/mylib.git@main#egg=mylib\n"), nil
		},
	}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected Pass for editable install, got %v: %s", r.Status, r.Message)
	}
}

// setupPythonDir creates a temp dir with a .venv/bin/pip stub and a
// requirements.txt containing the given content, then returns both.
func setupPythonDir(t *testing.T, requirements string) (dir string, pipBin string) {
	t.Helper()
	dir = t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".venv", "bin"), 0o755); err != nil {
		t.Fatalf("mkdir .venv/bin: %v", err)
	}
	pipBin = filepath.Join(dir, ".venv", "bin", "pip")
	if err := os.WriteFile(pipBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write pip stub: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(requirements), 0o644); err != nil {
		t.Fatalf("write requirements.txt: %v", err)
	}
	return dir, pipBin
}

func TestDepsCheck_Ruby_Pass_VendorBundle(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "vendor", "bundle"), 0o755); err != nil {
		t.Fatalf("mkdir vendor/bundle: %v", err)
	}
	c := &DepsCheck{Dir: dir, Stack: "ruby"}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected pass when vendor/bundle exists, got %v: %s", r.Status, r.Message)
	}
}

func TestDepsCheck_Ruby_Pass_GemfileLock(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Gemfile.lock"), []byte("GEM\n"), 0o644); err != nil {
		t.Fatalf("write Gemfile.lock: %v", err)
	}
	c := &DepsCheck{Dir: dir, Stack: "ruby"}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected pass when Gemfile.lock exists, got %v: %s", r.Status, r.Message)
	}
}

func TestDepsCheck_Ruby_Fail_NeitherPresent(t *testing.T) {
	dir := t.TempDir()
	c := &DepsCheck{Dir: dir, Stack: "ruby"}
	r := c.Run(context.Background())
	if r.Status != StatusFail {
		t.Errorf("expected fail when neither vendor/bundle nor Gemfile.lock exist, got %v", r.Status)
	}
	if !strings.Contains(r.Fix, "bundle install") {
		t.Errorf("expected fix to mention 'bundle install', got: %s", r.Fix)
	}
}

func TestDepsCheck_Ruby_VendorBundle_Takes_Priority(t *testing.T) {
	// Both present — vendor/bundle should win (pass with that message).
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "vendor", "bundle"), 0o755); err != nil {
		t.Fatalf("mkdir vendor/bundle: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Gemfile.lock"), []byte("GEM\n"), 0o644); err != nil {
		t.Fatalf("write Gemfile.lock: %v", err)
	}
	c := &DepsCheck{Dir: dir, Stack: "ruby"}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "vendor/bundle") {
		t.Errorf("expected vendor/bundle message to take priority, got: %s", r.Message)
	}
}