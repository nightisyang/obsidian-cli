package cmd

import (
	"fmt"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/search"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var tag string
	var prop string
	var limit int
	var contextChars int
	var pathPrefix string
	var caseSensitive bool
	var maxChars int

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search notes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("max-chars") && maxChars <= 0 {
				return errs.New(errs.ExitValidation, "--max-chars must be > 0")
			}
			text := ""
			if len(args) == 1 {
				text = args[0]
			}
			q, err := search.BuildQuery(text, tag, prop, limit, contextChars, pathPrefix, caseSensitive)
			if err != nil {
				return err
			}
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			results, err := rt.Backend.Search(rt.Context, q)
			if err != nil {
				return err
			}
			if maxChars > 0 {
				results = applySearchSnippetMaxChars(results, maxChars)
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(results)
			}
			for _, result := range results {
				if result.Line > 0 {
					rt.Printer.Println(fmt.Sprintf("%s:%d\t%s", result.Path, result.Line, result.Snippet))
				} else {
					rt.Printer.Println(fmt.Sprintf("%s\t%s", result.Path, result.Match))
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Search by tag")
	cmd.Flags().StringVar(&prop, "prop", "", "Search by property key=value")
	cmd.Flags().IntVar(&limit, "limit", 20, "Result limit")
	cmd.Flags().IntVar(&contextChars, "context", 80, "Snippet context chars")
	cmd.Flags().IntVar(&maxChars, "max-chars", 0, "Maximum snippet chars per result")
	cmd.Flags().StringVar(&pathPrefix, "path", "", "Restrict to path prefix")
	cmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case sensitive text search")
	return cmd
}
