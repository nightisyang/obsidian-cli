package frontmatter

import "testing"

func TestFrontmatterRoundTripPreservesUnknownFields(t *testing.T) {
	raw := "---\nkind: note\ntags:\n  - Alpha\ncustom_field:\n  nested: true\ncount: 3\n---\n\nBody text"
	values, body, hasFM, err := ParseDocument(raw)
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}
	if !hasFM {
		t.Fatalf("expected frontmatter block")
	}
	if body != "Body text" {
		t.Fatalf("unexpected body: %q", body)
	}

	fm := MapToFrontmatter(values)
	if fm.Kind != "note" {
		t.Fatalf("expected kind note, got %q", fm.Kind)
	}
	if fm.Extra == nil {
		t.Fatalf("expected extra map")
	}
	if _, ok := fm.Extra["custom_field"]; !ok {
		t.Fatalf("expected custom_field in extra")
	}

	rendered, err := RenderMarkdown(fm, body)
	if err != nil {
		t.Fatalf("RenderMarkdown error: %v", err)
	}
	v2, body2, _, err := ParseDocument(rendered)
	if err != nil {
		t.Fatalf("ParseDocument(rendered) error: %v", err)
	}
	if body2 != body {
		t.Fatalf("body changed after roundtrip: %q vs %q", body2, body)
	}
	if _, ok := v2["custom_field"]; !ok {
		t.Fatalf("custom_field lost after roundtrip")
	}
	if got, ok := v2["count"]; !ok || got != 3 {
		t.Fatalf("count field changed after roundtrip: %v", got)
	}
}
