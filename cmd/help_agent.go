package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/output"
	"github.com/spf13/cobra"
)

type agentContract struct {
	Path            string   `json:"path"`
	Use             string   `json:"use"`
	Summary         string   `json:"summary"`
	Intent          string   `json:"intent"`
	SideEffects     string   `json:"side_effects"`
	Idempotent      bool     `json:"idempotent"`
	SupportsDryRun  bool     `json:"supports_dry_run"`
	Arguments       []string `json:"arguments,omitempty"`
	RecommendedFlow []string `json:"recommended_flow,omitempty"`
}

type agentSkill struct {
	Name    string   `json:"name"`
	Purpose string   `json:"purpose"`
	Steps   []string `json:"steps"`
}

type agentHelpPayload struct {
	FormatVersion string          `json:"format_version"`
	Tool          string          `json:"tool"`
	Guidance      []string        `json:"guidance"`
	Skills        []agentSkill    `json:"skills,omitempty"`
	Commands      []agentContract `json:"commands"`
}

func newHelpCmd(root *cobra.Command) *cobra.Command {
	var agent bool
	var format string
	var skill string

	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long:  "Help about any command. Use --agent for LLM-oriented command contracts.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !agent && strings.TrimSpace(skill) == "" {
				if len(args) == 0 {
					return root.Help()
				}
				target, _, err := root.Find(args)
				if err != nil {
					return err
				}
				return target.Help()
			}

			payload, err := buildAgentHelpPayload(root, args, strings.TrimSpace(skill))
			if err != nil {
				return err
			}

			switch strings.ToLower(strings.TrimSpace(format)) {
			case "json":
				return output.WriteJSON(os.Stdout, payload)
			case "text", "":
				printAgentHelpText(payload)
				return nil
			default:
				return fmt.Errorf("unsupported --format %q, use text or json", format)
			}
		},
	}

	cmd.Flags().BoolVar(&agent, "agent", false, "Emit agent-focused command contracts")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text|json")
	cmd.Flags().StringVar(&skill, "skill", "", "Emit a targeted skill block (e.g. note-lifecycle, task-management)")
	return cmd
}

func buildAgentHelpPayload(root *cobra.Command, args []string, skill string) (agentHelpPayload, error) {
	target := root
	if len(args) > 0 {
		found, _, err := root.Find(args)
		if err != nil {
			return agentHelpPayload{}, err
		}
		target = found
	}

	commands := collectContracts(target)
	payload := agentHelpPayload{
		FormatVersion: "agent-help.v1",
		Tool:          "obsidian-cli",
		Guidance: []string{
			"Prefer read/discovery commands first: list, note list, search, schema.",
			"For mutating commands, use --dry-run whenever available before applying.",
			"Use --json for deterministic machine output and parse the data envelope.",
			"Use schema to discover flags and args instead of relying on assumptions.",
		},
		Commands: commands,
	}
	if skill != "" {
		payload.Skills = []agentSkill{skillByName(skill)}
	}
	return payload, nil
}

func collectContracts(root *cobra.Command) []agentContract {
	contracts := []agentContract{}
	var walk func(node *cobra.Command)
	walk = func(node *cobra.Command) {
		if shouldSkipFromCatalog(node) {
			return
		}
		relPath := relativeCommandPath(node)
		if relPath != "" {
			traits := traitForCommand(node)
			contract := agentContract{
				Path:            relPath,
				Use:             node.UseLine(),
				Summary:         node.Short,
				Intent:          traits.Intent,
				SideEffects:     traits.SideEffects,
				Idempotent:      traits.Idempotent,
				SupportsDryRun:  traits.SupportsDryRun || hasFlagRecursive(node, "dry-run"),
				Arguments:       extractArgs(node.Use),
				RecommendedFlow: recommendedFlowForPath(relPath),
			}
			contracts = append(contracts, contract)
		}
		for _, child := range node.Commands() {
			walk(child)
		}
	}
	walk(root)
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].Path < contracts[j].Path })
	return contracts
}

func printAgentHelpText(payload agentHelpPayload) {
	fmt.Printf("tool: %s\n", payload.Tool)
	fmt.Println("guidance:")
	for _, g := range payload.Guidance {
		fmt.Printf("- %s\n", g)
	}
	if len(payload.Skills) > 0 {
		for _, skill := range payload.Skills {
			fmt.Printf("\nskill: %s\n", skill.Name)
			fmt.Printf("purpose: %s\n", skill.Purpose)
			for i, step := range skill.Steps {
				fmt.Printf("%d. %s\n", i+1, step)
			}
		}
	}
	fmt.Println("\ncommands:")
	for _, c := range payload.Commands {
		fmt.Printf("- %s | intent=%s side_effects=%s idempotent=%t dry_run=%t\n", c.Path, c.Intent, c.SideEffects, c.Idempotent, c.SupportsDryRun)
	}
}

func extractArgs(use string) []string {
	parts := strings.Fields(use)
	args := []string{}
	for _, p := range parts {
		if strings.HasPrefix(p, "<") || strings.HasPrefix(p, "[") {
			args = append(args, p)
		}
	}
	return args
}

func hasFlagRecursive(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}
	if cmd.Flags().Lookup(name) != nil {
		return true
	}
	if cmd.PersistentFlags().Lookup(name) != nil {
		return true
	}
	return false
}

func recommendedFlowForPath(path string) []string {
	switch path {
	case "note create", "note append", "note prepend", "note move", "note delete", "template insert", "prop set", "prop delete", "daily append", "daily prepend", "task", "block set":
		return []string{"run with --dry-run", "apply mutation", "verify with read command"}
	default:
		return nil
	}
}

func skillByName(name string) agentSkill {
	key := strings.ToLower(strings.TrimSpace(name))
	switch key {
	case "note-lifecycle":
		return agentSkill{
			Name:    "note-lifecycle",
			Purpose: "Create, edit, move, and validate note changes safely.",
			Steps: []string{
				"Discover candidates with list/note list/search.",
				"Preview mutation with --dry-run when available.",
				"Apply write command and re-read using note get or search.",
			},
		}
	case "task-management":
		return agentSkill{
			Name:    "task-management",
			Purpose: "Read and update checkbox tasks with explicit references.",
			Steps: []string{
				"Locate tasks with tasks --path or tasks --daily.",
				"Use task --ref path:line for precise updates.",
				"Verify resulting status with task --ref.",
			},
		}
	default:
		return agentSkill{
			Name:    key,
			Purpose: "General obsidian-cli automation skill.",
			Steps: []string{
				"Use schema to discover command contracts.",
				"Prefer --json and parse structured output.",
				"Apply writes only after a dry-run/precondition step.",
			},
		}
	}
}
