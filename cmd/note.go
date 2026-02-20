package cmd

import "github.com/spf13/cobra"

func newNoteCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "note", Short: "Note operations"}
	cmd.AddCommand(newNoteCreateCmd())
	cmd.AddCommand(newNoteGetCmd())
	cmd.AddCommand(newNoteListCmd())
	cmd.AddCommand(newNoteDeleteCmd())
	cmd.AddCommand(newNoteAppendCmd())
	cmd.AddCommand(newNotePrependCmd())
	cmd.AddCommand(newNoteMoveCmd())
	return cmd
}
