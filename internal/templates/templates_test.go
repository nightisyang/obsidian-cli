package templates

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/vault"
)

func TestListAndReadTemplate(t *testing.T) {
	root := t.TempDir()
	templateDir := filepath.Join(root, ".obsidian", "templates")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	content := "Date: {{date}}\nTime: {{time}}\nTitle: {{title}}"
	if err := os.WriteFile(filepath.Join(templateDir, "Travel.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	cfg := vault.DefaultConfig()
	list, err := List(root, cfg)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Travel" {
		t.Fatalf("unexpected templates: %+v", list)
	}

	fixed := time.Date(2026, 2, 20, 3, 4, 0, 0, time.UTC)
	raw, err := Read(root, cfg, "Travel", "Trip", false, fixed)
	if err != nil {
		t.Fatalf("Read raw error: %v", err)
	}
	if raw.Content != content {
		t.Fatalf("unexpected raw content: %q", raw.Content)
	}

	resolved, err := Read(root, cfg, "Travel", "Trip", true, fixed)
	if err != nil {
		t.Fatalf("Read resolved error: %v", err)
	}
	expected := "Date: 2026-02-20\nTime: 03:04\nTitle: Trip"
	if resolved.Content != expected {
		t.Fatalf("unexpected resolved content: %q", resolved.Content)
	}
}
