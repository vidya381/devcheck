package check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoWorkCheck reads go.work and verifies that every module listed under a
// "use" directive exists on disk as a directory containing a go.mod file.
type GoWorkCheck struct {
	Dir string
}

func (c *GoWorkCheck) Name() string {
	return "Go workspace modules present"
}

func (c *GoWorkCheck) Run(_ context.Context) Result {
	workFile := filepath.Join(c.Dir, "go.work")
	paths, err := parseGoWorkUse(workFile)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("could not read go.work: %v", err),
		}
	}

	if len(paths) == 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: "go.work has no use directives",
		}
	}

	var missing []string
	for _, p := range paths {
		modPath := filepath.Join(c.Dir, p, "go.mod")
		if _, err := os.Stat(modPath); os.IsNotExist(err) {
			missing = append(missing, p)
		}
	}

	if len(missing) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("go.work references modules with missing go.mod: %s", strings.Join(missing, ", ")),
			Fix:     "ensure each path listed under 'use' in go.work exists and contains a go.mod file",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("all %d go.work module(s) present", len(paths)),
	}
}

// parseGoWorkUse returns the list of paths from "use" directives in go.work.
// It handles both single-line form ("use ./foo") and block form ("use (\n./foo\n)").
func parseGoWorkUse(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var paths []string
	inBlock := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Strip inline comments.
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		if line == "" {
			continue
		}

		if inBlock {
			if line == ")" {
				inBlock = false
				continue
			}
			paths = append(paths, line)
			continue
		}

		if strings.HasPrefix(line, "use") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "use"))
			if rest == "(" {
				inBlock = true
				continue
			}
			// Block opener on same line: "use (" already handled above;
			// single path form: "use ./foo"
			if rest != "" {
				paths = append(paths, rest)
			}
		}
	}

	return paths, scanner.Err()
}