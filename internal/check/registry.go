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
		if stack.GoWork {
			cs = append(cs, &GoWorkCheck{Dir: "."})
		}
	}
	if stack.Node {
		pm := stack.PackageManager
		if pm == "" {
			pm = "npm"
		}
		cs = append(cs, &BinaryCheck{Binary: "node"})
		cs = append(cs, &BinaryCheck{Binary: pm})
		cs = append(cs, &NodeVersionCheck{Dir: "."})
		cs = append(cs, &DepsCheck{Dir: ".", Stack: "node", PackageManager: pm})
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
		cs = append(cs, &ComposeImageCheck{})
	}
	if stack.Postgres {
		dbURL := os.Getenv("DATABASE_URL")
		cs = append(cs, &PortCheck{Service: "PostgreSQL", Port: portFromURL(dbURL, "5432")})
		cs = append(cs, &PostgresCheck{URL: dbURL})
	}
	if stack.MySQL {
		cs = append(cs, &MySQLCheck{URL: os.Getenv("MYSQL_URL")})
	}
	if stack.MongoDB {
		url := os.Getenv("MONGODB_URI")
		if url == "" {
			url = os.Getenv("MONGO_URL")
		}
		cs = append(cs, &MongoCheck{URL: url})
	}
	if stack.Redis {
		redisURL := os.Getenv("REDIS_URL")
		if redisURL == "" {
			redisURL = os.Getenv("REDIS_URI")
		}
		cs = append(cs, &PortCheck{Service: "Redis", Port: portFromURL(redisURL, "6379")})
		cs = append(cs, &RedisCheck{URL: redisURL})
	}
	if stack.MySQL {
		cs = append(cs, &PortCheck{Service: "MySQL", Port: portFromURL(os.Getenv("MYSQL_URL"), "3306")})
	}
	if stack.MongoDB {
		mongoURL := os.Getenv("MONGODB_URI")
		if mongoURL == "" {
			mongoURL = os.Getenv("MONGO_URL")
		}
		cs = append(cs, &PortCheck{Service: "MongoDB", Port: portFromURL(mongoURL, "27017")})
	}

	if stack.EnvExample {
		cs = append(cs, &EnvCheck{Dir: "."})
	}


	// Check .gitignore for sensitive file patterns
	if fileExists(".gitignore") {
		gitignorePatterns := []requiredPattern{
			{pattern: ".env", description: "Environment files with secrets"},
			{pattern: "*.log", description: "Log files"},
		}
		if stack.Node {
			gitignorePatterns = append(gitignorePatterns, requiredPattern{
				pattern:     "node_modules",
				description: "Node.js dependencies",
			})
		}
		if stack.Python {
			gitignorePatterns = append(gitignorePatterns, requiredPattern{
				pattern:     "__pycache__",
				description: "Python cache",
			})
		}
		cs = append(cs, &GitignoreCheck{Dir: ".", Patterns: gitignorePatterns})
	}

	return cs
}

// fileExists checks if a file exists in the current directory
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
