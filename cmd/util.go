package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nemethk/claude-code-session-manager/internal/ai"
	"github.com/nemethk/claude-code-session-manager/internal/cache"
	"github.com/nemethk/claude-code-session-manager/internal/session"
)

func shortenPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	if strings.HasPrefix(p, home) {
		return "~" + p[len(home):]
	}
	return p
}

type summarizeStatus int

const (
	summarizeGenerated summarizeStatus = iota
	summarizeNoContent
	summarizeError
)

// ensureSummary generates and caches an AI summary if one doesn't exist yet.
// Returns the outcome so callers can report progress accurately.
func ensureSummary(s session.Session) summarizeStatus {
	msgs, err := session.Messages(s.UUID, 8)
	if err != nil {
		return summarizeError
	}
	var clean []string
	for _, m := range msgs {
		if !strings.HasPrefix(strings.TrimSpace(m), "<") {
			clean = append(clean, m)
		}
	}
	if len(clean) == 0 {
		return summarizeNoContent
	}
	summary, err := ai.Summarize(clean)
	if err != nil {
		return summarizeError
	}
	cache.Set(s.UUID, summary)
	return summarizeGenerated
}

var absPathRe = regexp.MustCompile(`/[^\s,;"'<>()\[\]{}\\]+`)

// sessionSummary builds the display string for the SUMMARY column.
// Returns cached AI summary when available, falls back to heuristic.
func sessionSummary(s session.Session) string {
	if summary, ok := cache.Get(s.UUID); ok {
		return summary
	}
	msg := cleanMsg(s.FirstMsg)

	seen := map[string]bool{}
	var names []string
	for _, f := range s.Files {
		b := filepath.Base(f)
		if b == "." || b == "/" || seen[b] {
			continue
		}
		if strings.Contains(strings.ToLower(msg), strings.ToLower(b)) {
			continue
		}
		seen[b] = true
		names = append(names, b)
		if len(names) == 2 {
			break
		}
	}
	if len(names) > 0 {
		msg += " [" + strings.Join(names, ", ") + "]"
	}
	if len(msg) > 90 {
		msg = msg[:87] + "..."
	}
	return msg
}

var focusPrefixes = []string{"focus on ", "focus "}

func cleanMsg(msg string) string {
	msg = strings.TrimSpace(msg)
	lower := strings.ToLower(msg)
	for _, p := range focusPrefixes {
		if strings.HasPrefix(lower, p) {
			msg = strings.TrimSpace(msg[len(p):])
			break
		}
	}
	msg = absPathRe.ReplaceAllStringFunc(msg, func(p string) string {
		parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
		if len(parts) <= 2 {
			return p
		}
		return strings.Join(parts[len(parts)-2:], "/")
	})
	return strings.TrimSpace(msg)
}
