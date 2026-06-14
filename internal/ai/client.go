package ai

import (
	"fmt"
	"os/exec"
	"strings"
)

// Summarize sends user messages to the claude CLI and returns a one-line summary.
func Summarize(msgs []string) (string, error) {
	return run(buildPrompt(msgs))
}

// Detail sends user messages to the claude CLI and returns a detailed session summary.
func Detail(msgs []string) (string, error) {
	return run(buildDetailPrompt(msgs))
}

func run(prompt string) (string, error) {
	out, err := exec.Command("claude", "--no-session-persistence", "-p", prompt).Output()
	if err != nil {
		return "", fmt.Errorf("claude CLI error: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func buildPrompt(msgs []string) string {
	var sb strings.Builder
	sb.WriteString("Summarize this Claude Code developer session in one short sentence (max 10 words). Be specific about what was worked on. Output only the sentence, no preamble.\n\nUser messages:\n")
	for i, m := range msgs {
		if len(m) > 150 {
			m = m[:150] + "..."
		}
		fmt.Fprintf(&sb, "%d. %s\n", i+1, m)
	}
	return sb.String()
}

func buildDetailPrompt(msgs []string) string {
	var sb strings.Builder
	sb.WriteString(`Summarize this Claude Code developer session. Structure your response as:

**What was worked on:** (1-2 sentences)

**Key steps:**
- (bullet points of the main things done)

**Outcome:** (1 sentence on result or current state)

Be specific and technical. No preamble.

User messages:
`)
	for i, m := range msgs {
		if len(m) > 300 {
			m = m[:300] + "..."
		}
		fmt.Fprintf(&sb, "%d. %s\n", i+1, m)
	}
	return sb.String()
}
