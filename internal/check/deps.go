package check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// goModDownload is a variable so tests can stub it.
var goModDownload = func(dir string) error {
	cmd := exec.Command("go", "mod", "download", "-json", "all")
	cmd.Dir = dir
	return cmd.Run()
}

// pipFreezeRunner is a variable so tests can stub it.
var pipFreezeRunner = func(pipBin string) ([]byte, error) {
	return exec.Command(pipBin, "freeze").Output()
}

// pipCheckRunner is a variable so tests can stub it.
var pipCheckRunner = func(pipBin string) error {
	return exec.Command(pipBin, "check").Run()
}

type DepsCheck struct {
	Dir            string
	Stack          string // "node", "python", or "go"
	PackageManager string // "npm", "pnpm", or "yarn" (Node only; defaults to "npm")
	goCheck   func(dir string) error
	pipFreeze func(pipBin string) ([]byte, error)
	pipCheck  func(pipBin string) error
}

func (c *DepsCheck) Name() string {
	switch c.Stack {
	case "node":
		return "Node dependencies installed"
	case "python":
		return "Python dependencies installed"
	case "go":
		return "Go dependencies installed"
	case "ruby":
		return "Ruby dependencies installed"
	default:
		return "Project dependencies installed"
	}
}

func (c *DepsCheck) Run(_ context.Context) Result {
	switch c.Stack {
	case "node":
		return c.runNode()
	case "python":
		return c.runPython()
	case "go":
		return c.runGo()
	case "ruby":
		return c.runRuby()
	default:
		return Result{
			Name:    c.Name(),
			Status:  StatusSkipped,
			Message: "unknown stack type for dependency check",
		}
	}
}

func (c *DepsCheck) runNode() Result {
	nodeModules := filepath.Join(c.Dir, "node_modules")
	if dirExists(nodeModules) {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "node_modules directory exists",
		}
	}
	pm := c.PackageManager
	if pm == "" {
		pm = "npm"
	}
	return Result{
		Name:    c.Name(),
		Status:  StatusFail,
		Message: "node_modules directory not found",
		Fix:     fmt.Sprintf("run `%s install` to install Node dependencies", pm),
	}
}

func (c *DepsCheck) runPython() Result {
	venvDir := c.findVenv()
	if venvDir == "" {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Python virtual environment directory not found",
			Fix:     "create a virtual environment (e.g. `python -m venv .venv`) and install dependencies with `pip install -r requirements.txt` or equivalent",
		}
	}

	// No requirements.txt — venv existing is sufficient.
	reqFile := filepath.Join(c.Dir, "requirements.txt")
	if _, err := os.Stat(reqFile); os.IsNotExist(err) {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Python virtual environment directory exists",
		}
	}

	pipBin := c.findPipBin(venvDir)

	// Run pip check for dependency conflicts first.
	checkFn := c.pipCheck
	if checkFn == nil {
		checkFn = pipCheckRunner
	}
	if err := checkFn(pipBin); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("pip check reported dependency conflicts: %v", err),
			Fix:     "run `pip install -r requirements.txt` inside your virtual environment to resolve conflicting or missing packages",
		}
	}

	// Compare pip freeze against requirements.txt to catch missing packages.
	freezeFn := c.pipFreeze
	if freezeFn == nil {
		freezeFn = pipFreezeRunner
	}
	missing, err := findMissingRequirements(pipBin, reqFile, freezeFn)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("could not compare installed packages to requirements.txt: %v", err),
			Fix:     "ensure pip is available in your virtual environment and requirements.txt is readable",
		}
	}
	if len(missing) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("packages listed in requirements.txt but not installed: %s", strings.Join(missing, ", ")),
			Fix:     "run `pip install -r requirements.txt` inside your virtual environment",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Python virtual environment exists and packages match requirements.txt",
	}
}

// findVenv returns the path of the first venv directory found, or "".
func (c *DepsCheck) findVenv() string {
	for _, name := range []string{"venv", ".venv"} {
		p := filepath.Join(c.Dir, name)
		if dirExists(p) {
			return p
		}
	}
	return ""
}

// findPipBin returns the pip binary path inside venvDir (Unix or Windows layout).
func (c *DepsCheck) findPipBin(venvDir string) string {
	unix := filepath.Join(venvDir, "bin", "pip")
	if _, err := os.Stat(unix); err == nil {
		return unix
	}
	return filepath.Join(venvDir, "Scripts", "pip.exe")
}

// findMissingRequirements returns package names listed in requirements.txt
// that are absent from `pip freeze` output.
func findMissingRequirements(pipBin, reqFile string, freezeFn func(string) ([]byte, error)) ([]string, error) {
	required, err := parseRequirements(reqFile)
	if err != nil {
		return nil, fmt.Errorf("reading requirements.txt: %w", err)
	}
	out, err := freezeFn(pipBin)
	if err != nil {
		return nil, fmt.Errorf("running pip freeze: %w", err)
	}
	installed := parseFreeze(out)

	var missing []string
	for pkg := range required {
		if _, ok := installed[pkg]; !ok {
			missing = append(missing, pkg)
		}
	}
	return missing, nil
}

// parseRequirements reads requirements.txt and returns a set of lowercase
// package names, stripping version specifiers, extras, markers, and comments.
func parseRequirements(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pkgs := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip blanks, comments, and pip options (e.g. -r, --index-url).
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		// Strip inline comments.
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		// Extract bare package name before any version specifier, extra, or marker.
		name := strings.FieldsFunc(line, func(r rune) bool {
			return r == '=' || r == '!' || r == '<' || r == '>' || r == '[' || r == ';' || r == ' '
		})[0]
		pkgs[strings.ToLower(name)] = struct{}{}
	}
	return pkgs, scanner.Err()
}

// parseFreeze parses `pip freeze` output into a set of lowercase package names.
func parseFreeze(output []byte) map[string]struct{} {
	installed := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Editable installs: "-e git+...#egg=pkgname"
		if strings.HasPrefix(line, "-e ") {
			if idx := strings.Index(line, "#egg="); idx >= 0 {
				installed[strings.ToLower(strings.TrimSpace(line[idx+5:]))] = struct{}{}
			}
			continue
		}
		parts := strings.SplitN(line, "==", 2)
		installed[strings.ToLower(strings.TrimSpace(parts[0]))] = struct{}{}
	}
	return installed
}

func (c *DepsCheck) runRuby() Result {
	// vendor/bundle is the canonical signal that bundle install --path has been run.
	// Gemfile.lock is the fallback: it exists once bundle install has succeeded at least once.
	vendorBundle := filepath.Join(c.Dir, "vendor", "bundle")
	gemfileLock := filepath.Join(c.Dir, "Gemfile.lock")

	if dirExists(vendorBundle) {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "vendor/bundle directory exists; Ruby gems are installed",
		}
	}
	if _, err := os.Stat(gemfileLock); err == nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Gemfile.lock exists; Ruby gems have been installed",
		}
	}
	return Result{
		Name:    c.Name(),
		Status:  StatusFail,
		Message: "Gemfile.lock not found and vendor/bundle directory missing",
		Fix:     "run `bundle install` to install Ruby gems",
	}
}

func (c *DepsCheck) runGo() Result {
	vendorDir := filepath.Join(c.Dir, "vendor")
	if dirExists(vendorDir) {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "vendor directory exists; Go dependencies are vendored",
		}
	}
	check := c.goCheck
	if check == nil {
		check = goModDownload
	}
	if err := check(c.Dir); err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("Go module cache not populated: %v", err),
			Fix:     "run `go mod download` to download Go module dependencies",
		}
	}
	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Go module cache is populated",
	}
}