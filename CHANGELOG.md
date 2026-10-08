# Changelog

## Unreleased

## v0.1.0

First release.

- Detects Go, Node, Python, Ruby, Rust, Java, Docker and Docker Compose from project files, and Postgres, MySQL, MongoDB and Redis from environment variables.
- Checks binaries are on PATH: go, golangci-lint, node, npm/pnpm/yarn, python3, pip, ruby, bundler, rustc, cargo, java, mvn, gradle, docker.
- Version checks for Go (`go.mod`), Node (`.nvmrc` or `engines.node`) and Rust (`rust-toolchain.toml`).
- Dependency checks for Go, Node, Python and Ruby.
- Checks every path under `use` in `go.work` exists and has a `go.mod`.
- Checks `.env` has every key from `.env.example`, and warns on empty values. Handles quoted and multi-line values.
- Checks `.gitignore` covers `.env`, `*.log`, and `node_modules` or `__pycache__`.
- Checks for configured git hooks: `.husky` for Node, `pre-commit` for Python.
- Checks the Docker daemon responds, Compose services are running and Compose images are pulled.
- Port and ping checks for Postgres, MySQL, MongoDB and Redis.
- `devcheck.yml` with `require` for extra binaries and `skip` to drop checks by name.
- Flags: `--fix`, `--json`, `--ci`, `--verbose`, `--version`.
- Release binaries for linux, darwin and windows with `checksums.txt`, an install script, and a GitHub Action.
