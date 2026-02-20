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

	rest := normalized[len("---\n"):]
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		if strings.HasSuffix(rest, "\n---") {
			idx = len(rest) - len("\n---")
		} else {
			return map[string]any{}, normalized, false, nil
		}
	}

	yamlPart := rest[:idx]
	body := ""
	endOffset := idx + len("\n---\n")
	if endOffset <= len(rest) {
		body = rest[endOffset:]
		if strings.HasPrefix(body, "\n") {
			body = body[1:]
		}
	}

	result := map[string]any{}
	if strings.TrimSpace(yamlPart) != "" {
		if err := yaml.Unmarshal([]byte(yamlPart), &result); err != nil {
			return nil, "", false, err
		}
	}
	return result, body, true, nil
}
