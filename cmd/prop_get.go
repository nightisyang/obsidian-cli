package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPropGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <path> <key>",
		Short: "Get a frontmatter property",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			value, err := rt.Backend.PropGet(rt.Context, args[0], args[1])
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(map[string]any{"path": args[0], "key": args[1], "value": value})
			}
			rt.Printer.Println(fmt.Sprintf("%v", value))
			return nil
		},
	}
	return cmd
}
