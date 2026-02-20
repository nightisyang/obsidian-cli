package index

import (
	"reflect"
	"testing"
)

func TestParseWikiLinks(t *testing.T) {
	body := "[[Note One]] [[Folder/Note Two|Alias]] [[note one]] [[Third#Section]]"
	got := ParseWikiLinks(body)
	want := []string{"folder/note two", "note one", "third"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected links. got=%v want=%v", got, want)
	}
}
