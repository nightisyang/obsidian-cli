package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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
	var continueOnError bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply <spec.json>",
		Short: "Apply a batch of CLI operations from JSON",
		Long:  "JSON format:\n{\n  \"ops\": [\n    {\"id\": \"1\", \"args\": [\"note\", \"append\", \"daily.md\", \"- [ ] follow up\"], \"expect\": {\"ok\": \"true\", \"data.path\": \"daily.md\"}}\n  ]\n}\n\nOperations run in order. On failure, succeeded and failed operations are reported; no automatic rollback is performed.",
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

			succeededOps := make([]opsResult, 0, len(results))
			failedOps := make([]opsResult, 0)
			for _, result := range results {
				if result.OK {
					succeededOps = append(succeededOps, result)
				} else {
					failedOps = append(failedOps, result)
				}
			}

			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{
					"results":       results,
					"failed":        failed,
					"succeeded_ops": succeededOps,
					"failed_ops":    failedOps,
				})
			}
			if len(succeededOps) > 0 {
				rt.Printer.Println("succeeded operations:")
				for _, result := range succeededOps {
					rt.Printer.Println(fmt.Sprintf("- %v", result.Args))
				}
			}
			if len(failedOps) > 0 {
				rt.Printer.Println("failed operations:")
				for _, result := range failedOps {
					rt.Printer.Println(fmt.Sprintf("- %v (%s)", result.Args, result.Error))
				}
			}
			if failed {
				return errs.New(errs.ExitGeneric, "batch apply failed")
			}
			return nil
		},
	}

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
