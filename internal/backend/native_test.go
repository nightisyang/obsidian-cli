package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

func TestCreateNoteWithTemplate(t *testing.T) {
	root := t.TempDir()
	templateDir := filepath.Join(root, ".obsidian", "templates")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(templateDir, "Daily.md"), []byte("Hello {{title}} on {{date}}"), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	oldNow := now
	now = func() time.Time { return time.Date(2026, 2, 20, 1, 2, 3, 0, time.UTC) }
	t.Cleanup(func() { now = oldNow })

	b := NewNativeBackend(root, vault.DefaultConfig(), "native")
	created, err := b.CreateNote(context.Background(), note.CreateInput{Title: "Trip", Template: "Daily"})
	if err != nil {
		t.Fatalf("CreateNote error: %v", err)
	}
	if created.Body != "Hello Trip on 2026-02-20" {
		t.Fatalf("unexpected body: %q", created.Body)
	}
}

func TestListPluginsAndOpenURI(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".obsidian", "plugins", "calendar"), 0o755); err != nil {
		t.Fatalf("mkdir plugins: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".obsidian", "core-plugins.json"), []byte(`["sync","daily-notes"]`), 0o644); err != nil {
		t.Fatalf("write core plugins: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".obsidian", "community-plugins.json"), []byte(`["calendar"]`), 0o644); err != nil {
		t.Fatalf("write community plugins: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".obsidian", "plugins", "calendar", "manifest.json"), []byte(`{"name":"Calendar","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	if _, err := note.Create(root, note.CreateInput{Title: "Alpha"}); err != nil {
		t.Fatalf("create note: %v", err)
	}

	b := NewNativeBackend(root, vault.DefaultConfig(), "native")
	plugins, err := b.ListPlugins(context.Background(), "", false)
	if err != nil {
		t.Fatalf("ListPlugins error: %v", err)
	}
	if len(plugins) < 3 {
		t.Fatalf("expected plugins from core+community, got %+v", plugins)
	}

	result, err := b.OpenInObsidian(context.Background(), "alpha.md", false)
	if err != nil {
		t.Fatalf("OpenInObsidian error: %v", err)
	}
	if result.URI == "" {
		t.Fatalf("expected non-empty URI")
	}
}
