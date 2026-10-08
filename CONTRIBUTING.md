# Contributing

New checks, bug fixes and docs are all welcome. I reply to issues and pull requests within a day or two.

## Build and test

```bash
git clone https://github.com/YOUR_USERNAME/devcheck
cd devcheck
go build ./...
go test ./...
go vet ./...
```

`go build -o devcheck ./cmd/devcheck` gives you a binary to try in another directory. The repo gitignores `/devcheck` so it will not show up in your diff.

## Adding a check

Every check is a struct with two methods, in `internal/check/`. Read [`internal/check/binary.go`](internal/check/binary.go) first, it is the smallest one. [`internal/check/gowork.go`](internal/check/gowork.go) is a good second example: it parses a file and reports which entries are missing.

1. Add `internal/check/yourcheck.go` in package `check`:

```go
type YourCheck struct {
	Dir string
}

func (c *YourCheck) Name() string { return "your check name" }

func (c *YourCheck) Run(_ context.Context) Result {
	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "what is true right now",
		Fix:     "", // printed under the failure when --fix is passed
	}
}
```

2. Register it in [`internal/check/registry.go`](internal/check/registry.go), inside the `if` for the stack it belongs to. If it needs a file that may not exist, guard it with `fileExists`.
3. If the check needs to detect something new, add the field to `DetectedStack` in [`internal/detector/detector.go`](internal/detector/detector.go) and set it in `Detect`.
4. Add `internal/check/yourcheck_test.go`.

Statuses: `StatusPass`, `StatusWarn` for something worth mentioning that will not stop you working, `StatusFail` for something that will, and `StatusSkipped` when the check could not run at all. Skipped results are hidden unless `--verbose` is passed.

Messages are lowercase and say what is true, not what to do. The `Fix` field is where the instruction goes.

## Tests

Cover at least the pass case and the fail case. Use `t.TempDir()` for anything that reads files.

Do not make a test depend on a running service. Checks that talk to something take an injectable function so you can stub it: see `dialer` in `port.go` and `pinger` in `redis.go`. For a real connection failure without a server, dial a closed port on localhost.

## Pull requests

Fill in the checklist in the template. Link the issue in the description, for example "Closes #31". Comment on an issue before you start so two people do not write the same check.

## Labels

- `good first issue` — self-contained, nothing to ask me first
- `checker` — adds or changes a check
- `help wanted` — needs a judgement call or a service to test against
- `enhancement` — a feature that is not a check
- `bug` — something is broken
- `documentation` — docs only

## Questions

Open a Discussion or comment on the issue.
