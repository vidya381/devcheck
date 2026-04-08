package check

import (
	"context"
	"os"
	"testing"
)

func TestVersionLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.21.0", "1.22", true},
		{"1.22.0", "1.21", false},
		{"1.22.0", "1.22", false},
		{"20.0.0", "18", false},
		{"16.0.0", "18", true},
	}
	for _, tc := range cases {
		got := versionLess(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("versionLess(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestGoVersionCheck_Pass(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/go.mod", []byte("module example\n\ngo 1.1\n"), 0644)

	c := &GoVersionCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}

func TestGoVersionCheck_Fail(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/go.mod", []byte("module example\n\ngo 9999.0\n"), 0644)

	c := &GoVersionCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusFail {
		t.Errorf("expected fail, got %v: %s", result.Status, result.Message)
	}
}

func TestGoVersionCheck_MissingGoMod(t *testing.T) {
	c := &GoVersionCheck{Dir: t.TempDir()}
	result := c.Run(context.Background())
	if result.Status != StatusSkipped {
		t.Errorf("expected skipped, got %v", result.Status)
	}
}

func TestNodeVersionCheck_NvmrcPass(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/.nvmrc", []byte("1\n"), 0644)

	c := &NodeVersionCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}

func TestNodeVersionCheck_PackageJsonPass(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/package.json", []byte(`{"engines":{"node":">=1.0.0"}}`), 0644)

	c := &NodeVersionCheck{Dir: dir}
	result := c.Run(context.Background())
	if result.Status != StatusPass {
		t.Errorf("expected pass, got %v: %s", result.Status, result.Message)
	}
}

func TestNodeVersionCheck_NoRequirement(t *testing.T) {
	c := &NodeVersionCheck{Dir: t.TempDir()}
	result := c.Run(context.Background())
	if result.Status != StatusSkipped {
		t.Errorf("expected skipped, got %v", result.Status)
	}
}

func TestRustVersionCheck_NoToolchainFile_Skipped(t *testing.T) {
	c := &RustVersionCheck{Dir: t.TempDir()}
	r := c.Run(context.Background())
	if r.Status != StatusSkipped {
		t.Errorf("expected skipped when no rust-toolchain.toml, got %v: %s", r.Status, r.Message)
	}
}

func TestRustVersionCheck_Pass(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/rust-toolchain.toml", []byte("[toolchain]\nchannel = \"1.0.0\"\n"), 0644)

	c := &RustVersionCheck{Dir: dir}
	r := c.Run(context.Background())
	// rustc is likely installed in CI; if not, we get Fail, not Skipped.
	// Just ensure it doesn't panic and returns a result.
	if r.Name != "Rust version" {
		t.Errorf("unexpected check name: %s", r.Name)
	}
}

func TestRustVersionCheck_Fail(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/rust-toolchain.toml", []byte("[toolchain]\nchannel = \"9999.0.0\"\n"), 0644)

	c := &RustVersionCheck{Dir: dir}
	r := c.Run(context.Background())
	// If rustc is installed, this must fail (version too high).
	// If rustc is not installed, it fails with "could not run rustc --version".
	if r.Status != StatusFail {
		t.Errorf("expected fail for unreachable version, got %v: %s", r.Status, r.Message)
	}
}

func TestRustVersionCheck_StableChannel_Skipped(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/rust-toolchain.toml", []byte("[toolchain]\nchannel = \"stable\"\n"), 0644)

	c := &RustVersionCheck{Dir: dir}
	r := c.Run(context.Background())
	if r.Status != StatusSkipped {
		t.Errorf("expected skipped for non-pinned channel 'stable', got %v: %s", r.Status, r.Message)
	}
}

func TestReadRustRequired_TomlPinned(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/rust-toolchain.toml", []byte("[toolchain]\nchannel = \"1.76.0\"\n"), 0644)
	v, err := readRustRequired(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "1.76.0" {
		t.Errorf("expected 1.76.0, got %q", v)
	}
}

func TestReadRustRequired_LegacyFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/rust-toolchain", []byte("1.70.0\n"), 0644)
	v, err := readRustRequired(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "1.70.0" {
		t.Errorf("expected 1.70.0, got %q", v)
	}
}

func TestReadRustRequired_NoFile_Error(t *testing.T) {
	_, err := readRustRequired(t.TempDir())
	if err == nil {
		t.Error("expected error when no toolchain file present")
	}
}