package tasks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListFiltersAndUpdate(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "notes"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	one := "- [ ] todo one\n- [x] done one\n"
	two := "# Title\n- [-] custom state\n"
	if err := os.WriteFile(filepath.Join(root, "notes", "a.md"), []byte(one), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes", "b.md"), []byte(two), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	all, err := List(root, ListOptions{})
	if err != nil {
		t.Fatalf("List all error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(all))
	}

	done, err := List(root, ListOptions{Done: true})
	if err != nil {
		t.Fatalf("List done error: %v", err)
	}
	if len(done) != 1 || done[0].Status != "x" {
		t.Fatalf("unexpected done results: %+v", done)
	}

	custom, err := List(root, ListOptions{Status: "-"})
	if err != nil {
		t.Fatalf("List status error: %v", err)
	}
	if len(custom) != 1 || custom[0].Path != "notes/b.md" {
		t.Fatalf("unexpected custom results: %+v", custom)
	}

	updated, err := Update(root, Ref{Path: "notes/a.md", Line: 1}, UpdateInput{Toggle: true})
	if err != nil {
		t.Fatalf("Update toggle error: %v", err)
	}
	if updated.Status != "x" {
		t.Fatalf("expected toggled status x, got %q", updated.Status)
	}

	got, err := Get(root, Ref{Path: "notes/a.md", Line: 1})
	if err != nil {
		t.Fatalf("Get after update error: %v", err)
	}
	if got.Status != "x" {
		t.Fatalf("expected persisted status x, got %q", got.Status)
	}
}

func TestParseRef(t *testing.T) {
	ref, err := ParseRef("notes/a.md:12")
	if err != nil {
		t.Fatalf("ParseRef error: %v", err)
	}
	if ref.Path != "notes/a.md" || ref.Line != 12 {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}
