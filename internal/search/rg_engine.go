package search

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
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

	args := []string{"--json", "--line-number", "--glob", "*.md", "--no-messages"}
	if q.Limit > 0 {
		args = append(args, "--max-count", strconv.Itoa(q.Limit))
	}
	if !q.CaseSensitive {
		args = append(args, "-i")
	}
	args = append(args, q.Text, searchRoot)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "rg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errs.Wrap(errs.ExitGeneric, "failed to capture ripgrep stdout", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errs.Wrap(errs.ExitGeneric, "failed to capture ripgrep stderr", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, errs.Wrap(errs.ExitGeneric, "failed to start ripgrep search", err)
	}

	results, limitReached, parseErr := ParseRGOutputStream(stdout, q.Context, q.Limit)
	if parseErr != nil {
		cancel()
		_, _ = io.Copy(io.Discard, stderr)
		_ = cmd.Wait()
		return nil, parseErr
	}

	if limitReached {
		cancel()
	}

	stderrBytes, _ := io.ReadAll(stderr)
	waitErr := cmd.Wait()
	if waitErr != nil {
		// ripgrep exits 1 when no matches are found.
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) && exitErr.ExitCode() == 1 {
			return []SearchResult{}, nil
		}
		// Early cancellation after reaching limit is expected.
		if limitReached && (errors.Is(waitErr, context.Canceled) || errors.Is(runCtx.Err(), context.Canceled)) {
			waitErr = nil
		}
		if waitErr != nil {
			message := strings.TrimSpace(string(stderrBytes))
			if message == "" {
				message = waitErr.Error()
			}
			return nil, errs.Wrap(errs.ExitGeneric, "ripgrep search failed", fmt.Errorf("%s", message))
		}
	}

	for i := range results {
		rel, relErr := filepath.Rel(e.VaultRoot, results[i].Path)
		if relErr != nil {
			continue
		}
		results[i].Path = filepath.ToSlash(rel)
	}
	return results, nil
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
