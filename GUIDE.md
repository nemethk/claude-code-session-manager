# 📖 Guide

Real-world walkthroughs for three `ccsm` use cases:

1. **🔍 [Daily Navigation — Finding and Resuming Past Work](#scenario-1-daily-navigation--finding-and-resuming-past-work)** — browse your history, preview before resuming, fuzzy-pick with fzf
2. **🤖 [AI Summaries — Understanding What Was Accomplished](#scenario-2-ai-summaries--understanding-what-was-accomplished)** — warm the summary cache, drill into a session, use `ccsm summarize`
3. **⚙️ [Scripting — Automating with JSON Output](#scenario-3-scripting--automating-with-json-output)** — pipe `--json` into jq, build custom reports, integrate with other tools

---

# Scenario 1: 🔍 Daily Navigation — Finding and Resuming Past Work

## The Problem

You worked on a database migration issue three days ago. Claude helped you debug the query planner. You want to pick up that conversation — but you have 40+ sessions across several projects and no way to scan them.

## Step 1: List Recent Sessions

```bash
ccsm list

DATE        TIME   UUID                                  PROJECT             SUMMARY
2026-06-15  10:12  2803b936-7cb5-499f-804a-3804351f4b94  ~/Dev/api-service   debug the query planner for slow migration
2026-06-14  16:30  cc928331-e2fa-4542-992e-f1fded2deb08  ~/Dev/api-service   refactor the payment service auth flow
2026-06-14  09:23  a1b2c3d4-0f8e-4a12-bb23-9c8d7e6f5a4b  ~/Dev/myapp         add rate limiting middleware
...
```

Sessions are sorted newest first. The SUMMARY column shows a cleaned excerpt of your first message.

## Step 2: Filter by Project

Too many results? Narrow by project:

```bash
ccsm list --project api-service
```

Or filter by date:

```bash
ccsm list --since 2026-06-12
```

## Step 3: Search Across All Messages

Can't remember which project? Search by keyword — `ccsm search` scans every message in every session, not just the first one:

```bash
ccsm search "query planner"

DATE        TIME   UUID      PROJECT          TURN  MATCH
2026-06-15  10:12  2803b936  ~/Dev/api-service  #3   ...the query planner is choosing a seq scan...
```

The TURN and MATCH columns tell you exactly where in the conversation the term appeared and what surrounded it.

## Step 4: Preview Before Resuming

Found the session — confirm it's the right one before resuming:

```bash
ccsm show 2803b936

[1] debug the query planner for slow migration
[2] the query planner is choosing a seq scan even with the index present
[3] yes, EXPLAIN ANALYZE output: ...
```

UUID prefix is enough — the same 8-character shorthand git uses.

## Step 5: Resume

```bash
claude --resume 2803b936-7cb5-499f-804a-3804351f4b94
```

Claude picks up exactly where you left off — full conversation history intact.

## Shortcut: fzf Resume in One Command

With the `cr` alias from [INSTALL.md](INSTALL.md):

```bash
cr
```

This pipes `ccsm list` into fzf, lets you fuzzy-search the SUMMARY column, and passes the selected UUID to `claude --resume`. Browse, pick, resume — one keypress to confirm.

## Filtering Out Noise

Sessions with a single message (e.g., aborted starts or `claude -p` calls) add clutter. Hide them:

```bash
ccsm list --min-turns 2
```

## Scenario 1 Takeaways

- `ccsm list --project` and `--since` narrow a long list quickly
- `ccsm search` finds sessions by content anywhere in the conversation, not just the first message
- `ccsm show` is cheap confirmation before committing to `claude --resume`
- UUID prefix matching means you type 8 characters, not 36

---

# Scenario 2: 🤖 AI Summaries — Understanding What Was Accomplished

## The Problem

Your session list shows first-message excerpts — useful for finding sessions you started, but not for understanding what was *resolved*. A session titled "debug the query planner" might have ended with a fix, a workaround, or still open.

## Step 1: Warm the Cache

Generate AI one-liner summaries for all uncached sessions at once:

```bash
ccsm list --ai

Generating AI summaries for 8 session(s)...
  [1/8] 2803b936 ✓
  [2/8] cc928331 ✓
  [3/8] a1b2c3d4 ✓
  [4/8] 74437e33 (no content)
  [5/8] f9e8d7c6 ✓
  [6/8] b5a4c3d2 ✓
  [7/8] 91827364 ✓
  [8/8] 55443322 (error)
Done: 6 generated, 2 skipped

DATE        TIME   UUID      PROJECT           SUMMARY
2026-06-15  10:12  2803b936  ~/Dev/api-service  Fixed seq scan by adding composite index on (user_id, created_at)
2026-06-14  16:30  cc928331  ~/Dev/api-service  Refactored auth middleware to use sliding-window token refresh
...
```

This calls `claude` once per uncached session — it takes a moment the first time, then every future `ccsm list` shows the cached summaries instantly at no cost.

## Step 2: Reuse the Cache

```bash
ccsm list

DATE        TIME   UUID      PROJECT           SUMMARY
2026-06-15  10:12  2803b936  ~/Dev/api-service  Fixed seq scan by adding composite index on (user_id, created_at)
...
```

No `--ai` flag needed — the cached summaries appear automatically.

## Step 3: Drill Into a Session

The one-liner tells you *what* was done. To understand *how* — the key steps and outcome — run:

```bash
ccsm summarize 2803b936

Session:  2803b936-7cb5-499f-804a-3804351f4b94
Project:  ~/Dev/api-service
Date:     2026-06-15 10:12
Turns:    31

**What was worked on:** Diagnosed slow migration caused by sequential scan on the
`orders` table despite an existing index on `user_id`.

**Key steps:**
- Ran EXPLAIN ANALYZE to confirm seq scan
- Identified that the query filtered on both `user_id` and `created_at`
- Added a composite index on (user_id, created_at)
- Verified the index was picked up and query time dropped from 4.2s to 18ms

**Outcome:** Migration completed. Composite index added in a follow-up PR.
```

`ccsm summarize` reads the first 20 turns and asks Claude for a structured breakdown — what was worked on, key steps, and outcome.

## When to Use Each

| Command | Use when |
|---------|----------|
| `ccsm list` | Scanning your recent history at a glance |
| `ccsm list --ai` | You want outcome-aware summaries, not just first messages |
| `ccsm summarize <uuid>` | You need to understand what was actually resolved before deciding whether to resume |

## Scenario 2 Takeaways

- `ccsm list --ai` is a one-time cache warm — run it once, benefit on every future `ccsm list`
- `ccsm summarize` is on-demand and not cached — use it for deep dives before resuming a session
- Summaries use `claude --no-session-persistence` so they don't pollute your session list

---

# Scenario 3: ⚙️ Scripting — Automating with JSON Output

## The Problem

You want to build a weekly report of what you worked on, integrate session data into another tool, or automate session selection. The tabular output from `ccsm list` is easy to read but hard to parse reliably.

## Step 1: Get JSON Output

```bash
ccsm list --json | jq .
```

Every command that produces a list supports `--json`:

```bash
ccsm list --json
ccsm search <term> --json
```

## Step 2: Inspect the Shape

```bash
ccsm list --json | jq '.[0]'

{
  "uuid": "2803b936-7cb5-499f-804a-3804351f4b94",
  "project_path": "/home/user/Dev/api-service",
  "first_time": "2026-06-15T10:12:03Z",
  "last_time": "2026-06-15T11:44:22Z",
  "first_msg": "debug the query planner for slow migration",
  "files": [
    "/home/user/Dev/api-service/db/migrations/0042_add_index.sql",
    "/home/user/Dev/api-service/internal/db/query.go"
  ],
  "msg_count": 31
}
```

The `files` array contains filesystem paths extracted from your messages — a fast hint at which files were discussed.

## Step 3: Common Recipes

**Sessions from a specific project today:**

```bash
ccsm list --json \
  | jq '[.[] | select(.project_path | contains("api-service")) | select(.first_time | startswith("2026-06-15"))]'
```

**Sessions that touched a specific file:**

```bash
ccsm list --json \
  | jq '[.[] | select(.files[]? | contains("migrations"))]'
```

**Weekly summary — session count per project:**

```bash
ccsm list --since 2026-06-09 --json \
  | jq 'group_by(.project_path) | map({project: .[0].project_path, sessions: length})'
```

**Search results with match context:**

```bash
ccsm search "composite index" --json | jq '.[] | {uuid: .uuid, turn: .turn, match: .match}'
```

The `search --json` output includes `match` (the snippet) and `turn` (which turn it was found on).

## Step 4: Build a Resume Picker Without fzf

If you don't have fzf, build a simple numbered picker with jq and shell:

```bash
ccsm list --json \
  | jq -r '.[] | "\(.uuid[0:8])  \(.first_msg[0:60])"' \
  | nl -ba
```

Pick a number, copy the UUID prefix, and resume:

```bash
claude --resume <uuid-from-above>
```

## Environment Variable Override

For testing or scripting against a different directory:

```bash
CCSM_SESSIONS_DIR=/path/to/sessions ccsm list --json
```

Useful for running `ccsm` against a backup of `~/.claude/projects/` or a fixture directory in CI.

## Scenario 3 Takeaways

- `--json` output is stable and safe to parse — use it for any automation
- `files` is a best-effort extraction from message text, not a filesystem scan — useful for filtering but not exhaustive
- `search --json` includes `match` and `turn` alongside all session fields
- `CCSM_SESSIONS_DIR` overrides the sessions directory for testing or scripting against non-default paths

---

## Overall Takeaways

All three scenarios build on the same principle: **session files are structured data, not opaque blobs**. `ccsm` makes that data navigable at every level of depth.

| Scenario | The question it answers |
|----------|------------------------|
| Daily navigation | *Which session was it, and is it the right one?* |
| AI summaries | *What was actually resolved — and is it worth resuming?* |
| Scripting | *What have I been working on across all projects and time?* |

The binary handles the data; `claude --resume` handles the continuity. `ccsm` is the bridge between the two.
