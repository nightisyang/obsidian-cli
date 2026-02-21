package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
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
