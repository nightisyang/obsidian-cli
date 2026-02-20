package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLinksListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <path>",
		Short: "List outgoing wikilinks in a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			links, err := rt.Backend.OutgoingLinks(rt.Context, args[0])
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(links)
			}
			for _, link := range links {
				rt.Printer.Println(fmt.Sprintf("%s", link))
			}
			return nil
		},
	}
	return cmd
}
