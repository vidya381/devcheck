package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"sync"

	"github.com/spf13/cobra"
	"github.com/vidya381/devcheck/internal/check"
	"github.com/vidya381/devcheck/internal/config"
	"github.com/vidya381/devcheck/internal/detector"
	"github.com/vidya381/devcheck/internal/reporter"
)

// version is set by the release workflow via -ldflags. For `go install`
// builds it stays "dev" and buildVersion falls back to the module version.
var version = "dev"

// buildVersion returns the ldflags version when one was injected, otherwise
// the version the module system recorded. That covers `go install pkg@vX.Y.Z`
// and gives a pseudo-version for local builds.
func buildVersion() string {
	if version != "dev" && version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return version
}

// errCheckFailed is returned by run when --ci is set and a check failed.
// It is a signal, not a real error, so main exits 1 without printing it.
var errCheckFailed = errors.New("checks failed")

var (
	flagVerbose bool
	flagJSON    bool
	flagFix     bool
	flagCI      bool
)

func main() {
	root := &cobra.Command{
		Use:     "devcheck",
		Short:   "Check if your dev environment is ready to run this project",
		Version: buildVersion(),
		RunE:    run,
	}

	root.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show all checks including skipped")
	root.Flags().BoolVar(&flagJSON, "json", false, "Output results as JSON")
	root.Flags().BoolVar(&flagFix, "fix", false, "Show suggested fix for each failure")
	root.Flags().BoolVar(&flagCI, "ci", false, "Exit with code 1 on any failure (for CI pipelines)")

	root.SilenceErrors = true
	root.SilenceUsage = true

	if err := root.Execute(); err != nil {
		if !errors.Is(err, errCheckFailed) {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	dir, _ := os.Getwd()

	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	stack := detector.Detect(dir)
	checks := check.Build(stack)

	// Append extra binary checks from devcheck.yml require:
	for _, binary := range cfg.Require {
		checks = append(checks, &check.BinaryCheck{Binary: binary})
	}

	// Filter out checks the user asked to skip:
	if len(cfg.Skip) > 0 {
		skipSet := cfg.SkipSet()
		filtered := checks[:0]
		for _, c := range checks {
			if _, skip := skipSet[c.Name()]; !skip {
				filtered = append(filtered, c)
			}
		}
		checks = filtered
	}

	results := make([]check.Result, len(checks))
	var wg sync.WaitGroup

	for i, c := range checks {
		wg.Add(1)
		go func(i int, c check.Check) {
			defer wg.Done()
			results[i] = c.Run(context.Background())
		}(i, c)
	}
	wg.Wait()

	if flagJSON {
		reporter.RenderJSON(results)
	} else {
		reporter.Render(results, flagFix, flagVerbose)
	}

	if flagCI {
		for _, r := range results {
			if r.Status == check.StatusFail {
				return errCheckFailed
			}
		}
	}

	return nil
}
