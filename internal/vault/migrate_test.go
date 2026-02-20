package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMigrateFrontmatterDryRun(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "plain.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write plain note: %v", err)
	}
	withFM := "---\nkind: task\n---\n\nbody"
	if err := os.WriteFile(filepath.Join(root, "with-fm.md"), []byte(withFM), 0o644); err != nil {
		t.Fatalf("write frontmatter note: %v", err)
	}

	result, err := MigrateFrontmatter(root, MigrateOptions{DryRun: true, Kind: "task"})
	if err != nil {
		t.Fatalf("MigrateFrontmatter dry-run error: %v", err)
	}
	if result.Migrated != 0 {
		t.Fatalf("expected 0 migrated in dry-run, got %d", result.Migrated)
	}
	if result.Skipped != 1 {
		t.Fatalf("expected 1 skipped, got %d", result.Skipped)
	}
	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files in report, got %d", len(result.Files))
	}

	content, err := os.ReadFile(filepath.Join(root, "plain.md"))
	if err != nil {
		t.Fatalf("read plain note after dry-run: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("dry-run should not modify content, got %q", string(content))
	}
}

func TestMigrateFrontmatterWritesHeader(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "plain.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write plain note: %v", err)
	}

	mtime := time.Date(2026, 2, 20, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	result, err := MigrateFrontmatter(root, MigrateOptions{Kind: "idea"})
	if err != nil {
		t.Fatalf("MigrateFrontmatter error: %v", err)
	}
	if result.Migrated != 1 || result.Skipped != 0 {
		t.Fatalf("unexpected migration counters: %+v", result)
	}
	if len(result.Files) != 1 || result.Files[0].Action != "migrated" {
		t.Fatalf("unexpected file actions: %+v", result.Files)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migrated note: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "kind: idea") {
		t.Fatalf("missing kind header: %q", text)
	}
	if !strings.Contains(text, "tags: []") {
		t.Fatalf("missing tags header: %q", text)
	}
	if !strings.Contains(text, "created_at: 2026-02-20T03:04:05Z") {
		t.Fatalf("missing created_at header: %q", text)
	}
	if !strings.HasSuffix(text, "\n\nhello") {
		t.Fatalf("expected original body at end, got %q", text)
	}
}
