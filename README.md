# devcheck

[![CI](https://github.com/vidya381/devcheck/actions/workflows/ci.yml/badge.svg)](https://github.com/vidya381/devcheck/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/vidya381/devcheck)](https://github.com/vidya381/devcheck/releases) [![Go Report Card](https://goreportcard.com/badge/github.com/vidya381/devcheck)](https://goreportcard.com/report/github.com/vidya381/devcheck) [![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

devcheck looks at the files in a project, works out what stack it uses, and tells you what is missing or not running before you try to start it.

## Output

Run it in a project directory with no arguments.

```
$ devcheck
✅  node is installed
✅  npm is installed
✅  Node 22.23.2 installed (need 20.11.0)
❌  node_modules directory not found
⚠️   .husky directory not found; git hooks for Node are not configured
✅  docker is installed
❌  Docker daemon is not running
❌  missing keys: API_KEY
⚠️   Missing 1 sensitive file pattern
────────────────────────────────────────
  4 passed  2 warnings  3 failed
  Run `devcheck --fix` for suggestions
```

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/vidya381/devcheck/main/scripts/install.sh | bash
```

Or with Go:

```bash
go install github.com/vidya381/devcheck/cmd/devcheck@latest
```

Binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 and windows/amd64 are on the [releases page](https://github.com/vidya381/devcheck/releases).

## Usage

```
devcheck              run every check that applies to this project
devcheck --fix        print a suggested fix under each failure
devcheck --json       print results as JSON
devcheck --ci         exit 1 if any check failed
devcheck --verbose    also print checks that were skipped
devcheck --version    print the version
```

## What it checks

Nothing is configured. devcheck picks checks based on what it finds in the directory.

| Detected from | Checks |
|---|---|
| `go.mod` | `go` on PATH, version against the `go` directive, module cache populated or `vendor/` present |
| `.golangci.yml` | `golangci-lint` on PATH |
| `go.work` | every path under `use` exists and has a `go.mod` |
| `package.json` | `node` and the right package manager on PATH, version against `.nvmrc` or `engines.node`, `node_modules` present, `.husky` present |
| `requirements.txt` / `pyproject.toml` | `python3` and `pip` on PATH, `venv`/`.venv` present, `pip check` clean, every requirement installed, `.pre-commit-config.yaml` and `pre-commit` present |
| `Gemfile` | `ruby` and `bundler` on PATH, `Gemfile.lock` or `vendor/bundle` present |
| `Cargo.toml` | `rustc` and `cargo` on PATH, version against `rust-toolchain.toml` |
| `pom.xml` / `build.gradle` | `java` on PATH, plus `mvn` or `gradle` |
| `Dockerfile` | `docker` on PATH, daemon responding |
| `docker-compose.yml` / `compose.yml` | every service running, every image pulled |
| `DATABASE_URL` containing `postgres` | port open, server answers a ping |
| `DATABASE_URL` containing `mysql`, or `MYSQL_URL` | port open, server answers a ping |
| `MONGODB_URI` / `MONGO_URL` | port open, server answers a ping |
| `REDIS_URL` / `REDIS_URI` | port open, server answers a ping |
| `.env.example` | every key also exists in `.env`, warn on empty values |
| `.gitignore` | covers `.env`, `*.log`, and `node_modules` or `__pycache__` for those stacks |

The package manager comes from the lockfile: `pnpm-lock.yaml`, `yarn.lock`, or npm otherwise.

### devcheck.yml

Optional. Put it in the project root.

```yaml
require:
  - terraform
  - kubectl
skip:
  - "Git hooks configured for Node"
```

`require` adds a PATH check for each binary. `skip` drops checks by the name they report.

## Use in CI

```yaml
- uses: vidya381/devcheck@v0.1.0
```

Or with inputs:

```yaml
- uses: vidya381/devcheck@v0.1.0
  with:
    version: v0.1.0
    args: --fix
```

The action downloads the release binary for the runner and runs `devcheck --ci`, so the job fails if any check fails. If you would rather install it yourself:

```yaml
- run: go install github.com/vidya381/devcheck/cmd/devcheck@latest
- run: devcheck --ci
```

## Why I made it

Every time I set up a project I had not touched in a while, I lost half an hour to something small. Postgres not running, a key missing from `.env`, the wrong Node version. The error you get is never the error you have. I wanted one command that tells me all of it at once, before I run anything.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues labeled `good first issue` are the easiest place to start, and most of them are a single new check with a test.

## License

MIT. See [LICENSE](LICENSE).
