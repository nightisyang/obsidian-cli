package cmd

import (
	"github.com/spf13/cobra"
)

func newNoteGetCmd() *cobra.Command {
	var heading string
	cmd := &cobra.Command{
		Use:   "get <path>",
		Short: "Get a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if heading != "" {
				section, err := rt.Backend.GetHeading(rt.Context, args[0], heading)
				if err != nil {
					return err
				}
				if rt.Printer.JSON {
					return rt.Printer.PrintJSON(section)
				}
				rt.Printer.Println(section.Content)
				return nil
			}
			n, err := rt.Backend.GetNote(rt.Context, args[0])
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(n)
			}
			rt.Printer.Println(n.Raw)
			return nil
		},
	}
	cmd.Flags().StringVar(&heading, "heading", "", "Read only a specific heading section")
	return cmd
}
