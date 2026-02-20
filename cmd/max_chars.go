package cmd

import (
	"fmt"

	"github.com/nightisyang/obsidian-cli/internal/frontmatter"
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/nightisyang/obsidian-cli/internal/search"
)

func truncateChars(value string, max int) (string, bool) {
	if max <= 0 {
		return value, false
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value, false
	}
	return string(runes[:max]), true
}

func applyNoteBodyMaxChars(n note.Note, max int) (note.Note, error) {
	trimmed, changed := truncateChars(n.Body, max)
	if !changed {
		return n, nil
	}
	n.Body = trimmed + fmt.Sprintf("\n\n[TRUNCATED at %d chars \u2014 use --heading to read a specific section]", max)

	rendered, err := frontmatter.RenderMarkdown(n.Frontmatter, n.Body)
	if err != nil {
		return note.Note{}, err
	}
	n.Raw = rendered
	return n, nil
}

func applySearchSnippetMaxChars(results []search.SearchResult, max int) []search.SearchResult {
	if max <= 0 {
		return results
	}
	out := make([]search.SearchResult, len(results))
	copy(out, results)
	for i := range out {
		trimmed, changed := truncateChars(out[i].Snippet, max)
		if changed {
			out[i].Snippet = trimmed
		}
	}
	return out
}
