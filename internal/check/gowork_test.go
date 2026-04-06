package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeGoWork(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte(content), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}
}

func mkGoMod(t *testing.T, dir, rel string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
	if err := os.WriteFile(filepath.Join(p, "go.mod"), []byte("module example.com/mod\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func TestGoWorkCheck_Pass_BlockForm(t *testing.T) {
	dir := t.TempDir()
	writeGoWork(t, dir, "go 1.22\n\nuse (\n\t./svc/api\n\t./svc/worker\n)\n")
	mkGoMod(t, dir, "svc/api")
	mkGoMod(t, dir, "svc/worker")

	c := &GoWorkCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", r.Status, r.Message)
	}
}

func TestGoWorkCheck_Pass_SingleLineForm(t *testing.T) {
	dir := t.TempDir()
	writeGoWork(t, dir, "go 1.22\n\nuse ./svc/api\n")
	mkGoMod(t, dir, "svc/api")

	c := &GoWorkCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", r.Status, r.Message)
	}
}

func TestGoWorkCheck_Fail_MissingModule(t *testing.T) {
	dir := t.TempDir()
	writeGoWork(t, dir, "go 1.22\n\nuse (\n\t./svc/api\n\t./svc/missing\n)\n")
	mkGoMod(t, dir, "svc/api")
	// svc/missing intentionally absent

	c := &GoWorkCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusFail {
		t.Fatalf("expected fail, got %v: %s", r.Status, r.Message)
	}
	if !strings.Contains(r.Message, "svc/missing") {
		t.Errorf("expected missing path in message, got: %s", r.Message)
	}
}

func TestGoWorkCheck_Fail_DirExistsButNoGoMod(t *testing.T) {
	dir := t.TempDir()
	writeGoWork(t, dir, "go 1.22\n\nuse ./svc/api\n")
	// create the directory but no go.mod inside
	if err := os.MkdirAll(filepath.Join(dir, "svc", "api"), 0o755); err != nil {
		t.Fatal(err)
	}

	c := &GoWorkCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusFail {
		t.Errorf("expected fail when directory exists but go.mod missing, got %v", r.Status)
	}
}

func TestGoWorkCheck_Pass_NoUseDirectives(t *testing.T) {
	dir := t.TempDir()
	writeGoWork(t, dir, "go 1.22\n")

	c := &GoWorkCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected pass for go.work with no use directives, got %v: %s", r.Status, r.Message)
	}
}

func TestGoWorkCheck_Fail_MissingGoWorkFile(t *testing.T) {
	dir := t.TempDir()
	// no go.work written

	c := &GoWorkCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusFail {
		t.Errorf("expected fail when go.work is missing, got %v", r.Status)
	}
}

func TestGoWorkCheck_IgnoresComments(t *testing.T) {
	dir := t.TempDir()
	writeGoWork(t, dir, "go 1.22\n\nuse (\n\t./svc/api // main service\n\t// ./svc/disabled\n)\n")
	mkGoMod(t, dir, "svc/api")
	// svc/disabled should be ignored

	c := &GoWorkCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", r.Status, r.Message)
	}
}

func TestParseGoWorkUse_BlockAndSingle(t *testing.T) {
	dir := t.TempDir()
	writeGoWork(t, dir, "go 1.22\n\nuse (\n\t./a\n\t./b\n)\n\nuse ./c\n")

	paths, err := parseGoWorkUse(filepath.Join(dir, "go.work"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"./a", "./b", "./c"}
	if len(paths) != len(want) {
		t.Fatalf("expected %v, got %v", want, paths)
	}
	for i, p := range paths {
		if p != want[i] {
			t.Errorf("paths[%d]: expected %q, got %q", i, want[i], p)
		}
	}
}