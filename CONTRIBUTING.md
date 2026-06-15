# Contributing

Contributions are welcome — bug fixes, new features, documentation improvements.

---

## Development Setup

**Prerequisites:**
- Go 1.26 or later
- Git

**Clone and build:**

```bash
git clone git@github.com:nemethk/claude-code-session-manager.git
cd claude-code-session-manager
go mod download
make build
```

**Verify:**

```bash
./dist/ccsm --version
```

**Install locally:**

```bash
GOBIN=/usr/local/bin make install
ccsm --version
```

---

## Running Tests

### Unit Tests

Go unit tests cover the `internal/session` package — JSONL parsing, content extraction, and session filtering.

```bash
# all unit tests
make test

# with verbose output
make test-verbose

# specific package
go test ./internal/session/...
```

### End-to-End Tests

The `tests/` package builds and exercises the real binary. Each test creates an isolated sessions directory via `CCSM_SESSIONS_DIR` and asserts on command output.

```bash
# run e2e tests (builds binary automatically)
make test-e2e

# run everything — unit + e2e
make test-all
```

| File | What it tests |
|------|---------------|
| `main_test.go` | `TestMain` — builds binary before running tests |
| `list_test.go` | `list`, `list --json`, `list --project`, `search`, `show`, `show` not found |

All tests must pass before submitting a PR.

---

## Project Structure

```
ccsm/
├── cmd/                  # CLI commands (cobra)
│   ├── root.go           # root command, Execute()
│   ├── list.go           # ccsm list
│   ├── search.go         # ccsm search
│   ├── show.go           # ccsm show
│   └── util.go           # shared helpers
├── internal/
│   └── session/          # JSONL parsing, session extraction
│       ├── session.go
│       └── session_test.go
├── tests/                # e2e tests against the compiled binary
│   ├── main_test.go
│   └── list_test.go
├── skill/                # Claude Code /sessions skill
│   └── sessions.md
├── assets/               # logo and static assets
├── main.go
└── go.mod
```

---

## Making Changes

**1. Fork and create a branch**

```bash
git checkout -b feat/my-feature
```

**2. Write code and tests**

- add tests for new functionality
- run `make test-all` before committing

**3. Follow commit message conventions**

```
feat: short description of new feature
fix: short description of bug fix
docs: documentation changes
test: adding or updating tests
chore: maintenance, dependencies
```

These feed into automatic release notes — keep them clear and descriptive.

**4. Open a PR**

- describe what the change does and why
- link any related issues
- make sure CI is green

---

## Code Style

- run `go fmt ./...` before committing
- follow standard Go conventions
- keep functions small and focused
- add comments only when the **why** is non-obvious

---

## Reporting Bugs

Open an issue with:
- `ccsm --version`
- OS and shell
- steps to reproduce
- expected vs actual behaviour
