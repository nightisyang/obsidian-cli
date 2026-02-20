package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTagSearchCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <tag>",
		Short: "Find notes by tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			results, err := rt.Backend.SearchTag(rt.Context, args[0], limit)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(results)
			}
			for _, result := range results {
				rt.Printer.Println(fmt.Sprintf("%s\t%s", result.Path, result.Match))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Result limit")
	return cmd
}
