package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type DetectedStack struct {
	Go       bool
	GoWork   bool
	Node     bool
	// PackageManager is the Node package manager inferred from the lockfile.
	// Possible values: "npm", "pnpm", "yarn". Empty string when Node is false.
	PackageManager string
	Python   bool
	Ruby     bool
	Java     bool
	Maven    bool
	Gradle   bool
	Docker        bool
	DockerCompose bool
	Postgres bool
	Redis    bool
	MySQL    bool
	MongoDB    bool
	EnvExample bool
}

func Detect(dir string) DetectedStack {
	stack := DetectedStack{}

	stack.Go = fileExists(filepath.Join(dir, "go.mod"))
	stack.GoWork = fileExists(filepath.Join(dir, "go.work"))
	stack.Node = fileExists(filepath.Join(dir, "package.json"))
	if stack.Node {
		switch {
		case fileExists(filepath.Join(dir, "pnpm-lock.yaml")):
			stack.PackageManager = "pnpm"
		case fileExists(filepath.Join(dir, "yarn.lock")):
			stack.PackageManager = "yarn"
		default:
			stack.PackageManager = "npm"
		}
	}
	stack.Python = fileExists(filepath.Join(dir, "requirements.txt")) ||
		fileExists(filepath.Join(dir, "pyproject.toml"))
	stack.Ruby = fileExists(filepath.Join(dir, "Gemfile"))
	stack.Maven = fileExists(filepath.Join(dir, "pom.xml"))
	stack.Gradle = fileExists(filepath.Join(dir, "build.gradle"))
	stack.Java = stack.Maven || stack.Gradle
	stack.DockerCompose = fileExists(filepath.Join(dir, "docker-compose.yml")) ||
		fileExists(filepath.Join(dir, "docker-compose.yaml")) ||
		fileExists(filepath.Join(dir, "compose.yml")) ||
		fileExists(filepath.Join(dir, "compose.yaml"))
	stack.Docker = fileExists(filepath.Join(dir, "Dockerfile")) || stack.DockerCompose

	dbURL := os.Getenv("DATABASE_URL")
	stack.Postgres = strings.Contains(dbURL, "postgres")
	stack.MySQL = strings.Contains(dbURL, "mysql")
	stack.MongoDB = os.Getenv("MONGODB_URI") != "" || os.Getenv("MONGO_URL") != ""
	stack.Redis = os.Getenv("REDIS_URL") != "" || os.Getenv("REDIS_URI") != ""
	stack.EnvExample = fileExists(filepath.Join(dir, ".env.example"))

	return stack
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
