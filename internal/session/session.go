package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Session holds extracted metadata for one Claude Code session.
type Session struct {
	UUID        string    `json:"uuid"`
	ProjectPath string    `json:"project_path"`
	FilePath    string    `json:"-"`
	FirstTime   time.Time `json:"first_time"`
	LastTime    time.Time `json:"last_time"`
	FirstMsg    string    `json:"first_msg"`
	Files       []string  `json:"files"`
	MsgCount    int       `json:"msg_count"`
}

type rawEntry struct {
	Type      string   `json:"type"`
	IsMeta    bool     `json:"isMeta"`
	Timestamp string   `json:"timestamp"`
	CWD       string   `json:"cwd"`
	Message   *rawMsg  `json:"message"`
}

type rawMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// Dir returns the sessions directory. Override with CCSM_SESSIONS_DIR for tests.
func Dir() string {
	if d := os.Getenv("CCSM_SESSIONS_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// Load reads all sessions from the default sessions directory.
func Load() ([]Session, error) {
	return LoadFrom(Dir())
}

// LoadFrom reads all sessions from dir.
func LoadFrom(dir string) ([]Session, error) {
	projectDirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0)
	for _, pd := range projectDirs {
		if !pd.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(dir, pd.Name()))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
				continue
			}
			uuid := strings.TrimSuffix(e.Name(), ".jsonl")
			path := filepath.Join(dir, pd.Name(), e.Name())
			s, err := parseSession(uuid, path)
			if err != nil || s == nil {
				continue
			}
			s.FilePath = path
			sessions = append(sessions, *s)
		}
	}
	return sessions, nil
}

// FindByPrefix returns the first session whose UUID starts with the given prefix.
func FindByPrefix(prefix string) (*Session, error) {
	return FindByPrefixIn(Dir(), prefix)
}

// FindByPrefixIn searches for a session UUID prefix within dir.
func FindByPrefixIn(dir, prefix string) (*Session, error) {
	sessions, err := LoadFrom(dir)
	if err != nil {
		return nil, err
	}
	for i := range sessions {
		if strings.HasPrefix(sessions[i].UUID, prefix) {
			return &sessions[i], nil
		}
	}
	return nil, nil
}

// MatchText returns a snippet of the first user message in path that contains term
// (case-insensitive), along with the turn number. Returns "", 0 if no match.
func MatchText(path, term string) (snippet string, turn int) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0
	}
	defer f.Close()

	lower := strings.ToLower(term)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	msgNum := 0
	for scanner.Scan() {
		var e rawEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		if skipType(e.Type) || e.IsMeta || e.Type != "user" || e.Message == nil {
			continue
		}
		text, ok := extractText(e.Message.Content)
		if !ok {
			continue
		}
		msgNum++
		idx := strings.Index(strings.ToLower(text), lower)
		if idx < 0 {
			continue
		}
		// extract ~60 chars around the match
		start := idx - 30
		if start < 0 {
			start = 0
		}
		end := idx + len(term) + 30
		if end > len(text) {
			end = len(text)
		}
		snip := strings.ReplaceAll(text[start:end], "\n", " ")
		if start > 0 {
			snip = "..." + snip
		}
		if end < len(text) {
			snip = snip + "..."
		}
		return snip, msgNum
	}
	return "", 0
}

// Messages returns the first n typed user messages from the session matching the UUID prefix.
func Messages(prefix string, n int) ([]string, error) {
	return MessagesFrom(Dir(), prefix, n)
}

// MessagesFrom searches dir for a session matching the UUID prefix and returns first n user messages.
func MessagesFrom(dir, prefix string, n int) ([]string, error) {
	projectDirs, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, pd := range projectDirs {
		if !pd.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(dir, pd.Name()))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
				continue
			}
			uuid := strings.TrimSuffix(e.Name(), ".jsonl")
			if !strings.HasPrefix(uuid, prefix) {
				continue
			}
			path := filepath.Join(dir, pd.Name(), e.Name())
			return readMessages(path, n)
		}
	}
	return nil, nil
}

func parseSession(uuid, path string) (*Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := &Session{UUID: uuid}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	seenPaths := map[string]bool{}

	for scanner.Scan() {
		var e rawEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}

		if (e.Type == "user" || e.Type == "assistant") && !e.IsMeta {
			t, _ := time.Parse(time.RFC3339Nano, e.Timestamp)
			if !t.IsZero() {
				if s.FirstTime.IsZero() {
					s.FirstTime = t
				}
				s.LastTime = t
			}
		}

		if skipType(e.Type) || e.IsMeta || e.Type != "user" || e.Message == nil {
			continue
		}
		if s.ProjectPath == "" && e.CWD != "" {
			s.ProjectPath = e.CWD
		}
		if text, ok := extractText(e.Message.Content); ok {
			if s.FirstMsg == "" && !isInjected(text) {
				s.FirstMsg = truncate(text, 200)
			}
			for _, p := range extractPaths(text) {
				if !seenPaths[p] {
					seenPaths[p] = true
					s.Files = append(s.Files, p)
				}
			}
			s.MsgCount++
		}
	}

	if s.ProjectPath == "" && s.FirstTime.IsZero() {
		return nil, nil
	}
	return s, scanner.Err()
}

func readMessages(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var msgs []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	for scanner.Scan() && len(msgs) < n {
		var e rawEntry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		if skipType(e.Type) || e.IsMeta || e.Type != "user" || e.Message == nil {
			continue
		}
		if text, ok := extractText(e.Message.Content); ok {
			msgs = append(msgs, text)
		}
	}
	return msgs, scanner.Err()
}

var skipTypes = map[string]bool{
	"mode": true, "permission-mode": true,
	"file-history-snapshot": true, "attachment": true,
}

func skipType(t string) bool { return skipTypes[t] }

// isInjected returns true for Claude Code system-injected content (caveat blocks,
// slash command wrappers) that should not be treated as real user messages.
func isInjected(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(text), "<")
}

var pathRe = regexp.MustCompile(`(?:^|[\s,;:=])((~|\.\.?)?/[^\s,;"'<>()\[\]{}\\]+)`)

func extractPaths(text string) []string {
	matches := pathRe.FindAllStringSubmatch(text, -1)
	var paths []string
	for _, m := range matches {
		p := strings.TrimRight(m[1], ".,;:)")
		if len(p) > 2 && isFilesystemPath(p) {
			paths = append(paths, filepath.Clean(p))
		}
	}
	return paths
}

// isFilesystemPath rejects URL-like paths where the first component contains a dot
// (e.g. /github.com/..., /img.shields.io/...) while keeping real paths like /home/user/...
func isFilesystemPath(p string) bool {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") {
		return true
	}
	// strip all leading slashes (handles // from https:// matches)
	first := strings.SplitN(strings.TrimLeft(p, "/"), "/", 2)[0]
	return first != "" && !strings.Contains(first, ".")
}

func extractText(raw json.RawMessage) (string, bool) {
	var s string
	if json.Unmarshal(raw, &s) == nil && s != "" {
		return s, true
	}
	var blocks []json.RawMessage
	if json.Unmarshal(raw, &blocks) != nil || len(blocks) == 0 {
		return "", false
	}
	var first struct {
		Type string `json:"type"`
	}
	if json.Unmarshal(blocks[0], &first) != nil || first.Type == "tool_result" {
		return "", false
	}
	for _, b := range blocks {
		var tb struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if json.Unmarshal(b, &tb) == nil && tb.Type == "text" && tb.Text != "" {
			return tb.Text, true
		}
	}
	return "", false
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
