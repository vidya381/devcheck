package check

import (
	"os"

	"github.com/vidya381/devcheck/internal/detector"
)

func Build(stack detector.DetectedStack) []Check {
	var cs []Check

	if stack.Go {
		cs = append(cs, &BinaryCheck{Binary: "go"})
		// Only check golangci-lint if project has linting configuration
		if fileExists(".golangci.yml") || fileExists(".golangci.yaml") {
			cs = append(cs, &BinaryCheck{Binary: "golangci-lint"})
		}
		cs = append(cs, &GoVersionCheck{Dir: "."})
		cs = append(cs, &DepsCheck{Dir: ".", Stack: "go"})
	}
	if stack.Node {
		cs = append(cs, &BinaryCheck{Binary: "node"})
		cs = append(cs, &BinaryCheck{Binary: "npm"})
		cs = append(cs, &NodeVersionCheck{Dir: "."})
		cs = append(cs, &DepsCheck{Dir: ".", Stack: "node"})
		cs = append(cs, &GitHooksCheck{Dir: ".", Stack: "node"})
	}
	if stack.Python {
		cs = append(cs, &BinaryCheck{Binary: "python3"})
		cs = append(cs, &BinaryCheck{Binary: "pip"})
		cs = append(cs, &DepsCheck{Dir: ".", Stack: "python"})
		cs = append(cs, &GitHooksCheck{Dir: ".", Stack: "python"})
	}
	if stack.Java {
		cs = append(cs, &BinaryCheck{Binary: "java"})
		if stack.Maven {
			cs = append(cs, &BinaryCheck{Binary: "mvn"})
		}
		if stack.Gradle {
			cs = append(cs, &BinaryCheck{Binary: "gradle"})
		}
	}
	if stack.Docker {
		cs = append(cs, &BinaryCheck{Binary: "docker"})
		cs = append(cs, &DockerDaemonCheck{})
	}
	if stack.DockerCompose {
		cs = append(cs, &ComposeCheck{})
	}
	if stack.Postgres {
		cs = append(cs, &PostgresCheck{URL: os.Getenv("DATABASE_URL")})
	}
	if stack.Redis {
		url := os.Getenv("REDIS_URL")
		if url == "" {
			url = os.Getenv("REDIS_URI")
		}
		cs = append(cs, &RedisCheck{URL: url})
	}

	if stack.EnvExample {
		cs = append(cs, &EnvCheck{Dir: "."})
	}

	return cs
}

// fileExists checks if a file exists in the current directory
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
