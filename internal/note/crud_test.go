package note

import (
	"testing"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

func TestNoteCRUDLifecycle(t *testing.T) {
	root := t.TempDir()

	created, err := Create(root, CreateInput{Title: "My Test Note", Tags: []string{"alpha"}})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if created.Path != "my-test-note.md" {
		t.Fatalf("unexpected path: %s", created.Path)
	}
	if created.Frontmatter.CreatedAt == nil || created.Frontmatter.UpdatedAt == nil {
		t.Fatalf("timestamps were not set")
	}

	got, err := Get(root, created.Path)
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.Title != "my-test-note" {
		t.Fatalf("unexpected title: %s", got.Title)
	}

	appended, err := Append(root, created.Path, "line one")
	if err != nil {
		t.Fatalf("Append error: %v", err)
	}
	if appended.Body != "line one" {
		t.Fatalf("unexpected appended body: %q", appended.Body)
	}

	prepended, err := Prepend(root, created.Path, "intro")
	if err != nil {
		t.Fatalf("Prepend error: %v", err)
	}
	if prepended.Body != "intro\nline one" {
		t.Fatalf("unexpected prepended body: %q", prepended.Body)
	}

	listed, err := List(root, "", ListOptions{Recursive: true})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 note, got %d", len(listed))
	}

	if err := Delete(root, created.Path); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	_, err = Get(root, created.Path)
	if err == nil {
		t.Fatalf("expected not found after delete")
	}
	appErr, ok := err.(*errs.AppError)
	if !ok || appErr.Code != errs.ExitNotFound {
		t.Fatalf("expected not found app error, got %T (%v)", err, err)
	}
}
