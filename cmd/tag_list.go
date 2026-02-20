package cmd

import (
	"fmt"

	"github.com/nightisyang/obsidian-cli/internal/index"
	"github.com/spf13/cobra"
)

func newTagListCmd() *cobra.Command {
	var limit int
	var sortBy string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tags",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			tags, err := rt.Backend.ListTags(rt.Context, index.TagListOptions{Limit: limit, Sort: sortBy})
			if err != nil {
				return err
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(tags)
			}
			for _, tag := range tags {
				rt.Printer.Println(fmt.Sprintf("%s\t%d", tag.Tag, tag.Count))
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit tags")
	cmd.Flags().StringVar(&sortBy, "sort", "count", "Sort by count or name")
	return cmd
}
