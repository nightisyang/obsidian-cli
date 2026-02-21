package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type commandSchema struct {
	Path           string       `json:"path"`
	Use            string       `json:"use"`
	Short          string       `json:"short"`
	Intent         string       `json:"intent"`
	SideEffects    string       `json:"side_effects"`
	Idempotent     bool         `json:"idempotent"`
	SupportsDryRun bool         `json:"supports_dry_run"`
	Flags          []flagSchema `json:"flags,omitempty"`
	Arguments      []string     `json:"arguments,omitempty"`
	Subcommands    []string     `json:"subcommands,omitempty"`
}

type flagSchema struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage"`
}

func newSchemaCmd(root *cobra.Command) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "schema [command path]",
		Short: "Export machine-readable command schema",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			schemas := []commandSchema{}
			if len(args) == 0 {
				schemas = collectSchemas(root)
			} else {
				target, _, err := root.Find(args)
				if err != nil {
					return err
				}
				schemas = append(schemas, schemaForCommand(target))
			}

			switch strings.ToLower(strings.TrimSpace(format)) {
			case "json", "":
				return output.WriteJSON(os.Stdout, map[string]any{
					"format_version": "schema.v1",
					"tool":           "obsidian-cli",
					"commands":       schemas,
				})
			case "text":
				for _, schema := range schemas {
					fmt.Printf("%s\n", schema.Path)
					for _, flag := range schema.Flags {
						fmt.Printf("  --%s (%s)\n", flag.Name, flag.Type)
					}
				}
				return nil
			default:
				return fmt.Errorf("unsupported --format %q, use text or json", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "Output format: json|text")
	return cmd
}

func collectSchemas(root *cobra.Command) []commandSchema {
	items := []commandSchema{}
	var walk func(node *cobra.Command)
	walk = func(node *cobra.Command) {
		if shouldSkipFromCatalog(node) {
			return
		}
		if rel := relativeCommandPath(node); rel != "" {
			items = append(items, schemaForCommand(node))
		}
		for _, child := range node.Commands() {
			walk(child)
		}
	}
	walk(root)
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items
}

func schemaForCommand(cmd *cobra.Command) commandSchema {
	traits := traitForCommand(cmd)
	flags := []flagSchema{}
	cmd.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		required := false
		if values, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(values) > 0 {
			required = true
		}
		flags = append(flags, flagSchema{
			Name:      flag.Name,
			Shorthand: flag.Shorthand,
			Type:      flag.Value.Type(),
			Required:  required,
			Default:   flag.DefValue,
			Usage:     flag.Usage,
		})
	})
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })

	subcommands := []string{}
	for _, child := range cmd.Commands() {
		if shouldSkipFromCatalog(child) {
			continue
		}
		subcommands = append(subcommands, child.Name())
	}
	sort.Strings(subcommands)

	return commandSchema{
		Path:           relativeCommandPath(cmd),
		Use:            cmd.UseLine(),
		Short:          cmd.Short,
		Intent:         traits.Intent,
		SideEffects:    traits.SideEffects,
		Idempotent:     traits.Idempotent,
		SupportsDryRun: traits.SupportsDryRun || hasFlagRecursive(cmd, "dry-run"),
		Flags:          flags,
		Arguments:      extractArgs(cmd.Use),
		Subcommands:    subcommands,
	}
}
