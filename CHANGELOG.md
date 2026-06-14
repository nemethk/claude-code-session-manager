# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `ccsm list` — list all sessions sorted by last active, with `--project`, `--since`, and `--json` flags
- `ccsm search <term>` — filter sessions by first message or project path, with `--json` flag
- `ccsm show <uuid>` — print first N user turns from a session, with `--turns` flag
- UUID prefix matching — short prefixes work everywhere (e.g. `ccsm show 2803b936`)
- `CCSM_SESSIONS_DIR` env override for isolated testing
- `/sessions` Claude Code skill for natural language session search
- goreleaser configuration for multi-platform releases (linux/darwin × amd64/arm64)
- Homebrew tap support
- Comprehensive test suite (unit + end-to-end)

### Security
- 4 MB scanner buffer prevents OOM on sessions with large tool outputs
- Sessions directory is read-only — `ccsm` never writes to `~/.claude/`

### Planned
- `ccsm delete <uuid>` — remove a session file
- Summary caching — pre-generate one-line summaries in `~/.claude/sessions-cache.json`
- Windows support

---

For development history, see [git log](https://github.com/nemethk/ccsm/commits/main).
