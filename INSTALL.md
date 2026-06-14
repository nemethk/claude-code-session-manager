# 📦 Installation Guide

Step-by-step instructions for installing and using `ccsm`.

---

## ✓ Prerequisites

- **🍎 macOS** or **🐧 Linux** (Windows not supported yet)
- One of: `curl`, `brew`, or `go` (for installation)
- **[Claude Code](https://claude.ai/claude-code)** — required only for `ccsm list --ai` and `ccsm summarize`
- **[fzf](https://github.com/junegunn/fzf)** — optional, for the fuzzy-resume alias

---

## Step 1: Install the Binary

Choose one method:

### Option A: curl (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/nemethk/claude-code-session-manager/main/scripts/install.sh | bash
```

The script detects your OS and architecture, downloads the correct binary from the latest release, and places it in `/usr/local/bin`.

### Option B: Homebrew

```bash
brew install nemethk/tap/ccsm
```

### Option C: go install

```bash
GOBIN=/usr/local/bin go install github.com/nemethk/claude-code-session-manager@latest
```

Requires Go 1.26 or later.

---

## Step 2: Verify the Installation

```bash
ccsm --version
```

You should see the version number. If you get `command not found`, check that `/usr/local/bin` is in your `PATH`.

---

## Step 3: List Your Sessions

```bash
ccsm list
```

`ccsm` reads `~/.claude/projects/` automatically — no configuration needed.

---

## Step 4: Resume a Session (optional shortcut)

If you have `fzf` installed, add this alias to your shell config (`~/.zshrc` or `~/.bashrc`):

```bash
alias cr='ccsm list | fzf | awk '"'"'{print $3}'"'"' | xargs claude --resume'
```

Reload your shell:

```bash
source ~/.zshrc   # or ~/.bashrc
```

Now `cr` fuzzy-picks any session and resumes it in one step.

---

## Step 5: Install the `/sessions` Skill (optional)

The `/sessions` skill adds natural language session search inside any Claude Code conversation.

```bash
cp skill/sessions.md ~/.claude/skills/sessions.md
```

Or download directly:

```bash
curl -fsSL https://raw.githubusercontent.com/nemethk/claude-code-session-manager/main/skill/sessions.md \
  -o ~/.claude/skills/sessions.md
```

Then use it inside Claude Code:

```
/sessions find postgres migration
/sessions resume 2803b936
```

---

## 🚀 Upgrading

### Option 1: Re-run the install script (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/nemethk/claude-code-session-manager/main/scripts/install.sh | bash
```

### Option 2: Homebrew

```bash
brew upgrade ccsm
```

### Option 3: Go

```bash
GOBIN=/usr/local/bin go install github.com/nemethk/claude-code-session-manager@latest
```

### ✓ Verify upgrade

```bash
ccsm --version
```

---

## 🗑️ Uninstalling

```bash
sudo rm /usr/local/bin/ccsm
```

Remove the summary cache (optional):

```bash
rm -rf ~/.cache/ccsm
```

Remove the `/sessions` skill (optional):

```bash
rm ~/.claude/skills/sessions.md
```

---

## 🔧 Troubleshooting

**`command not found: ccsm`**
- Check that `/usr/local/bin` is in your `PATH`

**`ccsm list` shows no sessions**
- Confirm Claude Code has been used at least once and has sessions in `~/.claude/projects/`
- Override the sessions directory with `CCSM_SESSIONS_DIR` if your sessions are elsewhere

**`ccsm list --ai` produces no AI summaries**
- Confirm `claude --version` works — Claude Code must be installed and authenticated
- Run `ccsm summarize <uuid>` for a single session to test the AI integration directly

**`ccsm list --ai` created extra sessions in `~/.claude/projects/`**
- This can happen with older versions. Upgrade to the latest release — `--no-session-persistence` is used by default since v0.2.0
- Filter out single-turn noise sessions with `ccsm list --min-turns 2`
