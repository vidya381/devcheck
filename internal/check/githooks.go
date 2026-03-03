package check

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var lookPath = exec.LookPath

type GitHooksCheck struct {
	Dir   string
	Stack string // "node" or "python"
}

func (c *GitHooksCheck) Name() string {
	switch c.Stack {
	case "node":
		return "Git hooks configured for Node"
	case "python":
		return "Git hooks configured for Python"
	default:
		return "Git hooks configured"
	}
}

func (c *GitHooksCheck) Run(_ context.Context) Result {
	if !dirExists(filepath.Join(c.Dir, ".git")) {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkipped,
			Message: "not a git repository (no .git directory)",
		}
	}

	switch c.Stack {
	case "node":
		return c.runNode()
	case "python":
		return c.runPython()
	default:
		return Result{
			Name:    c.Name(),
			Status:  StatusSkipped,
			Message: "unknown stack type for git hooks check",
		}
	}
}

func (c *GitHooksCheck) runNode() Result {
	huskyDir := filepath.Join(c.Dir, ".husky")
	if dirExists(huskyDir) {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: ".husky directory exists; git hooks are configured",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusWarn,
		Message: ".husky directory not found; git hooks for Node are not configured",
		Fix:     "set up Husky (or another git hooks tool) to run linting/tests before commits",
	}
}

func (c *GitHooksCheck) runPython() Result {
	configPath := filepath.Join(c.Dir, ".pre-commit-config.yaml")
	_, err := os.Stat(configPath)
	configExists := err == nil

	_, err = lookPath("pre-commit")
	preCommitInstalled := err == nil

	if configExists && preCommitInstalled {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "pre-commit is installed and .pre-commit-config.yaml exists",
		}
	}

	var details string
	if !configExists && !preCommitInstalled {
		details = ".pre-commit-config.yaml not found and pre-commit is not installed"
	} else if !configExists {
		details = ".pre-commit-config.yaml not found"
	} else {
		details = "pre-commit is not installed"
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusWarn,
		Message: fmt.Sprintf("git hooks for Python are not fully configured: %s", details),
		Fix:     "install pre-commit and add a .pre-commit-config.yaml so hooks can run linting/tests before commits",
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

