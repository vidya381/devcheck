package detector

import (
	"os"
	"path/filepath"
	"strings"
)

type DetectedStack struct {
	Go       bool
	Node     bool
	Python   bool
	Java     bool
	Maven    bool
	Gradle   bool
	Docker   bool
	Postgres bool
	Redis    bool
	MySQL    bool
	MongoDB  bool
}

func Detect(dir string) DetectedStack {
	stack := DetectedStack{}

	stack.Go = fileExists(filepath.Join(dir, "go.mod"))
	stack.Node = fileExists(filepath.Join(dir, "package.json"))
	stack.Python = fileExists(filepath.Join(dir, "requirements.txt")) ||
		fileExists(filepath.Join(dir, "pyproject.toml"))
	stack.Maven = fileExists(filepath.Join(dir, "pom.xml"))
	stack.Gradle = fileExists(filepath.Join(dir, "build.gradle"))
	stack.Java = stack.Maven || stack.Gradle
	stack.Docker = fileExists(filepath.Join(dir, "Dockerfile")) ||
		fileExists(filepath.Join(dir, "docker-compose.yml")) ||
		fileExists(filepath.Join(dir, "docker-compose.yaml"))

	dbURL := os.Getenv("DATABASE_URL")
	stack.Postgres = strings.Contains(dbURL, "postgres")
	stack.MySQL = strings.Contains(dbURL, "mysql")
	stack.MongoDB = os.Getenv("MONGODB_URI") != "" || os.Getenv("MONGO_URL") != ""
	stack.Redis = os.Getenv("REDIS_URL") != "" || os.Getenv("REDIS_URI") != ""

	return stack
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
