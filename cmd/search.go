package cmd

import (
	"fmt"
	"strings"

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
	var withMeta bool
	var strict bool

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search notes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := ""
			if len(args) == 1 {
				text = args[0]
			}
			return runSearch(cmd, text, tag, prop, limit, contextChars, pathPrefix, caseSensitive, maxChars, withMeta, strict)
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "Search by tag")
	cmd.Flags().StringVar(&prop, "prop", "", "Search by property key=value")
	cmd.Flags().IntVar(&limit, "limit", 20, "Result limit")
	cmd.Flags().IntVar(&contextChars, "context", 80, "Snippet context chars")
	cmd.Flags().IntVar(&maxChars, "max-chars", 0, "Maximum snippet chars per result")
	cmd.Flags().StringVar(&pathPrefix, "path", "", "Restrict to path prefix")
	cmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case sensitive text search")
	cmd.Flags().BoolVar(&withMeta, "with-meta", false, "Include metadata and warnings in output")
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail when warnings are present (for agent guardrails)")
	return cmd
}

func newSearchContentCmd() *cobra.Command {
	var limit int
	var contextChars int
	var pathPrefix string
	var caseSensitive bool
	var maxChars int
	var withMeta bool
	var strict bool

	cmd := &cobra.Command{
		Use:    "search-content <query>",
		Short:  "Search note content for text",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, args[0], "", "", limit, contextChars, pathPrefix, caseSensitive, maxChars, withMeta, strict)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Result limit")
	cmd.Flags().IntVar(&contextChars, "context", 80, "Snippet context chars")
	cmd.Flags().IntVar(&maxChars, "max-chars", 0, "Maximum snippet chars per result")
	cmd.Flags().StringVar(&pathPrefix, "path", "", "Restrict to path prefix")
	cmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case sensitive text search")
	cmd.Flags().BoolVar(&withMeta, "with-meta", false, "Include metadata and warnings in output")
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail when warnings are present (for agent guardrails)")
	return cmd
}

func runSearch(cmd *cobra.Command, text, tag, prop string, limit, contextChars int, pathPrefix string, caseSensitive bool, maxChars int, withMeta bool, strict bool) error {
	if cmd.Flags().Changed("max-chars") && maxChars <= 0 {
		return errs.New(errs.ExitValidation, "--max-chars must be > 0")
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
	metadata := newOperationMetadata(strict)
	metadata.CacheStatus = "on_demand"
	if max, err := sourceFileMTimeMax(rt.VaultRoot); err == nil {
		metadata.SourceFileMTimeMax = max
	}
	warnings := []string{}
	if q.Limit > 0 && len(results) >= q.Limit {
		metadata.Truncated = true
		warnings = append(warnings, "result limit reached; output may be truncated")
	}
	if strict && len(warnings) > 0 {
		return errs.NewDetailed(
			errs.ExitValidation,
			"strict_mode_violation",
			"Increase limits or disable --strict when truncation is acceptable.",
			strings.Join(warnings, "; "),
		)
	}
	if rt.Printer.JSON {
		if withMeta {
			return rt.Printer.PrintJSON(map[string]any{
				"results":  results,
				"metadata": metadata,
				"warnings": warnings,
			})
		}
		return rt.Printer.PrintJSON(results)
	}
	for _, result := range results {
		if result.Line > 0 {
			rt.Printer.Println(fmt.Sprintf("%s:%d\t%s", result.Path, result.Line, result.Snippet))
		} else {
			rt.Printer.Println(fmt.Sprintf("%s\t%s", result.Path, result.Match))
		}
	}
	if withMeta {
		rt.Printer.Println(fmt.Sprintf("metadata: generated_at=%s cache_status=%s truncated=%t", metadata.GeneratedAt, metadata.CacheStatus, metadata.Truncated))
		if metadata.SourceFileMTimeMax != "" {
			rt.Printer.Println("source_file_mtime_max: " + metadata.SourceFileMTimeMax)
		}
		for _, warning := range warnings {
			rt.Printer.Println("warning: " + warning)
		}
	}
	return nil
}
