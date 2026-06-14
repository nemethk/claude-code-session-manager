Manage Claude Code sessions using the `ccsm` CLI.

## Setup check

If `ccsm` is not installed, tell the user to run:
```
cd ~/DevOps/Git/claude-code-config/_dev/ccsm && make install
```

## Commands available

- `ccsm list --json` — all sessions as a JSON array
- `ccsm search <term> --json` — sessions matching term (first message or project path)
- `ccsm show <uuid-prefix>` — first 5 user turns from a session (use `--turns N` for more)

## Behavior by input

**No argument or "list"**
Run `ccsm list --json`. Display a numbered table:
```
#   Date        Project                     First message
1.  2026-06-14  ~/DevOps/Git/my-project     fix the auth bug
```
Include the UUID after the row (e.g., `UUID: 2803b936-...`) so the user can copy it.

**Search / find query** (e.g., `/sessions find postgres` or `/sessions kubernetes`)
Run `ccsm list --json`. Scan all sessions, filter and rank by relevance to the query
(match against `first_msg` and `project_path`). Show the top matches with a one-line
explanation of why each is relevant.
For deeper inspection of a candidate, run `ccsm show <uuid>` and summarize the turns.

**`show <uuid-or-number>`**
If a number from a previous list, resolve it to the UUID. Run `ccsm show <uuid>`.
Display the turns so the user can decide whether to resume.

**`resume <uuid-or-number>`**
Resolve the UUID if needed. Print the exact command for the user to run themselves:
```
claude --resume <uuid>
```
Claude cannot invoke `claude --resume` — the user must run it in their terminal.

## Output style

- Keep it compact: one table row per session, not paragraphs
- Dates in YYYY-MM-DD format
- Shorten project paths with `~` for home directory
- If the session list is large (>20), ask the user to narrow with `--project` or `--since`
