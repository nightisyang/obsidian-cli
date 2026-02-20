package note

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func TestDailyPathAndReadCreate(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".obsidian"), 0o755); err != nil {
		t.Fatalf("mkdir .obsidian: %v", err)
	}
	cfg := `{"folder":"Journal","format":"YYYY/MM/DD"}`
	if err := os.WriteFile(filepath.Join(root, ".obsidian", "daily-notes.json"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write daily config: %v", err)
	}

	at := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)
	path, err := ResolveDailyPath(root, at)
	if err != nil {
		t.Fatalf("ResolveDailyPath error: %v", err)
	}
	if path != "Journal/2026/02/20.md" {
		t.Fatalf("unexpected daily path: %s", path)
	}

	n, resolved, err := DailyRead(root, at, true)
	if err != nil {
		t.Fatalf("DailyRead create error: %v", err)
	}
	if resolved != path || n.Path != path {
		t.Fatalf("unexpected daily read result: %s %+v", resolved, n)
	}
	if n.Frontmatter.CreatedAt == nil || n.Frontmatter.UpdatedAt == nil {
		t.Fatalf("expected timestamps on created daily note")
	}
}

func TestReadHeading(t *testing.T) {
	root := t.TempDir()
	raw := "# Top\n\n## Work\nline 1\nline 2\n## Home\nline 3\n"
	if err := os.WriteFile(filepath.Join(root, "note.md"), []byte(raw), 0o644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	section, err := ReadHeading(root, "note.md", "Work")
	if err != nil {
		t.Fatalf("ReadHeading error: %v", err)
	}
	if section.Level != 2 || section.Content != "line 1\nline 2" {
		t.Fatalf("unexpected section: %+v", section)
	}
}

func TestGetAndSetBlock(t *testing.T) {
	root := t.TempDir()
	raw := "paragraph text ^para1\n\n- [ ] Task item\n^task1\n"
	if err := os.WriteFile(filepath.Join(root, "block.md"), []byte(raw), 0o644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	inline, err := GetBlock(root, "block.md", "para1")
	if err != nil {
		t.Fatalf("GetBlock inline error: %v", err)
	}
	if inline.Content != "paragraph text" {
		t.Fatalf("unexpected inline block: %+v", inline)
	}

	anchor, err := GetBlock(root, "block.md", "task1")
	if err != nil {
		t.Fatalf("GetBlock anchor error: %v", err)
	}
	if anchor.Content != "- [ ] Task item" {
		t.Fatalf("unexpected anchor block: %+v", anchor)
	}

	updated, err := SetBlock(root, "block.md", "para1", "new text")
	if err != nil {
		t.Fatalf("SetBlock error: %v", err)
	}
	if updated.Content != "new text" {
		t.Fatalf("unexpected updated block: %+v", updated)
	}

	if _, err := GetBlock(root, "block.md", "AB"); err == nil {
		t.Fatalf("expected invalid block id error")
	} else {
		appErr, ok := err.(*errs.AppError)
		if !ok || appErr.Code != errs.ExitValidation {
			t.Fatalf("expected validation app error, got %T (%v)", err, err)
		}
	}
}
