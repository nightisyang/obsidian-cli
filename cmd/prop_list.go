package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newPropListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <path>",
		Short: "List frontmatter properties",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			props, err := rt.Backend.PropList(rt.Context, args[0])
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(props)
			}
			keys := make([]string, 0, len(props))
			for k := range props {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				rt.Printer.Println(fmt.Sprintf("%s: %v", k, props[k]))
			}
			return nil
		},
	}
	return cmd
}
