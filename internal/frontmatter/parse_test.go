package frontmatter

import "testing"

func TestParseDocumentEmptyFrontmatter(t *testing.T) {
	values, body, hasFM, err := ParseDocument("---\n---\nBody")
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}
	if !hasFM {
		t.Fatalf("expected frontmatter block")
	}
	if len(values) != 0 {
		t.Fatalf("expected empty frontmatter map, got %+v", values)
	}
	if body != "Body" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestParseDocumentFrontmatterOnly(t *testing.T) {
	values, body, hasFM, err := ParseDocument("---\nstatus: ok\n---\n")
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}
	if !hasFM {
		t.Fatalf("expected frontmatter block")
	}
	if values["status"] != "ok" {
		t.Fatalf("unexpected status value: %+v", values["status"])
	}
	if body != "" {
		t.Fatalf("expected empty body, got %q", body)
	}
}

func TestParseDocumentMalformedYAML(t *testing.T) {
	// Malformed YAML should be tolerated: return empty frontmatter + body, no error.
	values, body, hasFM, err := ParseDocument("---\nfoo: [1\n---\nBody")
	if err != nil {
		t.Fatalf("expected no error for malformed YAML (tolerant mode), got: %v", err)
	}
	if hasFM {
		t.Fatalf("expected hasFM=false when YAML parse fails")
	}
	if len(values) != 0 {
		t.Fatalf("expected empty values map for malformed YAML, got %+v", values)
	}
	if body != "Body" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestParseDocumentHashTagSanitization(t *testing.T) {
	// Obsidian convention: tags with # prefix in flow sequences should be normalized.
	raw := "---\ntags: [#architecture, #brain-evolution, #milestone]\nstatus: active\n---\nBody"
	values, body, hasFM, err := ParseDocument(raw)
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}
	if !hasFM {
		t.Fatalf("expected frontmatter block")
	}
	tags, ok := values["tags"].([]any)
	if !ok || len(tags) == 0 {
		t.Fatalf("expected tags slice, got %T %v", values["tags"], values["tags"])
	}
	if values["status"] != "active" {
		t.Fatalf("expected status=active, got %v", values["status"])
	}
	if body != "Body" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestParseDocumentWithoutFrontmatter(t *testing.T) {
	values, body, hasFM, err := ParseDocument("Title\nBody")
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}
	if hasFM {
		t.Fatalf("did not expect frontmatter")
	}
	if len(values) != 0 {
		t.Fatalf("expected empty map, got %+v", values)
	}
	if body != "Title\nBody" {
		t.Fatalf("unexpected body: %q", body)
	}
}
