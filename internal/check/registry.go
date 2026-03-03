package check

import (
	"os"

	"github.com/vidya381/devcheck/internal/detector"
)

func Build(stack detector.DetectedStack) []Check {
	var cs []Check

	if stack.Go {
		cs = append(cs, &BinaryCheck{Binary: "go"})
		cs = append(cs, &BinaryCheck{Binary: "golangci-lint"})
		cs = append(cs, &GoVersionCheck{Dir: "."})
	}
	if stack.Node {
		cs = append(cs, &BinaryCheck{Binary: "node"})
		cs = append(cs, &BinaryCheck{Binary: "npm"})
		cs = append(cs, &NodeVersionCheck{Dir: "."})
	}
	if stack.Python {
		cs = append(cs, &BinaryCheck{Binary: "python3"})
		cs = append(cs, &BinaryCheck{Binary: "pip"})
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
