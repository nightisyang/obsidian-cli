package frontmatter

import (
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseDocument(raw string) (map[string]any, string, bool, error) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return map[string]any{}, normalized, false, nil
	}

	lines := strings.Split(normalized, "\n")
	closing := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			closing = i
			break
		}
	}
	if closing < 0 {
		return map[string]any{}, normalized, false, nil
	}

	yamlPart := strings.Join(lines[1:closing], "\n")
	body := ""
	if closing+1 < len(lines) {
		body = strings.Join(lines[closing+1:], "\n")
	}
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}

	result := map[string]any{}
	if strings.TrimSpace(yamlPart) != "" {
		sanitized := sanitizeYAML(yamlPart)
		if err := yaml.Unmarshal([]byte(sanitized), &result); err != nil {
			// Malformed frontmatter: treat as no frontmatter rather than failing.
			// Callers should still be able to read/list the note.
			return map[string]any{}, body, false, nil
		}
	}
	return result, body, true, nil
}

// sanitizeYAML normalizes common Obsidian conventions that are technically
// invalid YAML. Specifically, tags like [#tag1, #tag2] use '#' which the YAML
// spec treats as a comment when preceded by whitespace. We strip the '#'
// prefix from values inside flow sequences so the YAML parses correctly.
func sanitizeYAML(yamlPart string) string {
	lines := strings.Split(yamlPart, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "tags:") {
			lines[i] = sanitizeHashTags(line)
		}
	}
	return strings.Join(lines, "\n")
}

// sanitizeHashTags removes '#' prefix from flow-sequence entries in a tags line.
// "tags: [#foo, #bar]" → "tags: [foo, bar]"
func sanitizeHashTags(line string) string {
	colon := strings.Index(line, ":")
	if colon < 0 {
		return line
	}
	key := line[:colon+1]
	val := strings.TrimSpace(line[colon+1:])
	if !strings.HasPrefix(val, "[") {
		return line
	}
	// Strip '#' from each item inside the brackets.
	inner := val[1 : len(val)-1]
	parts := strings.Split(inner, ",")
	for j, p := range parts {
		trimmed := strings.TrimSpace(p)
		if strings.HasPrefix(trimmed, "#") {
			parts[j] = strings.Replace(p, "#", "", 1)
		}
	}
	return key + " [" + strings.Join(parts, ",") + "]"
}
