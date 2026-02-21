package backend

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestBacklinksRebuildsWhenCacheIsStale(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("[[b]]"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.md"), []byte("target"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	b := NewNativeBackend(root, vault.DefaultConfig(), "native")
	before, err := b.Backlinks(context.Background(), "b.md", false)
	if err != nil {
		t.Fatalf("Backlinks initial error: %v", err)
	}
	if len(before) != 1 || before[0] != "a.md" {
		t.Fatalf("unexpected initial backlinks: %+v", before)
	}

	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("no links"), 0o644); err != nil {
		t.Fatalf("rewrite a: %v", err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(filepath.Join(root, "a.md"), future, future); err != nil {
		t.Fatalf("chtimes a: %v", err)
	}

	after, err := b.Backlinks(context.Background(), "b.md", false)
	if err != nil {
		t.Fatalf("Backlinks stale-rebuild error: %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("expected stale cache rebuild to remove backlinks, got %+v", after)
	}
}

func TestPropDelete(t *testing.T) {
	root := t.TempDir()
	content := `---
status: active
topic: infra
---

hello`
	if err := os.WriteFile(filepath.Join(root, "alpha.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	b := NewNativeBackend(root, vault.DefaultConfig(), "native")
	n, err := b.PropDelete(context.Background(), "alpha.md", "status")
	if err != nil {
		t.Fatalf("PropDelete error: %v", err)
	}
	if n.Path != "alpha.md" {
		t.Fatalf("unexpected note path after PropDelete: %+v", n)
	}

	updated, err := os.ReadFile(filepath.Join(root, "alpha.md"))
	if err != nil {
		t.Fatalf("read updated note: %v", err)
	}
	if strings.Contains(string(updated), "status: active") {
		t.Fatalf("expected status key to be removed, got %q", string(updated))
	}
}
