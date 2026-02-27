package check

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

type EnvCheck struct {
	Dir string
}

func (c *EnvCheck) Name() string {
	return ".env has all required keys"
}

func (c *EnvCheck) Run(_ context.Context) Result {
	exampleKeys, err := parseEnvKeys(c.Dir + "/.env.example")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "could not read .env.example",
		}
	}

	actualKeys, err := parseEnvKeys(c.Dir + "/.env")
	if err != nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: ".env file not found",
			Fix:     "copy .env.example to .env and fill in the values",
		}
	}

	var missing []string
	for key := range exampleKeys {
		if _, ok := actualKeys[key]; !ok {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: fmt.Sprintf("missing keys: %s", strings.Join(missing, ", ")),
			Fix:     "add the missing keys to your .env file",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("all %d keys present", len(exampleKeys)),
	}
}

func parseEnvKeys(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	keys := make(map[string]struct{})
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, _, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if key != "" {
			keys[key] = struct{}{}
		}
	}
	return keys, scanner.Err()
}
