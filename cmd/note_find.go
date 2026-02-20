package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/note"
	"github.com/spf13/cobra"
)

func newNoteFindCmd() *cobra.Command {
	var kind string
	var tags []string
	var status string
	var topic string
	var since string

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find notes by metadata",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}

			sinceTime, err := parseSinceDate(since)
			if err != nil {
				return err
			}

			results, err := note.FindByMetadata(rt.VaultRoot, note.FindFilters{
				Kind:   kind,
				Tags:   tags,
				Status: status,
				Topic:  topic,
				Since:  sinceTime,
			})
			if err != nil {
				return err
			}

			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(results)
			}
			for _, result := range results {
				rt.Printer.Println(fmt.Sprintf(
					"%s: kind=%s tags=[%s] status=%s",
					result.Path,
					result.Kind,
					strings.Join(result.Tags, ","),
					result.Status,
				))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&kind, "kind", "", "Filter by frontmatter kind")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Filter by tag (repeatable, AND semantics)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by frontmatter status")
	cmd.Flags().StringVar(&topic, "topic", "", "Filter by frontmatter topic")
	cmd.Flags().StringVar(&since, "since", "", "Filter by created_at >= date (YYYY-MM-DD or RFC3339)")
	return cmd
}

func parseSinceDate(raw string) (*time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		u := t.UTC()
		return &u, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		u := t.UTC()
		return &u, nil
	}

	return nil, errs.New(errs.ExitValidation, "--since must be YYYY-MM-DD or RFC3339")
}
