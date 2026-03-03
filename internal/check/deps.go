package check

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
)

// goModDownload is a variable so tests can stub it.
var goModDownload = func(dir string) error {
	cmd := exec.Command("go", "mod", "download", "-json", "all")
	cmd.Dir = dir
	return cmd.Run()
}

type DepsCheck struct {
	Dir     string
	Stack   string // "node", "python", or "go"
	goCheck func(dir string) error
}

func (c *DepsCheck) Name() string {
	switch c.Stack {
	case "node":
		return "Node dependencies installed"
	case "python":
		return "Python dependencies installed"
	case "go":
		return "Go dependencies installed"
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

	return Result{
		Name:    c.Name(),
		Status:  StatusFail,
		Message: "node_modules directory not found",
		Fix:     "run `npm install` or `pnpm install` to install Node dependencies",
	}
}

func (c *DepsCheck) runPython() Result {
	venv := filepath.Join(c.Dir, "venv")
	dotVenv := filepath.Join(c.Dir, ".venv")

	if dirExists(venv) || dirExists(dotVenv) {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "Python virtual environment directory exists",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusFail,
		Message: "Python virtual environment directory not found",
		Fix:     "create a virtual environment (e.g. `python -m venv .venv`) and install dependencies with `pip install -r requirements.txt` or equivalent",
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

