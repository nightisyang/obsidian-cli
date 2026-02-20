package frontmatter

import (
	"fmt"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type Data struct {
	Title     string
	Tags      []string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Kind      string
	Topic     string
	Status    string
	Extra     map[string]any
}

func MapToFrontmatter(values map[string]any) Data {
	fm := Data{Extra: map[string]any{}}
	for k, v := range values {
		switch k {
		case "title":
			fm.Title = toString(v)
		case "tags":
			fm.Tags = toStringSlice(v)
		case "kind":
			fm.Kind = toString(v)
		case "topic":
			fm.Topic = toString(v)
		case "status":
			fm.Status = toString(v)
		case "created_at":
			if t := toTime(v); t != nil {
				fm.CreatedAt = t
			}
		case "updated_at":
			if t := toTime(v); t != nil {
				fm.UpdatedAt = t
			}
		default:
			fm.Extra[k] = v
		}
	}
	if len(fm.Extra) == 0 {
		fm.Extra = nil
	}
	return fm
}

func FrontmatterToMap(fm Data) map[string]any {
	out := map[string]any{}
	for k, v := range fm.Extra {
		out[k] = v
	}
	if fm.Title != "" {
		out["title"] = fm.Title
	}
	if len(fm.Tags) > 0 {
		out["tags"] = dedupeStrings(fm.Tags)
	}
	if fm.Kind != "" {
		out["kind"] = fm.Kind
	}
	if fm.Topic != "" {
		out["topic"] = fm.Topic
	}
	if fm.Status != "" {
		out["status"] = fm.Status
	}
	if fm.CreatedAt != nil {
		out["created_at"] = fm.CreatedAt.UTC().Format(time.RFC3339)
	}
	if fm.UpdatedAt != nil {
		out["updated_at"] = fm.UpdatedAt.UTC().Format(time.RFC3339)
	}
	return out
}

func RenderMarkdown(fm Data, body string) (string, error) {
	values := FrontmatterToMap(fm)
	if len(values) == 0 {
		return body, nil
	}

	node := &yaml.Node{Kind: yaml.MappingNode}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: k}
		valueNode := &yaml.Node{}
		if err := valueNode.Encode(values[k]); err != nil {
			return "", fmt.Errorf("encode yaml value %s: %w", k, err)
		}
		node.Content = append(node.Content, keyNode, valueNode)
	}

	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{node}}
	payload, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	return "---\n" + string(payload) + "---\n\n" + body, nil
}

func toString(input any) string {
	switch v := input.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func toStringSlice(input any) []string {
	out := []string{}
	switch v := input.(type) {
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	case []string:
		for _, item := range v {
			if item != "" {
				out = append(out, item)
			}
		}
	case string:
		if v != "" {
			out = append(out, v)
		}
	}
	return dedupeStrings(out)
}

func toTime(input any) *time.Time {
	switch v := input.(type) {
	case time.Time:
		t := v.UTC()
		return &t
	case string:
		if v == "" {
			return nil
		}
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			t = t.UTC()
			return &t
		}
	}
	return nil
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, value := range in {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
