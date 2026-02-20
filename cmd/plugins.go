package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPluginsCmd() *cobra.Command {
	var filter string
	var enabledOnly bool
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			items, err := rt.Backend.ListPlugins(rt.Context, filter, enabledOnly)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(items)
			}
			for _, item := range items {
				rt.Printer.Println(fmt.Sprintf("%s\t%s\t%s\tenabled=%t", item.Type, item.ID, item.Version, item.Enabled))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Filter by plugin type: core|community")
	cmd.Flags().BoolVar(&enabledOnly, "enabled", false, "Show only enabled plugins")
	return cmd
}

func newCommandsCmd() *cobra.Command {
	var filter string
	cmd := &cobra.Command{
		Use:   "commands",
		Short: "List known command IDs (from hotkeys map in native mode)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			ids, err := rt.Backend.ListCommandIDs(rt.Context, filter)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(ids)
			}
			for _, id := range ids {
				rt.Printer.Println(id)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Filter by command ID prefix")
	return cmd
}

func newCommandCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command <id>",
		Short: "Execute an Obsidian command (requires app/API mode)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if err := rt.Backend.ExecuteCommand(rt.Context, args[0]); err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{"executed": args[0]})
			}
			rt.Printer.Println("executed: " + args[0])
			return nil
		},
	}
	return cmd
}
