package search

import (
	"fmt"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

type QueryType string

const (
	QueryText QueryType = "text"
	QueryTag  QueryType = "tag"
	QueryProp QueryType = "prop"
)

type Query struct {
	Type          QueryType `json:"type"`
	Text          string    `json:"text,omitempty"`
	Tag           string    `json:"tag,omitempty"`
	PropKey       string    `json:"prop_key,omitempty"`
	PropValue     string    `json:"prop_value,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Context       int       `json:"context,omitempty"`
	Path          string    `json:"path,omitempty"`
	CaseSensitive bool      `json:"case_sensitive,omitempty"`
}

type SearchResult struct {
	Path      string `json:"path"`
	Line      int    `json:"line,omitempty"`
	Column    int    `json:"column,omitempty"`
	Match     string `json:"match"`
	Snippet   string `json:"snippet"`
	MatchType string `json:"match_type"`
}

func BuildQuery(text, tag, prop string, limit, context int, path string, caseSensitive bool) (Query, error) {
	hasText := strings.TrimSpace(text) != ""
	hasTag := strings.TrimSpace(tag) != ""
	hasProp := strings.TrimSpace(prop) != ""
	count := 0
	if hasText {
		count++
	}
	if hasTag {
		count++
	}
	if hasProp {
		count++
	}
	if count != 1 {
		return Query{}, errs.New(errs.ExitValidation, "provide exactly one of query text, --tag, or --prop")
	}

	q := Query{Limit: limit, Context: context, Path: strings.TrimSpace(path), CaseSensitive: caseSensitive}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Context < 0 {
		q.Context = 80
	}

	if hasText {
		q.Type = QueryText
		q.Text = text
		return q, nil
	}
	if hasTag {
		q.Type = QueryTag
		q.Tag = strings.TrimPrefix(strings.TrimSpace(tag), "#")
		return q, nil
	}

	key, value, err := ParsePropFilter(prop)
	if err != nil {
		return Query{}, err
	}
	q.Type = QueryProp
	q.PropKey = key
	q.PropValue = value
	return q, nil
}

func ParsePropFilter(raw string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(raw), "=", 2)
	if len(parts) != 2 {
		return "", "", errs.New(errs.ExitValidation, "property filter must be key=value")
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" {
		return "", "", errs.New(errs.ExitValidation, "property key is required")
	}
	if value == "" {
		return "", "", errs.New(errs.ExitValidation, "property value is required")
	}
	return key, value, nil
}

func (q Query) String() string {
	switch q.Type {
	case QueryText:
		return q.Text
	case QueryTag:
		return "#" + q.Tag
	case QueryProp:
		return fmt.Sprintf("%s=%s", q.PropKey, q.PropValue)
	default:
		return ""
	}
}
