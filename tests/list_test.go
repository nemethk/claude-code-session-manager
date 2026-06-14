package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func sessionsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func addSession(t *testing.T, dir, slug, uuid, content string) {
	t.Helper()
	pd := filepath.Join(dir, slug)
	if err := os.MkdirAll(pd, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pd, uuid+".jsonl"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "CCSM_SESSIONS_DIR="+dir)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestListEmpty(t *testing.T) {
	dir := sessionsDir(t)
	out, err := run(t, dir, "list")
	if err != nil {
		t.Fatalf("ccsm list: %v\n%s", err, out)
	}
	if !strings.Contains(out, "DATE") {
		t.Errorf("expected header, got: %s", out)
	}
}

func TestListShowsSession(t *testing.T) {
	dir := sessionsDir(t)
	addSession(t, dir, "-home-user-test", "aaaaaaaa-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user/test","message":{"role":"user","content":"hello ccsm"}}`+"\n",
	)

	out, err := run(t, dir, "list")
	if err != nil {
		t.Fatalf("ccsm list: %v\n%s", err, out)
	}
	if !strings.Contains(out, "hello ccsm") {
		t.Errorf("expected first message in output, got:\n%s", out)
	}
	if !strings.Contains(out, "aaaaaaaa-0000-0000-0000-000000000000") {
		t.Errorf("expected UUID in output, got:\n%s", out)
	}
}

func TestListJSON(t *testing.T) {
	dir := sessionsDir(t)
	addSession(t, dir, "-home-user", "bbbbbbbb-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"json test"}}`+"\n",
	)

	out, err := run(t, dir, "list", "--json")
	if err != nil {
		t.Fatalf("ccsm list --json: %v\n%s", err, out)
	}

	var sessions []map[string]interface{}
	if err := json.Unmarshal([]byte(out), &sessions); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(sessions) != 1 {
		t.Fatalf("want 1 session, got %d", len(sessions))
	}
	if sessions[0]["first_msg"] != "json test" {
		t.Errorf("first_msg: got %v", sessions[0]["first_msg"])
	}
}

func TestListFilterProject(t *testing.T) {
	dir := sessionsDir(t)
	addSession(t, dir, "-home-user-alpha", "cccccccc-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user/alpha","message":{"role":"user","content":"alpha session"}}`+"\n",
	)
	addSession(t, dir, "-home-user-beta", "dddddddd-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user/beta","message":{"role":"user","content":"beta session"}}`+"\n",
	)

	out, err := run(t, dir, "list", "--project", "alpha")
	if err != nil {
		t.Fatalf("ccsm list --project alpha: %v\n%s", err, out)
	}
	if !strings.Contains(out, "alpha session") {
		t.Errorf("expected alpha session, got:\n%s", out)
	}
	if strings.Contains(out, "beta session") {
		t.Errorf("beta session should be filtered out, got:\n%s", out)
	}
}

func TestSearch(t *testing.T) {
	dir := sessionsDir(t)
	addSession(t, dir, "-home-user", "eeeeeeee-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"debugging postgres connection"}}`+"\n",
	)
	addSession(t, dir, "-home-user", "ffffffff-0000-0000-0000-000000000000",
		`{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"kubernetes ingress setup"}}`+"\n",
	)

	out, err := run(t, dir, "search", "postgres")
	if err != nil {
		t.Fatalf("ccsm search: %v\n%s", err, out)
	}
	if !strings.Contains(out, "postgres") {
		t.Errorf("expected postgres result, got:\n%s", out)
	}
	if strings.Contains(out, "kubernetes") {
		t.Errorf("kubernetes should not appear in postgres search, got:\n%s", out)
	}
}

func TestShow(t *testing.T) {
	dir := sessionsDir(t)
	content := `{"type":"user","timestamp":"2026-06-01T10:00:00Z","cwd":"/home/user","message":{"role":"user","content":"first turn"}}` + "\n" +
		`{"type":"assistant","timestamp":"2026-06-01T10:00:01Z"}` + "\n" +
		`{"type":"user","timestamp":"2026-06-01T10:00:02Z","cwd":"/home/user","message":{"role":"user","content":"second turn"}}` + "\n"
	addSession(t, dir, "-home-user", "showtest-0000-0000-0000-000000000000", content)

	out, err := run(t, dir, "show", "showtest")
	if err != nil {
		t.Fatalf("ccsm show: %v\n%s", err, out)
	}
	if !strings.Contains(out, "first turn") || !strings.Contains(out, "second turn") {
		t.Errorf("expected both turns, got:\n%s", out)
	}
	if !strings.Contains(out, "Turn 1") || !strings.Contains(out, "Turn 2") {
		t.Errorf("expected turn labels, got:\n%s", out)
	}
}

func TestShowNotFound(t *testing.T) {
	dir := sessionsDir(t)
	_, err := run(t, dir, "show", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing session")
	}
}
