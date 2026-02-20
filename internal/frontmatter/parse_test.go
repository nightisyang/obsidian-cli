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
	_, _, _, err := ParseDocument("---\nfoo: [1\n---\nBody")
	if err == nil {
		t.Fatalf("expected parse error for malformed YAML")
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
