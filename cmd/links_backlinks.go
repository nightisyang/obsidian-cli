package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLinksBacklinksCmd() *cobra.Command {
	var rebuild bool

	cmd := &cobra.Command{
		Use:   "backlinks <path>",
		Short: "List backlinks for a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			backlinks, err := rt.Backend.Backlinks(rt.Context, args[0], rebuild)
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(backlinks)
			}
			for _, path := range backlinks {
				rt.Printer.Println(fmt.Sprintf("%s", path))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&rebuild, "index", false, "Force index rebuild")
	return cmd
}
