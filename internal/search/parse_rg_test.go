package search

import (
	"strings"
	"testing"
)

func TestParseRGOutputStreamStopsAtLimit(t *testing.T) {
	raw := strings.Join([]string{
		`{"type":"match","data":{"path":{"text":"/vault/a.md"},"lines":{"text":"hello world"},"line_number":2,"submatches":[{"match":{"text":"hello"},"start":0,"end":5}]}}`,
		`{"type":"match","data":{"path":{"text":"/vault/b.md"},"lines":{"text":"hello two"},"line_number":3,"submatches":[{"match":{"text":"hello"},"start":0,"end":5}]}}`,
	}, "\n")

	results, limitReached, err := ParseRGOutputStream(strings.NewReader(raw), 80, 1)
	if err != nil {
		t.Fatalf("ParseRGOutputStream error: %v", err)
	}
	if !limitReached {
		t.Fatalf("expected limitReached=true")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Path != "/vault/a.md" {
		t.Fatalf("unexpected first result: %+v", results[0])
	}
}

func TestParseRGOutputStreamHandlesLongLines(t *testing.T) {
	longLine := strings.Repeat("x", 150000)
	raw := `{"type":"match","data":{"path":{"text":"/vault/a.md"},"lines":{"text":"` + longLine + `"},"line_number":1,"submatches":[{"match":{"text":"x"},"start":149999,"end":150000}]}}`

	results, limitReached, err := ParseRGOutputStream(strings.NewReader(raw), 5, 10)
	if err != nil {
		t.Fatalf("ParseRGOutputStream long line error: %v", err)
	}
	if limitReached {
		t.Fatalf("did not expect limitReached")
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Snippet == "" {
		t.Fatalf("expected non-empty snippet")
	}
}
