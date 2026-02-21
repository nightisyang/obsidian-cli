package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nightisyang/obsidian-cli/internal/app"
	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/spf13/cobra"
)

type opsSpec struct {
	Ops []opsItem `json:"ops"`
}

type opsItem struct {
	ID     string            `json:"id,omitempty"`
	Args   []string          `json:"args"`
	Expect map[string]string `json:"expect,omitempty"`
}

type opsResult struct {
	ID       string   `json:"id,omitempty"`
	Args     []string `json:"args"`
	OK       bool     `json:"ok"`
	ExitCode int      `json:"exit_code"`
	Error    string   `json:"error,omitempty"`
	Stdout   any      `json:"stdout,omitempty"`
	Stderr   string   `json:"stderr,omitempty"`
}

func newOpsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ops",
		Short: "Batch operation commands",
	}
	cmd.AddCommand(newOpsApplyCmd())
	return cmd
}

func newOpsApplyCmd() *cobra.Command {
	var rollback bool
	var continueOnError bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply <spec.json>",
		Short: "Apply a batch of CLI operations from JSON",
		Long:  "JSON format:\n{\n  \"ops\": [\n    {\"id\": \"1\", \"args\": [\"note\", \"append\", \"daily.md\", \"- [ ] follow up\"], \"expect\": {\"ok\": \"true\", \"data.path\": \"daily.md\"}}\n  ]\n}",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}

			spec, err := readOpsSpec(args[0])
			if err != nil {
				return err
			}
			if len(spec.Ops) == 0 {
				return errs.New(errs.ExitValidation, "ops list is empty")
			}

			if dryRun {
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run": true,
						"ops":     spec.Ops,
					})
				}
				rt.Printer.Println(fmt.Sprintf("dry-run: would execute %d operations", len(spec.Ops)))
				return nil
			}

			snapshotDir := ""
			if rollback {
				snapshotDir, err = os.MkdirTemp("", "obsidian-cli-ops-snapshot-*")
				if err != nil {
					return errs.Wrap(errs.ExitGeneric, "failed to create rollback snapshot dir", err)
				}
				defer os.RemoveAll(snapshotDir)
				if err := copyDir(rt.VaultRoot, snapshotDir); err != nil {
					return errs.Wrap(errs.ExitGeneric, "failed to create rollback snapshot", err)
				}
			}

			results := []opsResult{}
			failed := false
			for _, op := range spec.Ops {
				result := executeOpsItem(rt, op)
				results = append(results, result)
				if !result.OK {
					failed = true
					if !continueOnError {
						break
					}
				}
			}

			rolledBack := false
			if failed && rollback {
				if err := restoreSnapshot(snapshotDir, rt.VaultRoot); err != nil {
					return errs.Wrap(errs.ExitGeneric, "failed to rollback after batch failure", err)
				}
				rolledBack = true
			}

			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{
					"results":     results,
					"failed":      failed,
					"rolled_back": rolledBack,
				})
			}
			for _, result := range results {
				if result.OK {
					rt.Printer.Println(fmt.Sprintf("ok: %v", result.Args))
				} else {
					rt.Printer.Println(fmt.Sprintf("failed: %v (%s)", result.Args, result.Error))
				}
			}
			if rolledBack {
				rt.Printer.Println("rollback: restored vault snapshot")
			}
			if failed {
				return errs.New(errs.ExitGeneric, "batch apply failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&rollback, "rollback", false, "Restore vault snapshot if any operation fails")
	cmd.Flags().BoolVar(&continueOnError, "continue-on-error", false, "Continue applying operations after a failure")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation list without executing")
	return cmd
}

func readOpsSpec(path string) (opsSpec, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return opsSpec{}, errs.Wrap(errs.ExitValidation, "failed to read ops spec", err)
	}
	var spec opsSpec
	if err := json.Unmarshal(payload, &spec); err != nil {
		return opsSpec{}, errs.Wrap(errs.ExitValidation, "invalid ops spec JSON", err)
	}
	for i, op := range spec.Ops {
		if len(op.Args) == 0 {
			return opsSpec{}, errs.New(errs.ExitValidation, fmt.Sprintf("ops[%d].args is required", i))
		}
	}
	return spec, nil
}

func executeOpsItem(rt *app.Runtime, op opsItem) opsResult {
	baseArgs := []string{"--vault", rt.VaultRoot, "--json"}
	if rt.ConfigPath != "" {
		baseArgs = append(baseArgs, "--config", rt.ConfigPath)
	}
	if rt.EffectiveMode != "" {
		baseArgs = append(baseArgs, "--mode", rt.EffectiveMode)
	}
	baseArgs = append(baseArgs, op.Args...)

	cmd := exec.Command(os.Args[0], baseArgs...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	result := opsResult{
		ID:   op.ID,
		Args: append([]string(nil), op.Args...),
	}
	if err := cmd.Run(); err != nil {
		exitCode := errs.ExitGeneric
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		}
		result.OK = false
		result.ExitCode = exitCode
		result.Error = err.Error()
		result.Stderr = stderr.String()
		return result
	}

	result.OK = true
	result.ExitCode = 0
	result.Stderr = stderr.String()

	var parsed any
	if unmarshalErr := json.Unmarshal(stdout.Bytes(), &parsed); unmarshalErr == nil {
		result.Stdout = parsed
	} else {
		result.Stdout = stdout.String()
	}
	if len(op.Expect) > 0 {
		ok, message := checkExpectations(parsed, op.Expect)
		if !ok {
			result.OK = false
			result.ExitCode = errs.ExitValidation
			result.Error = message
		}
	}
	return result
}

func checkExpectations(payload any, expects map[string]string) (bool, string) {
	for key, expected := range expects {
		actual, ok := lookupPath(payload, key)
		if !ok {
			return false, fmt.Sprintf("expectation key not found: %s", key)
		}
		if fmt.Sprintf("%v", actual) != expected {
			return false, fmt.Sprintf("expectation failed for %s: expected %q got %v", key, expected, actual)
		}
	}
	return true, ""
}

func lookupPath(payload any, path string) (any, bool) {
	if path == "" {
		return payload, true
	}
	current := payload
	segments := splitPath(path)
	for _, segment := range segments {
		switch typed := current.(type) {
		case map[string]any:
			next, ok := typed[segment]
			if !ok {
				return nil, false
			}
			current = next
		default:
			return nil, false
		}
	}
	return current, true
}

func splitPath(path string) []string {
	out := []string{}
	part := ""
	for _, r := range path {
		if r == '.' {
			if part != "" {
				out = append(out, part)
				part = ""
			}
			continue
		}
		part += string(r)
	}
	if part != "" {
		out = append(out, part)
	}
	return out
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func restoreSnapshot(snapshot, target string) error {
	entries, err := os.ReadDir(target)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if rmErr := os.RemoveAll(filepath.Join(target, entry.Name())); rmErr != nil {
			return rmErr
		}
	}
	return copyDir(snapshot, target)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
