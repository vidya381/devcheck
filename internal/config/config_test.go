package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoFile_ReturnsEmptyConfig(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("expected no error when config file absent, got: %v", err)
	}
	if len(cfg.Require) != 0 || len(cfg.Skip) != 0 {
		t.Errorf("expected zero-value config, got: %+v", cfg)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	content := `
require:
  - terraform
  - aws
skip:
  - golangci-lint
`
	writeFile(t, filepath.Join(dir, Filename), content)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Require) != 2 || cfg.Require[0] != "terraform" || cfg.Require[1] != "aws" {
		t.Errorf("unexpected Require: %v", cfg.Require)
	}
	if len(cfg.Skip) != 1 || cfg.Skip[0] != "golangci-lint" {
		t.Errorf("unexpected Skip: %v", cfg.Skip)
	}
}

func TestLoad_RequireOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, Filename), "require:\n  - kubectl\n")
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Require) != 1 || cfg.Require[0] != "kubectl" {
		t.Errorf("unexpected Require: %v", cfg.Require)
	}
	if len(cfg.Skip) != 0 {
		t.Errorf("expected empty Skip, got: %v", cfg.Skip)
	}
}

func TestLoad_SkipOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, Filename), "skip:\n  - Node version\n")
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Skip) != 1 || cfg.Skip[0] != "Node version" {
		t.Errorf("unexpected Skip: %v", cfg.Skip)
	}
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, Filename), "require: [unclosed")
	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_EmptyFile_ReturnsEmptyConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, Filename), "")
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error for empty file: %v", err)
	}
	if len(cfg.Require) != 0 || len(cfg.Skip) != 0 {
		t.Errorf("expected zero-value config for empty file, got: %+v", cfg)
	}
}

func TestSkipSet(t *testing.T) {
	cfg := Config{Skip: []string{"golangci-lint", "Node version"}}
	s := cfg.SkipSet()
	if _, ok := s["golangci-lint"]; !ok {
		t.Error("expected 'golangci-lint' in skip set")
	}
	if _, ok := s["Node version"]; !ok {
		t.Error("expected 'Node version' in skip set")
	}
	if _, ok := s["docker installed"]; ok {
		t.Error("did not expect 'docker installed' in skip set")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
