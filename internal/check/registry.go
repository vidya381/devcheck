package check

import (
	"github.com/vidya381/devcheck/internal/checks"
	"github.com/vidya381/devcheck/internal/detector"
)

func Build(stack detector.DetectedStack) []Check {
	var cs []Check

	if stack.Go {
		cs = append(cs, &checks.BinaryCheck{Binary: "go"})
	}
	if stack.Node {
		cs = append(cs, &checks.BinaryCheck{Binary: "node"})
		cs = append(cs, &checks.BinaryCheck{Binary: "npm"})
	}
	if stack.Python {
		cs = append(cs, &checks.BinaryCheck{Binary: "python3"})
		cs = append(cs, &checks.BinaryCheck{Binary: "pip"})
	}
	if stack.Java {
		cs = append(cs, &checks.BinaryCheck{Binary: "java"})
		if stack.Maven {
			cs = append(cs, &checks.BinaryCheck{Binary: "mvn"})
		}
		if stack.Gradle {
			cs = append(cs, &checks.BinaryCheck{Binary: "gradle"})
		}
	}
	if stack.Docker {
		// add Docker checks
	}
	if stack.Postgres {
		// add Postgres checks
	}
	if stack.Redis {
		// add Redis checks
	}

	// always run env check if .env.example exists
	// cs = append(cs, &EnvCheck{})

	return cs
}
