package check

import (
	"context"
	"os"
	"strings"
	"testing"
)

func writeGitignore(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(dir+"/.gitignore", []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .gitignore: %v", err)
	}
}

func universalPatterns() []requiredPattern {
	return []requiredPattern{
		{pattern: ".env", description: "Environment files with secrets"},
		{pattern: "*.log", description: "Log files"},
	}
}

func allPatterns() []requiredPattern {
	return []requiredPattern{
		{pattern: ".env", description: "Environment files with secrets"},
		{pattern: "*.log", description: "Log files"},
		{pattern: "node_modules", description: "Node.js dependencies"},
		{pattern: "__pycache__", description: "Python cache"},
	}
}

// Pass cases

func TestGitignoreCheck_AllPatternsPresent(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, ".env\n*.log\nnode_modules\n__pycache__\n")

	c := &GitignoreCheck{Dir: dir, Patterns: allPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}

func TestGitignoreCheck_TrailingSlashVariant(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, ".env\n*.log\nnode_modules/\n__pycache__/\n")

	c := &GitignoreCheck{Dir: dir, Patterns: allPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusPass {
		t.Errorf("expected pass with trailing slash variant, got %v: %s", result.Status, result.Message)
	}
}

func TestGitignoreCheck_DoubleStarVariant(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, ".env\n*.log\n**/node_modules\n**/__pycache__\n")

	c := &GitignoreCheck{Dir: dir, Patterns: allPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusPass {
		t.Errorf("expected pass with **/ variant, got %v: %s", result.Status, result.Message)
	}
}

func TestGitignoreCheck_DoubleStarWithSlashVariant(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, ".env\n*.log\n**/node_modules/\n**/__pycache__/\n")

	c := &GitignoreCheck{Dir: dir, Patterns: allPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusPass {
		t.Errorf("expected pass with **/ trailing slash variant, got %v: %s", result.Status, result.Message)
	}
}

func TestGitignoreCheck_CommentsAndBlankLinesIgnored(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, "# this is a comment\n\n.env\n*.log\n")

	c := &GitignoreCheck{Dir: dir, Patterns: universalPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}

// Warn cases

func TestGitignoreCheck_MissingEnv(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, "*.log\n")

	c := &GitignoreCheck{Dir: dir, Patterns: universalPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusWarn {
		t.Errorf("expected warn, got %v", result.Status)
	}
	if !strings.Contains(result.Fix, ".env") {
		t.Errorf("expected Fix to mention .env, got: %s", result.Fix)
	}
}

func TestGitignoreCheck_MissingMultiplePatterns(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, "# empty gitignore\n")

	c := &GitignoreCheck{Dir: dir, Patterns: allPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusWarn {
		t.Errorf("expected warn, got %v", result.Status)
	}
	if !strings.Contains(result.Message, "4") {
		t.Errorf("expected message to report 4 missing patterns, got: %s", result.Message)
	}
}

func TestGitignoreCheck_FixMessageListsMissingPatterns(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, "*.log\n")

	c := &GitignoreCheck{Dir: dir, Patterns: universalPatterns()}
	result := c.Run(context.Background())

	if !strings.Contains(result.Fix, ".env") {
		t.Errorf("expected Fix to contain .env, got: %s", result.Fix)
	}
	if strings.Contains(result.Fix, "*.log") {
		t.Errorf("expected Fix not to mention *.log (already present), got: %s", result.Fix)
	}
}

func TestGitignoreCheck_FixMessageOrderIsConsistent(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, "# nothing\n")

	c := &GitignoreCheck{Dir: dir, Patterns: universalPatterns()}
	result1 := c.Run(context.Background())
	result2 := c.Run(context.Background())

	if result1.Fix != result2.Fix {
		t.Errorf("Fix message order is non-deterministic:\nrun1: %s\nrun2: %s", result1.Fix, result2.Fix)
	}
}

// Stack filtering

func TestGitignoreCheck_NodeModulesSkippedWhenNotInPatterns(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, ".env\n*.log\n") // no node_modules

	// Only universal patterns passed — simulates a Go project
	c := &GitignoreCheck{Dir: dir, Patterns: universalPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusPass {
		t.Errorf("expected pass for Go project without node_modules, got %v: %s", result.Status, result.Message)
	}
}

func TestGitignoreCheck_NodeModulesWarnedWhenInPatterns(t *testing.T) {
	dir := t.TempDir()
	writeGitignore(t, dir, ".env\n*.log\n") // missing node_modules

	c := &GitignoreCheck{Dir: dir, Patterns: allPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusWarn {
		t.Errorf("expected warn for Node project missing node_modules, got %v", result.Status)
	}
	if !strings.Contains(result.Fix, "node_modules") {
		t.Errorf("expected Fix to mention node_modules, got: %s", result.Fix)
	}
}

// Fail case

func TestGitignoreCheck_NoGitignoreFile(t *testing.T) {
	dir := t.TempDir() // no .gitignore created

	c := &GitignoreCheck{Dir: dir, Patterns: universalPatterns()}
	result := c.Run(context.Background())

	if result.Status != StatusFail {
		t.Errorf("expected fail when .gitignore missing, got %v", result.Status)
	}
}
func TestPluralPatterns(t *testing.T) {
	if got := pluralPatterns(1); got != "pattern" {
		t.Errorf("pluralPatterns(1) = %q, want %q", got, "pattern")
	}
	if got := pluralPatterns(2); got != "patterns" {
		t.Errorf("pluralPatterns(2) = %q, want %q", got, "patterns")
	}
}
