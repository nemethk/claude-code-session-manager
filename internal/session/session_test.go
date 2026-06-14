package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExtractText_string(t *testing.T) {
	raw, _ := json.Marshal("hello world")
	text, ok := extractText(raw)
	if !ok || text != "hello world" {
		t.Fatalf("want %q true, got %q %v", "hello world", text, ok)
	}
}

func TestExtractText_textBlock(t *testing.T) {
	raw := json.RawMessage(`[{"type":"text","text":"from block"}]`)
	text, ok := extractText(raw)
	if !ok || text != "from block" {
		t.Fatalf("want %q true, got %q %v", "from block", text, ok)
	}
}

func TestExtractText_toolResult(t *testing.T) {
	raw := json.RawMessage(`[{"type":"tool_result","tool_use_id":"x","content":"result"}]`)
	_, ok := extractText(raw)
	if ok {
		t.Fatal("tool_result should not extract as typed text")
	}
}

func TestExtractText_emptyString(t *testing.T) {
	raw, _ := json.Marshal("")
	_, ok := extractText(raw)
	if ok {
		t.Fatal("empty string should not be ok")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("no truncation: want hello, got %q", got)
	}
	if got := truncate("hello world", 8); got != "hello wo..." {
		t.Errorf("truncated: want %q, got %q", "hello wo...", got)
	}
}

func makeSession(t *testing.T, dir, slug, uuid, content string) {
	t.Helper()
	projectDir := filepath.Join(dir, slug)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, uuid+".jsonl"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadFrom_basic(t *testing.T) {
	dir := t.TempDir()
	makeSession(t, dir, "-home-user-test", "abc12345-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user/test","message":{"role":"user","content":"hello ccsm"}}`+"\n",
	)

	sessions, err := LoadFrom(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("want 1 session, got %d", len(sessions))
	}
	s := sessions[0]
	if s.UUID != "abc12345-0000-0000-0000-000000000000" {
		t.Errorf("UUID: got %s", s.UUID)
	}
	if s.ProjectPath != "/home/user/test" {
		t.Errorf("ProjectPath: got %s", s.ProjectPath)
	}
	if s.FirstMsg != "hello ccsm" {
		t.Errorf("FirstMsg: got %q", s.FirstMsg)
	}
	want := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	if !s.FirstTime.Equal(want) {
		t.Errorf("FirstTime: want %v, got %v", want, s.FirstTime)
	}
	if s.MsgCount != 1 {
		t.Errorf("MsgCount: want 1, got %d", s.MsgCount)
	}
}

func TestLoadFrom_skipsMetaTypes(t *testing.T) {
	dir := t.TempDir()
	content := `{"type":"mode","mode":"normal"}` + "\n" +
		`{"type":"permission-mode","permissionMode":"default"}` + "\n" +
		`{"type":"file-history-snapshot","timestamp":"2026-06-01T09:59:59Z"}` + "\n" +
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"real message"}}` + "\n"
	makeSession(t, dir, "-home-user", "sess-0000-0000-0000-000000000001", content)

	sessions, err := LoadFrom(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].FirstMsg != "real message" {
		t.Fatalf("unexpected: %+v", sessions)
	}
}

func TestLoadFrom_skipsToolResults(t *testing.T) {
	dir := t.TempDir()
	// content is a real JSON array (tool result), not a string
	content := `{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"x","content":"lots of output"}]}}` + "\n" +
		`{"type":"user","timestamp":"2026-06-01T10:00:01Z","cwd":"/home/user","message":{"role":"user","content":"actual user question"}}` + "\n"
	makeSession(t, dir, "-home-user", "sess-0000-0000-0000-000000000002", content)

	sessions, err := LoadFrom(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 {
		t.Fatalf("want 1 session, got %d", len(sessions))
	}
	if sessions[0].FirstMsg != "actual user question" {
		t.Errorf("FirstMsg should skip tool result, got %q", sessions[0].FirstMsg)
	}
}

func TestLoadFrom_empty(t *testing.T) {
	dir := t.TempDir()
	sessions, err := LoadFrom(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("want 0 sessions, got %d", len(sessions))
	}
}

func TestLoadFrom_lastTime(t *testing.T) {
	dir := t.TempDir()
	content := `{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"hello"}}` + "\n" +
		`{"type":"assistant","timestamp":"2026-06-01T10:05:00Z"}` + "\n"
	makeSession(t, dir, "-home-user", "sess-0000-0000-0000-000000000003", content)

	sessions, err := LoadFrom(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 6, 1, 10, 5, 0, 0, time.UTC)
	if !sessions[0].LastTime.Equal(want) {
		t.Errorf("LastTime: want %v, got %v", want, sessions[0].LastTime)
	}
}

func TestFindByPrefixIn(t *testing.T) {
	dir := t.TempDir()
	makeSession(t, dir, "-home-user", "aabbccdd-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"find me"}}`,
	)

	s, err := FindByPrefixIn(dir, "aabbccdd")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil || s.FirstMsg != "find me" {
		t.Fatalf("unexpected: %+v", s)
	}

	s, err = FindByPrefixIn(dir, "xxxxxxxx")
	if err != nil || s != nil {
		t.Fatalf("should not find: %+v, %v", s, err)
	}
}

func TestMessagesFrom(t *testing.T) {
	dir := t.TempDir()
	content := `{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"turn one"}}` + "\n" +
		`{"type":"assistant","timestamp":"2026-06-01T10:00:01Z"}` + "\n" +
		`{"type":"user","timestamp":"2026-06-01T10:00:02Z","cwd":"/home/user","message":{"role":"user","content":"turn two"}}` + "\n"
	makeSession(t, dir, "-home-user", "msgtest0-0000-0000-0000-000000000000", content)

	msgs, err := MessagesFrom(dir, "msgtest0", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 || msgs[0] != "turn one" || msgs[1] != "turn two" {
		t.Fatalf("unexpected messages: %v", msgs)
	}
}

func mustMarshal(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
