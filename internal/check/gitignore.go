package check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

type requiredPattern struct {
	pattern     string
	description string
}

type GitignoreCheck struct {
	Dir      string
	Patterns []requiredPattern
}

func (c *GitignoreCheck) Name() string {
	return ".gitignore covers sensitive files"
}

func pluralPatterns(n int) string {
	if n == 1 {
		return "pattern"
	}
	return "patterns"
}

func (c *GitignoreCheck) Run(_ context.Context) Result {
	gitignorePath := c.Dir + "/.gitignore"

	// Read .gitignore content
	file, err := os.Open(gitignorePath)
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: ".gitignore not found",
			Fix:     "Create a .gitignore file with common patterns",
		}
	}
	defer file.Close()

	// Parse .gitignore and collect all patterns
	patterns := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns[line] = true
	}

	// Check which patterns are missing
	var missing []string
	for _, rp := range c.Patterns {
		// Check if pattern or a variant exists
		found := false
		for existingPattern := range patterns {
			// Match exact pattern or with trailing slash (for directories)
			if existingPattern == rp.pattern ||
				existingPattern == rp.pattern+"/" ||
				existingPattern == "**/"+rp.pattern ||
				existingPattern == "**/"+rp.pattern+"/" {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, fmt.Sprintf("%s (%s)", rp.pattern, rp.description))
		}
	}

	if len(missing) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("Missing %d sensitive file %s", len(missing), pluralPatterns(len(missing))),
			Fix:     fmt.Sprintf("Add these patterns to .gitignore:\n  %s", strings.Join(missing, "\n  ")),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Common sensitive files are gitignored",
	}
}
