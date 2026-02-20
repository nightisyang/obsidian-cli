package search

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/errs"
)

type RGEngine struct {
	VaultRoot string
}

func (e *RGEngine) Search(ctx context.Context, q Query) ([]SearchResult, error) {
	if q.Type != QueryText {
		return nil, errs.New(errs.ExitValidation, "ripgrep engine supports text queries only")
	}

	searchRoot, err := e.resolveSearchRoot(q.Path)
	if err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("rg"); err != nil {
		// rg not available — fall back to stdlib engine.
		fb := &StdlibEngine{VaultRoot: e.VaultRoot}
		return fb.Search(ctx, q)
	}

	args := []string{"--json", "--line-number", "--glob", "*.md"}
	if !q.CaseSensitive {
		args = append(args, "-i")
	}
	args = append(args, q.Text, searchRoot)
	cmd := exec.CommandContext(ctx, "rg", args...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return []SearchResult{}, nil
			}
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, errs.Wrap(errs.ExitGeneric, "ripgrep search failed", errWithMessage(message))
	}

	results, parseErr := ParseRGOutput(stdout, q.Context, q.Limit)
	if parseErr != nil {
		return nil, parseErr
	}
	for i := range results {
		results[i].Path = filepath.ToSlash(strings.TrimPrefix(results[i].Path, e.VaultRoot+string(filepath.Separator)))
	}
	return results, nil
}

type errWithMessage string

func (e errWithMessage) Error() string {
	return string(e)
}

func (e *RGEngine) resolveSearchRoot(pathPrefix string) (string, error) {
	root := filepath.Clean(e.VaultRoot)
	if strings.TrimSpace(pathPrefix) == "" {
		return root, nil
	}

	candidate := filepath.Clean(filepath.Join(root, filepath.FromSlash(pathPrefix)))
	if candidate == root || strings.HasPrefix(candidate, root+string(filepath.Separator)) {
		return candidate, nil
	}
	return "", errs.New(errs.ExitValidation, "search path escapes vault root")
}
