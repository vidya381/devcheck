package check

import (
	"context"
	"os"
	"testing"
)

func TestEnvCheck_Pass(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/.env.example", []byte("DB_URL=\nAPI_KEY=\n"), 0644)
	os.WriteFile(dir+"/.env", []byte("DB_URL=postgres://localhost\nAPI_KEY=secret\n"), 0644)

	c := &EnvCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}

func TestEnvCheck_MissingKeys(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/.env.example", []byte("DB_URL=\nAPI_KEY=\nSECRET=\n"), 0644)
	os.WriteFile(dir+"/.env", []byte("DB_URL=postgres://localhost\n"), 0644)

	c := &EnvCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusFail {
		t.Errorf("expected fail, got %v", result.Status)
	}
}

func TestEnvCheck_MissingEnvFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/.env.example", []byte("DB_URL=\n"), 0644)

	c := &EnvCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusFail {
		t.Errorf("expected fail, got %v", result.Status)
	}
}

func TestEnvCheck_IgnoresComments(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/.env.example", []byte("# this is a comment\nDB_URL=\n\nAPI_KEY=\n"), 0644)
	os.WriteFile(dir+"/.env", []byte("DB_URL=postgres://localhost\nAPI_KEY=secret\n"), 0644)

	c := &EnvCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}
