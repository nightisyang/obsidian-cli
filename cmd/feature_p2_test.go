package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/search"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

type testEnvelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	resetRootOptsForTest()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = outWriter
	os.Stderr = errWriter
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	command := newRootCmd()
	command.SetArgs(args)
	runErr := command.Execute()

	_ = outWriter.Close()
	_ = errWriter.Close()
	stdout, _ := io.ReadAll(outReader)
	stderr, _ := io.ReadAll(errReader)
	return string(stdout), string(stderr), runErr
}

func resetRootOptsForTest() {
	rootOpts.vault = ""
	rootOpts.config = ""
	rootOpts.mode = ""
	rootOpts.json = false
	rootOpts.quiet = false
	rootOpts.noteSizeMaxBytes = 131072
	rootOpts.noOrphanNotes = false
}

func parseEnvelope(t *testing.T, raw string) testEnvelope {
	t.Helper()
	var env testEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("decode envelope: %v; raw=%q", err, raw)
	}
	if !env.OK {
		t.Fatalf("expected ok envelope, got error: %+v", env.Error)
	}
	return env
}

func TestNoteFindCommandTextAndJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	one := `---
kind: task
tags: [alpha, beta]
status: open
topic: infra
created_at: 2026-02-10T00:00:00Z
---

hello`
	two := `---
kind: task
tags: [alpha]
status: done
topic: infra
created_at: 2026-01-10T00:00:00Z
---

hello`
	if err := os.WriteFile(filepath.Join(root, "tasks", "one.md"), []byte(one), 0o644); err != nil {
		t.Fatalf("write one: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tasks", "two.md"), []byte(two), 0o644); err != nil {
		t.Fatalf("write two: %v", err)
	}

	stdout, stderr, err := runCLI(
		t,
		"--vault", root,
		"note", "find",
		"--kind", "task",
		"--tag", "alpha",
		"--tag", "beta",
		"--status", "open",
		"--topic", "infra",
		"--since", "2026-02-01",
	)
	if err != nil {
		t.Fatalf("note find text error: %v (stderr=%q)", err, stderr)
	}
	expected := "tasks/one.md: kind=task tags=[alpha,beta] status=open\n"
	if stdout != expected {
		t.Fatalf("unexpected note find text output:\nwant=%q\ngot=%q", expected, stdout)
	}

	stdout, stderr, err = runCLI(
		t,
		"--vault", root,
		"--json",
		"note", "find",
		"--kind", "task",
		"--tag", "alpha",
	)
	if err != nil {
		t.Fatalf("note find json error: %v (stderr=%q)", err, stderr)
	}
	env := parseEnvelope(t, stdout)
	var records []map[string]any
	if err := json.Unmarshal(env.Data, &records); err != nil {
		t.Fatalf("decode records: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if _, ok := records[0]["body"]; ok {
		t.Fatalf("note find json should not include body")
	}
}

func TestVaultMigrateCommandDryRunAndJSON(t *testing.T) {
	root := t.TempDir()
	plainPath := filepath.Join(root, "plain.md")
	if err := os.WriteFile(plainPath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write plain: %v", err)
	}
	withFM := "---\nkind: task\n---\n\nbody"
	if err := os.WriteFile(filepath.Join(root, "with-fm.md"), []byte(withFM), 0o644); err != nil {
		t.Fatalf("write with-fm: %v", err)
	}

	mtime := time.Date(2026, 2, 20, 8, 9, 10, 0, time.UTC)
	if err := os.Chtimes(plainPath, mtime, mtime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	stdout, stderr, err := runCLI(t, "--vault", root, "vault", "migrate", "--dry-run", "--kind", "memo")
	if err != nil {
		t.Fatalf("vault migrate dry-run error: %v (stderr=%q)", err, stderr)
	}
	if !strings.Contains(stdout, "would migrate: plain.md\n") {
		t.Fatalf("expected dry-run migration line, got %q", stdout)
	}
	if !strings.Contains(stdout, "Migrated 0 notes, 1 already had frontmatter\n") {
		t.Fatalf("expected dry-run summary, got %q", stdout)
	}

	stdout, stderr, err = runCLI(t, "--vault", root, "--json", "vault", "migrate", "--kind", "memo")
	if err != nil {
		t.Fatalf("vault migrate json error: %v (stderr=%q)", err, stderr)
	}
	env := parseEnvelope(t, stdout)
	var result vault.MigrationResult
	if err := json.Unmarshal(env.Data, &result); err != nil {
		t.Fatalf("decode migration result: %v", err)
	}
	if result.Migrated != 1 || result.Skipped != 1 {
		t.Fatalf("unexpected migration summary: %+v", result)
	}

	content, err := os.ReadFile(plainPath)
	if err != nil {
		t.Fatalf("read migrated file: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "kind: memo") || !strings.Contains(text, "tags: []") {
		t.Fatalf("missing migrated frontmatter fields: %q", text)
	}
}

func TestMaxCharsOnNoteGetAndSearch(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "long.md"), []byte("0123456789"), 0o644); err != nil {
		t.Fatalf("write long note: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "search.md"), []byte("prefix hello world suffix"), 0o644); err != nil {
		t.Fatalf("write search note: %v", err)
	}

	stdout, stderr, err := runCLI(t, "--vault", root, "note", "get", "long.md", "--max-chars", "5")
	if err != nil {
		t.Fatalf("note get max-chars error: %v (stderr=%q)", err, stderr)
	}
	expectedMessage := "[TRUNCATED at 5 chars \u2014 use --heading to read a specific section]"
	if !strings.Contains(stdout, expectedMessage) {
		t.Fatalf("expected truncation message in output, got %q", stdout)
	}
	if !strings.HasPrefix(stdout, "01234\n\n") {
		t.Fatalf("expected body to be truncated to first 5 chars, got %q", stdout)
	}

	stdout, stderr, err = runCLI(t, "--vault", root, "--json", "note", "get", "long.md", "--max-chars", "5")
	if err != nil {
		t.Fatalf("note get json max-chars error: %v (stderr=%q)", err, stderr)
	}
	env := parseEnvelope(t, stdout)
	var gotNote note.Note
	if err := json.Unmarshal(env.Data, &gotNote); err != nil {
		t.Fatalf("decode note json: %v", err)
	}
	if !strings.Contains(gotNote.Body, expectedMessage) {
		t.Fatalf("expected truncation marker in json body, got %q", gotNote.Body)
	}

	stdout, stderr, err = runCLI(t, "--vault", root, "search", "hello", "--context", "120", "--max-chars", "6")
	if err != nil {
		t.Fatalf("search max-chars error: %v (stderr=%q)", err, stderr)
	}
	line := strings.TrimSpace(stdout)
	parts := strings.SplitN(line, "\t", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected search output format: %q", stdout)
	}
	if utf8.RuneCountInString(parts[1]) > 6 {
		t.Fatalf("expected text search snippet <= 6 chars, got %q", parts[1])
	}

	stdout, stderr, err = runCLI(t, "--vault", root, "--json", "search", "hello", "--context", "120", "--max-chars", "6")
	if err != nil {
		t.Fatalf("search json max-chars error: %v (stderr=%q)", err, stderr)
	}
	env = parseEnvelope(t, stdout)
	var results []search.SearchResult
	if err := json.Unmarshal(env.Data, &results); err != nil {
		t.Fatalf("decode search json: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected at least one search result")
	}
	if utf8.RuneCountInString(results[0].Snippet) > 6 {
		t.Fatalf("expected json search snippet <= 6 chars, got %q", results[0].Snippet)
	}
}
