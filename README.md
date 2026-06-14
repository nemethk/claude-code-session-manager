<div align="center">
  <img src="assets/ccsm-450x300.png" width="300" alt="ccsm" />

  <h1>Claude Code Session Manager — ccsm</h1>

  <p>List, search, and summarize your saved Claude Code sessions from the terminal.</p>

  <p>
    <a href="https://github.com/nemethk/claude-code-session-manager/releases"><img src="https://img.shields.io/github/v/release/nemethk/claude-code-session-manager" alt="Latest Release" /></a>
    <a href="https://github.com/nemethk/claude-code-session-manager/actions/workflows/ci.yaml"><img src="https://img.shields.io/github/actions/workflow/status/nemethk/claude-code-session-manager/ci.yaml?label=ci" alt="CI" /></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/nemethk/claude-code-session-manager" alt="License" /></a>
    <img src="https://img.shields.io/badge/go-1.26-00ADD8?logo=go" alt="Go 1.26" />
    <img src="https://img.shields.io/badge/platform-linux%20%7C%20macOS-blue" alt="Platform" />
    <a href="https://claude.ai/claude-code"><img src="https://img.shields.io/badge/Claude%20Code-enabled-8A2BE2?logo=anthropic" alt="Claude Code" /></a>
  </p>
</div>

---

## ❓ The Challenge

Claude Code saves every session locally — but finding a past conversation requires guessing a UUID or scrolling through `~/.claude/projects/`. There is no built-in way to search or list what you've worked on.

`ccsm` solves that.

---

## ⚙️ How It Works

Sessions are stored as JSONL files in `~/.claude/projects/<project-slug>/<uuid>.jsonl`. `ccsm` walks that directory, parses the messages, and presents them as a searchable, summarizable list.

```bash
ccsm list

DATE        TIME   UUID                                  PROJECT              SUMMARY
2026-06-15  10:12  2803b936-7cb5-499f-804a-3804351f4b94  ~/Dev/api-service    debug the query planner for slow migration
2026-06-14  16:30  cc928331-e2fa-4542-992e-f1fded2deb08  ~/Dev/api-service    refactor the payment service auth flow
2026-06-14  09:23  a1b2c3d4-0f8e-4a12-bb23-9c8d7e6f5a4b  ~/Dev/myapp          add rate limiting middleware
```

---

## 💡 Why It Matters

| Feature | Benefit |
|---------|---------|
| SUMMARY column | cleaned first message with key file references — no raw UUIDs |
| `--ai` flag | Claude-generated one-liner per session, cached in `~/.cache/ccsm/` |
| Full-text search | finds sessions by content in any turn, not just the first message |
| TURN + MATCH output | shows exactly where in the conversation a search term appeared |
| UUID prefix matching | type 8 characters instead of 36, same as `git log` |
| `ccsm summarize` | structured breakdown — what was worked on, key steps, outcome |

---

## 👥 Who Uses It

| Audience | Use case |
|----------|----------|
| Individual developer | find any past session by keyword, date, or project — then resume it |
| Team lead / manager | review what was worked on across projects without opening session files |
| Skill author | inspect past sessions to understand how Claude responded to specific prompts |
| Scripter / automator | pipe `--json` output into jq for custom reports and session dashboards |

See [GUIDE.md](GUIDE.md) for full walkthroughs of all three scenarios.

---

## 🔧 Prerequisites

| Dependency | Required | Purpose |
|---|---|---|
| [Claude Code](https://claude.ai/claude-code) | for `--ai` and `summarize` | Generates AI summaries via `claude` CLI |
| [fzf](https://github.com/junegunn/fzf) | optional | Fuzzy session picker for resume alias |

---

## 📦 Installation

> For a complete step-by-step guide including the fzf alias and `/sessions` skill see [INSTALL.md](INSTALL.md).

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/nemethk/claude-code-session-manager/main/scripts/install.sh | bash
```

**Homebrew:**

```bash
brew install nemethk/tap/ccsm
```

**Go:**

```bash
GOBIN=/usr/local/bin go install github.com/nemethk/claude-code-session-manager@latest
```

---

## 🛠️ Commands

### `ccsm list`

List all sessions, sorted by most recently active.

```bash
ccsm list                          # all sessions
ccsm list --project myapp          # filter by project path substring
ccsm list --since 2026-06-01       # filter by date
ccsm list --min-turns 2            # hide single-message sessions
ccsm list --ai                     # generate AI summaries (cached for future runs)
ccsm list --json                   # machine-readable JSON output
```

### `ccsm search <term>`

Search every user message across all turns in all sessions.

```bash
ccsm search "query planner"
ccsm search postgres --json

DATE        TIME   UUID      PROJECT           TURN  MATCH
2026-06-15  10:12  2803b936  ~/Dev/api-service  #3   ...the query planner is choosing a seq scan...
```

### `ccsm show <uuid>`

Print the first N raw user messages from a session — preview before resuming.

```bash
ccsm show 2803b936                 # UUID prefix is enough
ccsm show 2803b936 --turns 10      # show more turns (default: 5)
```

### `ccsm summarize <uuid>`

Generate a detailed AI summary: what was worked on, key steps, and outcome.

```bash
ccsm summarize 2803b936

Session:  2803b936-7cb5-499f-804a-3804351f4b94
Project:  ~/Dev/api-service
Date:     2026-06-15 10:12
Turns:    31

**What was worked on:** ...
**Key steps:** ...
**Outcome:** ...
```

---

## ▶️ Resume a Session

```bash
claude --resume 2803b936-7cb5-499f-804a-3804351f4b94
```

**With fzf** — fuzzy-pick and resume in one command:

```bash
ccsm list | fzf | awk '{print $3}' | xargs claude --resume
```

Add as a shell alias:

```bash
alias cr='ccsm list | fzf | awk '"'"'{print $3}'"'"' | xargs claude --resume'
```

---

## 🤖 Claude Code Skill

`ccsm` ships with a `/sessions` skill that adds natural language search on top of the binary.

```bash
# install
cp skill/sessions.md ~/.claude/skills/sessions.md
```

```
/sessions                              → numbered list of all sessions
/sessions find postgres migration      → Claude filters by relevance
/sessions show 2803b936                → inspect turns before resuming
/sessions resume 2803b936              → prints the claude --resume command
```

---

## ⚡ Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `CCSM_SESSIONS_DIR` | `~/.claude/projects` | Override the sessions directory |

---

## 🧪 Development

```bash
go test ./...
```

Tests are split into two packages:

- `internal/session` — unit tests for JSONL parsing, path extraction, text matching
- `tests/` — end-to-end tests that build the binary and run it against fixture sessions

---

## 🤝 Contributing

Bug fixes, new features, and documentation improvements are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## 📜 License

[MIT](LICENSE)