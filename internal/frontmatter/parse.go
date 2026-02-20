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
		if err := yaml.Unmarshal([]byte(yamlPart), &result); err != nil {
			return nil, "", false, err
		}
	}
	return result, body, true, nil
}
