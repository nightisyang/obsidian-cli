package search

import "testing"

func TestBuildQueryModes(t *testing.T) {
	textQ, err := BuildQuery("hello", "", "", 10, 50, "", false)
	if err != nil {
		t.Fatalf("BuildQuery text error: %v", err)
	}
	if textQ.Type != QueryText || textQ.Text != "hello" {
		t.Fatalf("unexpected text query: %+v", textQ)
	}

	tagQ, err := BuildQuery("", "#project", "", 20, 80, "notes", false)
	if err != nil {
		t.Fatalf("BuildQuery tag error: %v", err)
	}
	if tagQ.Type != QueryTag || tagQ.Tag != "project" {
		t.Fatalf("unexpected tag query: %+v", tagQ)
	}

	propQ, err := BuildQuery("", "", "status=done", 5, 80, "", true)
	if err != nil {
		t.Fatalf("BuildQuery prop error: %v", err)
	}
	if propQ.Type != QueryProp || propQ.PropKey != "status" || propQ.PropValue != "done" {
		t.Fatalf("unexpected prop query: %+v", propQ)
	}
}

func TestBuildQueryValidation(t *testing.T) {
	if _, err := BuildQuery("a", "b", "", 0, 0, "", false); err == nil {
		t.Fatalf("expected validation error for multiple query modes")
	}
	if _, _, err := ParsePropFilter("broken"); err == nil {
		t.Fatalf("expected ParsePropFilter error")
	}
}
