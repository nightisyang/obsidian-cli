package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const agentWorkflowHelp = `Canonical workflow for agents:
- orient: note find --kind X --status active
- load: note get <path> --max-chars N
- write: note create "Title" --kind X --content "..." then note append
- update: prop set <path> <key> <value>
- batch: ops apply <file.json>
- discover: schema`

func newAgentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "agent",
		Short: "Print canonical workflow for agents",
		Long:  agentWorkflowHelp,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), agentWorkflowHelp)
			return err
		},
	}
}
