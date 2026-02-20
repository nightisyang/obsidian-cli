package note

import (
	"regexp"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

var headingLinePattern = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)

func ReadHeading(vaultRoot, path, heading string) (HeadingSection, error) {
	n, err := Read(vaultRoot, path)
	if err != nil {
		return HeadingSection{}, err
	}

	target := normalizeHeading(heading)
	if target == "" {
		return HeadingSection{}, errs.New(errs.ExitValidation, "heading is required")
	}

	lines := strings.Split(n.Body, "\n")
	start := -1
	level := 0
	foundHeading := ""
	for i, line := range lines {
		m := headingLinePattern.FindStringSubmatch(strings.TrimRight(line, "\r"))
		if len(m) != 3 {
			continue
		}
		lineHeading := normalizeHeading(m[2])
		if lineHeading != target {
			continue
		}
		start = i
		level = len(m[1])
		foundHeading = strings.TrimSpace(m[2])
		break
	}
	if start < 0 {
		return HeadingSection{}, errs.New(errs.ExitNotFound, "heading not found")
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		m := headingLinePattern.FindStringSubmatch(strings.TrimRight(lines[i], "\r"))
		if len(m) != 3 {
			continue
		}
		if len(m[1]) <= level {
			end = i
			break
		}
	}

	content := strings.Join(lines[start+1:end], "\n")
	return HeadingSection{
		Path:      n.Path,
		Heading:   foundHeading,
		Level:     level,
		Content:   content,
		StartLine: start + 1,
		EndLine:   end,
	}, nil
}

func normalizeHeading(raw string) string {
	out := strings.TrimSpace(raw)
	out = strings.TrimSpace(strings.TrimPrefix(out, "#"))
	out = strings.TrimSpace(strings.TrimSuffix(out, "#"))
	out = strings.Join(strings.Fields(out), " ")
	return strings.ToLower(out)
}
