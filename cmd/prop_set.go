package cmd

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newPropSetCmd() *cobra.Command {
	var dryRun bool
	var ifHash string
	cmd := &cobra.Command{
		Use:   "set <path> <key> <value>",
		Short: "Set a frontmatter property",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if err := verifyHashPrecondition(rt, args[0], ifHash); err != nil {
				return err
			}
			value := parseLiteral(args[2])
			if dryRun {
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(map[string]any{
						"dry_run": true,
						"action":  "prop.set",
						"path":    args[0],
						"key":     args[1],
						"value":   value,
					})
				}
				rt.Printer.Println("dry-run: would set property " + args[1] + " on " + args[0])
				return nil
			}
			n, err := rt.Backend.PropSet(rt.Context, args[0], args[1], value)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println("updated: " + n.Path)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview operation without writing files")
	cmd.Flags().StringVar(&ifHash, "if-hash", "", "Require current note SHA256 hash before writing")
	return cmd
}

func parseLiteral(raw string) any {
	value := strings.TrimSpace(raw)
	if value == "true" || value == "false" {
		return value == "true"
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		inside := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "["), "]"))
		if inside == "" {
			return []string{}
		}
		parts := strings.Split(inside, ",")
		items := make([]string, 0, len(parts))
		for _, p := range parts {
			items = append(items, strings.TrimSpace(p))
		}
		return items
	}
	return raw
}
