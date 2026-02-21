package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchWithMetaIncludesMetadataAndWarnings(t *testing.T) {
	root := t.TempDir()
	content := "alpha\nalpha\n"
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	stdout, stderr, err := runCLI(
		t,
		"--vault", root,
		"--json",
		"search-content", "alpha",
		"--limit", "1",
		"--with-meta",
	)
	if err != nil {
		t.Fatalf("search with meta failed: %v (stderr=%q)", err, stderr)
	}
	env := parseEnvelope(t, stdout)
	var payload struct {
		Results  []map[string]any `json:"results"`
		Metadata struct {
			Truncated bool `json:"truncated"`
		} `json:"metadata"`
		Warnings []string `json:"warnings"`
	}
	if decodeErr := json.Unmarshal(env.Data, &payload); decodeErr != nil {
		t.Fatalf("decode search payload: %v", decodeErr)
	}
	if len(payload.Results) != 1 {
		t.Fatalf("expected 1 result due to limit, got %+v", payload.Results)
	}
	if !payload.Metadata.Truncated {
		t.Fatalf("expected truncated metadata: %+v", payload.Metadata)
	}
	if len(payload.Warnings) == 0 {
		t.Fatalf("expected warnings for truncation")
	}
}

func TestSearchStrictFailsOnWarnings(t *testing.T) {
	root := t.TempDir()
	content := "alpha\nalpha\n"
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	_, stderr, err := runCLI(
		t,
		"--vault", root,
		"search-content", "alpha",
		"--limit", "1",
		"--with-meta",
		"--strict",
	)
	if err == nil {
		t.Fatalf("expected strict mode failure")
	}
	if stderr != "" {
		t.Logf("stderr: %q", stderr)
	}
}

func TestGraphStrictFailsWhenTruncated(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("[[b]]\n[[c]]"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.md"), []byte("b"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "c.md"), []byte("c"), 0o644); err != nil {
		t.Fatalf("write c: %v", err)
	}

	_, _, err := runCLI(
		t,
		"--vault", root,
		"graph", "neighborhood", "a.md",
		"--depth", "1",
		"--node-limit", "1",
		"--strict",
	)
	if err == nil {
		t.Fatalf("expected strict mode failure for truncated graph")
	}
}

func TestNoteSizeLimitFailsOnCreate(t *testing.T) {
	root := t.TempDir()

	_, _, err := runCLI(
		t,
		"--vault", root,
		"--note-size-max-bytes", "1",
		"note", "create", "Large",
		"--content", "hello world",
	)
	if err == nil {
		t.Fatalf("expected note size limit failure on create")
	}
	if !strings.Contains(err.Error(), "exceeding 1 bytes") {
		t.Fatalf("unexpected note size limit error: %v", err)
	}
}

func TestNoteSizeLimitFailsOnAppend(t *testing.T) {
	root := t.TempDir()

	_, stderr, err := runCLI(
		t,
		"--vault", root,
		"note", "create", "Large",
		"--content", "seed",
	)
	if err != nil {
		t.Fatalf("note create seed failed: %v (stderr=%q)", err, stderr)
	}

	_, _, err = runCLI(
		t,
		"--vault", root,
		"--note-size-max-bytes", "1",
		"note", "append", "large.md", "more",
	)
	if err == nil {
		t.Fatalf("expected note size limit failure on append")
	}
	if !strings.Contains(err.Error(), "exceeding 1 bytes") {
		t.Fatalf("unexpected note size limit error: %v", err)
	}
}

func TestBacklinkValidationFailsOnCreateWithMissingTarget(t *testing.T) {
	root := t.TempDir()

	_, _, err := runCLI(
		t,
		"--vault", root,
		"note", "create", "Source",
		"--content", "See [[missing-target]]",
	)
	if err == nil {
		t.Fatalf("expected backlink target validation failure")
	}
	if !strings.Contains(err.Error(), "unresolved wikilinks: missing-target") {
		t.Fatalf("unexpected backlink validation error: %v", err)
	}
}

func TestBacklinkValidationPassesWhenTargetExists(t *testing.T) {
	root := t.TempDir()

	_, stderr, err := runCLI(
		t,
		"--vault", root,
		"note", "create", "Target",
	)
	if err != nil {
		t.Fatalf("create target failed: %v (stderr=%q)", err, stderr)
	}

	_, stderr, err = runCLI(
		t,
		"--vault", root,
		"note", "create", "Source",
		"--content", "See [[target]]",
	)
	if err != nil {
		t.Fatalf("create source with valid link failed: %v (stderr=%q)", err, stderr)
	}
}

func TestNoOrphanNotesFailsForIsolatedNote(t *testing.T) {
	root := t.TempDir()

	_, _, err := runCLI(
		t,
		"--vault", root,
		"--no-orphan-notes",
		"note", "create", "Isolated",
	)
	if err == nil {
		t.Fatalf("expected no-orphan-notes validation failure")
	}
	if !strings.Contains(err.Error(), "no graph connections") {
		t.Fatalf("unexpected no-orphan-notes error: %v", err)
	}
}

func TestNoOrphanNotesPassesWithOutgoingLink(t *testing.T) {
	root := t.TempDir()

	_, stderr, err := runCLI(
		t,
		"--vault", root,
		"note", "create", "Hub",
	)
	if err != nil {
		t.Fatalf("create hub failed: %v (stderr=%q)", err, stderr)
	}

	_, stderr, err = runCLI(
		t,
		"--vault", root,
		"--no-orphan-notes",
		"note", "create", "Node",
		"--content", "Related: [[hub]]",
	)
	if err != nil {
		t.Fatalf("create node with outgoing link failed: %v (stderr=%q)", err, stderr)
	}
}
