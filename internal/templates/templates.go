package templates

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

type TemplateInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Template struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

func List(vaultRoot string, cfg vault.Config) ([]TemplateInfo, error) {
	dir := templatesDir(vaultRoot, cfg)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []TemplateInfo{}, nil
		}
		return nil, err
	}

	out := []TemplateInfo{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".md") {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		out = append(out, TemplateInfo{
			Name: base,
			Path: filepath.ToSlash(filepath.Join(strings.Trim(cfg.TemplatesDir, "/"), name)),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func Read(vaultRoot string, cfg vault.Config, name, title string, resolve bool, at time.Time) (Template, error) {
	absPath, relPath, err := resolveTemplatePath(vaultRoot, cfg, name)
	if err != nil {
		return Template{}, err
	}

	payload, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Template{}, errs.New(errs.ExitNotFound, "template not found")
		}
		return Template{}, err
	}
	content := string(payload)
	if resolve {
		content = ResolveVariables(content, title, at)
	}

	return Template{
		Name:    strings.TrimSuffix(filepath.Base(relPath), filepath.Ext(relPath)),
		Path:    filepath.ToSlash(relPath),
		Content: content,
	}, nil
}

func ResolveVariables(content, title string, at time.Time) string {
	replacer := strings.NewReplacer(
		"{{date}}", at.Format("2006-01-02"),
		"{{time}}", at.Format("15:04"),
		"{{title}}", title,
	)
	return replacer.Replace(content)
}

func templatesDir(vaultRoot string, cfg vault.Config) string {
	dir := strings.TrimSpace(cfg.TemplatesDir)
	if dir == "" {
		dir = ".obsidian/templates"
	}
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Clean(filepath.Join(vaultRoot, dir))
}

func resolveTemplatePath(vaultRoot string, cfg vault.Config, name string) (string, string, error) {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return "", "", errs.New(errs.ExitValidation, "template name is required")
	}
	dir := templatesDir(vaultRoot, cfg)

	candidates := []string{}
	if strings.Contains(cleanName, "/") || strings.Contains(cleanName, "\\") {
		candidates = append(candidates, cleanName)
	} else {
		candidates = append(candidates, cleanName+".md", cleanName)
	}

	for _, candidate := range candidates {
		base := strings.TrimPrefix(strings.ReplaceAll(candidate, "\\", "/"), "/")
		base = filepath.FromSlash(base)
		if !strings.HasSuffix(strings.ToLower(base), ".md") {
			base += ".md"
		}
		abs := filepath.Clean(filepath.Join(dir, base))
		if !isPathWithinRoot(dir, abs) {
			return "", "", errs.New(errs.ExitValidation, "template path escapes templates directory")
		}
		if _, err := os.Stat(abs); err == nil {
			rel, _ := filepath.Rel(vaultRoot, abs)
			return abs, filepath.ToSlash(rel), nil
		}
	}

	infos, err := List(vaultRoot, cfg)
	if err != nil {
		return "", "", err
	}
	lower := strings.ToLower(cleanName)
	for _, info := range infos {
		if strings.ToLower(info.Name) == lower {
			abs := filepath.Join(vaultRoot, filepath.FromSlash(info.Path))
			return abs, info.Path, nil
		}
	}
	return "", "", errs.New(errs.ExitNotFound, "template not found")
}

func isPathWithinRoot(root, candidate string) bool {
	cleanRoot := filepath.Clean(root)
	cleanCandidate := filepath.Clean(candidate)
	if cleanCandidate == cleanRoot {
		return true
	}
	return strings.HasPrefix(cleanCandidate, cleanRoot+string(filepath.Separator))
}
