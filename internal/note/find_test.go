package note

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindByMetadataFilters(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	noteOne := `---
kind: task
tags:
  - alpha
  - beta
status: open
topic: infra
created_at: 2026-02-10T09:00:00Z
---

content`
	noteTwo := `---
kind: task
tags: [alpha]
status: done
topic: infra
created_at: 2026-02-11T09:00:00Z
---

content`
	if err := os.WriteFile(filepath.Join(root, "tasks", "one.md"), []byte(noteOne), 0o644); err != nil {
		t.Fatalf("write note one: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "tasks", "two.md"), []byte(noteTwo), 0o644); err != nil {
		t.Fatalf("write note two: %v", err)
	}

	since := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	results, err := FindByMetadata(root, FindFilters{
		Kind:   "task",
		Tags:   []string{"alpha", "beta"},
		Status: "open",
		Topic:  "infra",
		Since:  &since,
	})
	if err != nil {
		t.Fatalf("FindByMetadata error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d: %+v", len(results), results)
	}
	if results[0].Path != "tasks/one.md" {
		t.Fatalf("unexpected path: %s", results[0].Path)
	}
	if results[0].CreatedAt != "2026-02-10T09:00:00Z" {
		t.Fatalf("unexpected created_at: %s", results[0].CreatedAt)
	}
}

func TestFindByMetadataSupportsDateOnlyCreatedAt(t *testing.T) {
	root := t.TempDir()
	raw := `---
kind: note
created_at: "2026-02-15"
---

body`
	if err := os.WriteFile(filepath.Join(root, "date-only.md"), []byte(raw), 0o644); err != nil {
		t.Fatalf("write date note: %v", err)
	}

	since := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	results, err := FindByMetadata(root, FindFilters{Since: &since})
	if err != nil {
		t.Fatalf("FindByMetadata error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for date-only created_at, got %d", len(results))
	}
	if results[0].CreatedAt != "2026-02-15T00:00:00Z" {
		t.Fatalf("unexpected created_at normalization: %s", results[0].CreatedAt)
	}
}
