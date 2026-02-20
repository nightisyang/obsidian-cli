package cmd

import "github.com/spf13/cobra"

func newNoteAppendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "append <path> <content>",
		Short: "Append content to a note",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			n, err := rt.Backend.AppendNote(rt.Context, args[0], args[1])
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
