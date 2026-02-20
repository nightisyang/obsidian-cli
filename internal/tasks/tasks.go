package tasks

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/index"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

var taskLinePattern = regexp.MustCompile(`^(\s*[-*+]\s+\[)([^\]])(\]\s*)(.*)$`)

type Task struct {
	Path   string `json:"path"`
	Line   int    `json:"line"`
	Status string `json:"status"`
	Text   string `json:"text"`
	Raw    string `json:"raw"`
}

type ListOptions struct {
	Path   string
	Status string
	Done   bool
	Todo   bool
	Limit  int
}

type Ref struct {
	Path string `json:"path"`
	Line int    `json:"line"`
}

type UpdateInput struct {
	Status string
	Toggle bool
	Done   bool
	Todo   bool
}

func ParseRef(raw string) (Ref, error) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) < 2 {
		return Ref{}, errs.New(errs.ExitValidation, "task ref must be path:line")
	}
	linePart := parts[len(parts)-1]
	pathPart := strings.Join(parts[:len(parts)-1], ":")
	line := 0
	for _, ch := range linePart {
		if ch < '0' || ch > '9' {
			return Ref{}, errs.New(errs.ExitValidation, "task ref line must be numeric")
		}
		line = (line * 10) + int(ch-'0')
	}
	if line <= 0 {
		return Ref{}, errs.New(errs.ExitValidation, "task ref line must be greater than zero")
	}
	if strings.TrimSpace(pathPart) == "" {
		return Ref{}, errs.New(errs.ExitValidation, "task ref path is required")
	}
	return Ref{Path: pathPart, Line: line}, nil
}

func List(vaultRoot string, opts ListOptions) ([]Task, error) {
	paths := []string{}
	if strings.TrimSpace(opts.Path) != "" {
		abs, rel, err := vault.ResolveNoteAbs(vaultRoot, opts.Path)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(abs); err != nil {
			if os.IsNotExist(err) {
				return []Task{}, nil
			}
			return nil, err
		}
		paths = append(paths, filepath.ToSlash(rel))
	} else {
		files, err := index.ListMarkdownFiles(vaultRoot)
		if err != nil {
			return nil, err
		}
		for _, abs := range files {
			rel, _ := filepath.Rel(vaultRoot, abs)
			paths = append(paths, filepath.ToSlash(rel))
		}
	}

	out := []Task{}
	for _, path := range paths {
		raw, err := os.ReadFile(filepath.Join(vaultRoot, filepath.FromSlash(path)))
		if err != nil {
			return nil, err
		}
		tasks := tasksFromContent(path, string(raw))
		for _, t := range tasks {
			if !matchesFilter(t, opts) {
				continue
			}
			out = append(out, t)
			if opts.Limit > 0 && len(out) >= opts.Limit {
				return out, nil
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Path == out[j].Path {
			return out[i].Line < out[j].Line
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

func Get(vaultRoot string, ref Ref) (Task, error) {
	items, err := List(vaultRoot, ListOptions{Path: ref.Path})
	if err != nil {
		return Task{}, err
	}
	for _, item := range items {
		if item.Line == ref.Line {
			return item, nil
		}
	}
	return Task{}, errs.New(errs.ExitNotFound, "task not found")
}

func Update(vaultRoot string, ref Ref, in UpdateInput) (Task, error) {
	abs, rel, err := vault.ResolveNoteAbs(vaultRoot, ref.Path)
	if err != nil {
		return Task{}, err
	}
	payload, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return Task{}, errs.New(errs.ExitNotFound, "note not found")
		}
		return Task{}, err
	}

	raw := strings.ReplaceAll(string(payload), "\r\n", "\n")
	hasTrailingNewline := strings.HasSuffix(raw, "\n")
	lines := strings.Split(raw, "\n")
	if hasTrailingNewline {
		lines = lines[:len(lines)-1]
	}
	if ref.Line < 1 || ref.Line > len(lines) {
		return Task{}, errs.New(errs.ExitNotFound, "task line out of range")
	}

	m := taskLinePattern.FindStringSubmatch(lines[ref.Line-1])
	if len(m) != 5 {
		return Task{}, errs.New(errs.ExitNotFound, "task not found at line")
	}

	status := m[2]
	nextStatus, err := nextTaskStatus(status, in)
	if err != nil {
		return Task{}, err
	}

	lines[ref.Line-1] = m[1] + nextStatus + m[3] + m[4]
	updated := strings.Join(lines, "\n")
	if hasTrailingNewline {
		updated += "\n"
	}
	if err := os.WriteFile(abs, []byte(updated), 0o644); err != nil {
		return Task{}, err
	}

	return Task{
		Path:   filepath.ToSlash(rel),
		Line:   ref.Line,
		Status: nextStatus,
		Text:   strings.TrimSpace(m[4]),
		Raw:    lines[ref.Line-1],
	}, nil
}

func tasksFromContent(path, raw string) []Task {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(raw, "\n")
	items := []Task{}
	for i, line := range lines {
		m := taskLinePattern.FindStringSubmatch(line)
		if len(m) != 5 {
			continue
		}
		items = append(items, Task{
			Path:   path,
			Line:   i + 1,
			Status: m[2],
			Text:   strings.TrimSpace(m[4]),
			Raw:    line,
		})
	}
	return items
}

func matchesFilter(task Task, opts ListOptions) bool {
	if opts.Status != "" && task.Status != opts.Status {
		return false
	}
	if opts.Done && !isDoneStatus(task.Status) {
		return false
	}
	if opts.Todo && isDoneStatus(task.Status) {
		return false
	}
	return true
}

func isDoneStatus(status string) bool {
	return strings.EqualFold(status, "x")
}

func nextTaskStatus(current string, in UpdateInput) (string, error) {
	if in.Toggle {
		if isDoneStatus(current) {
			return " ", nil
		}
		return "x", nil
	}
	if in.Done {
		return "x", nil
	}
	if in.Todo {
		return " ", nil
	}
	if in.Status != "" {
		runes := []rune(in.Status)
		if len(runes) != 1 {
			return "", errs.New(errs.ExitValidation, "status must be a single character")
		}
		return string(runes[0]), nil
	}
	return current, nil
}
