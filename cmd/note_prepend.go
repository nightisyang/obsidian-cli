package cmd

import "github.com/spf13/cobra"

func newNotePrependCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepend <path> <content>",
		Short: "Prepend content to a note body",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			n, err := rt.Backend.PrependNote(rt.Context, args[0], args[1])
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
	return cmd
}
