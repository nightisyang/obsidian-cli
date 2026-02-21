package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/app"
	"github.com/nightisyang/obsidian-cli/internal/errs"
	"github.com/nightisyang/obsidian-cli/internal/search"
	"github.com/spf13/cobra"
)

type graphNode struct {
	Path      string `json:"path"`
	IsSeed    bool   `json:"is_seed"`
	Score     int    `json:"score"`
	InDegree  int    `json:"in_degree"`
	OutDegree int    `json:"out_degree"`
}

type graphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type graphContextPack struct {
	Query    string                `json:"query,omitempty"`
	Seeds    []search.SearchResult `json:"seeds,omitempty"`
	Nodes    []graphNode           `json:"nodes"`
	Edges    []graphEdge           `json:"edges"`
	Metadata operationMetadata     `json:"metadata"`
	Warnings []string              `json:"warnings,omitempty"`
}

type queueItem struct {
	Path  string
	Depth int
}

func newGraphCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Graph retrieval operations",
	}
	cmd.AddCommand(newGraphContextCmd())
	cmd.AddCommand(newGraphNeighborhoodCmd())
	return cmd
}

func newGraphContextCmd() *cobra.Command {
	var seedLimit int
	var depth int
	var nodeLimit int
	var pathPrefix string
	var caseSensitive bool
	var strict bool

	cmd := &cobra.Command{
		Use:   "context <query>",
		Short: "Build a query-seeded context graph from note relationships",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if seedLimit <= 0 {
				seedLimit = 5
			}
			if depth < 0 {
				depth = 0
			}
			if nodeLimit <= 0 {
				nodeLimit = 200
			}

			q, err := search.BuildQuery(args[0], "", "", seedLimit, 80, pathPrefix, caseSensitive)
			if err != nil {
				return err
			}
			seeds, err := rt.Backend.Search(rt.Context, q)
			if err != nil {
				return err
			}
			seedPaths := uniqueSeedPaths(seeds)
			nodes, edges, warnings, truncated, err := expandNeighborhood(rt, seedPaths, depth, nodeLimit, strict)
			if err != nil {
				return err
			}
			metadata := newOperationMetadata(strict)
			metadata.CacheStatus = "backlinks_in_memory_auto"
			metadata.Truncated = truncated
			if max, err := sourceFileMTimeMax(rt.VaultRoot); err == nil {
				metadata.SourceFileMTimeMax = max
			}
			if strict && len(warnings) > 0 {
				return errs.NewDetailed(
					errs.ExitValidation,
					"strict_mode_violation",
					"Increase graph limits, reduce depth, or disable --strict.",
					strings.Join(warnings, "; "),
				)
			}

			payload := graphContextPack{
				Query:    args[0],
				Seeds:    seeds,
				Nodes:    nodes,
				Edges:    edges,
				Metadata: metadata,
				Warnings: warnings,
			}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(payload)
			}
			rt.Printer.Println(fmt.Sprintf("seeds: %d nodes: %d edges: %d", len(seedPaths), len(nodes), len(edges)))
			for _, warning := range warnings {
				rt.Printer.Println("warning: " + warning)
			}
			for _, n := range nodes {
				if n.IsSeed {
					rt.Printer.Println("seed: " + n.Path)
				}
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&seedLimit, "seed-limit", 5, "Max number of search seed notes")
	cmd.Flags().IntVar(&depth, "depth", 1, "Link expansion depth")
	cmd.Flags().IntVar(&nodeLimit, "node-limit", 200, "Max graph nodes")
	cmd.Flags().StringVar(&pathPrefix, "path", "", "Restrict seed search to a path prefix")
	cmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "Case sensitive text search")
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail when warnings are present (for agent guardrails)")
	return cmd
}

func newGraphNeighborhoodCmd() *cobra.Command {
	var depth int
	var nodeLimit int
	var strict bool

	cmd := &cobra.Command{
		Use:   "neighborhood <path>",
		Short: "Expand outgoing links and backlinks around a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := getRuntime(cmd)
			if err != nil {
				return err
			}
			if depth < 0 {
				depth = 0
			}
			if nodeLimit <= 0 {
				nodeLimit = 200
			}
			nodes, edges, warnings, truncated, err := expandNeighborhood(rt, []string{normalizeGraphPath(args[0])}, depth, nodeLimit, strict)
			if err != nil {
				return err
			}
			metadata := newOperationMetadata(strict)
			metadata.CacheStatus = "backlinks_in_memory_auto"
			metadata.Truncated = truncated
			if max, err := sourceFileMTimeMax(rt.VaultRoot); err == nil {
				metadata.SourceFileMTimeMax = max
			}
			if strict && len(warnings) > 0 {
				return errs.NewDetailed(
					errs.ExitValidation,
					"strict_mode_violation",
					"Increase graph limits, reduce depth, or disable --strict.",
					strings.Join(warnings, "; "),
				)
			}
			payload := graphContextPack{Nodes: nodes, Edges: edges, Metadata: metadata, Warnings: warnings}
			if rt.Printer.JSON {
				return rt.Printer.PrintJSON(payload)
			}
			rt.Printer.Println(fmt.Sprintf("nodes: %d edges: %d", len(nodes), len(edges)))
			for _, warning := range warnings {
				rt.Printer.Println("warning: " + warning)
			}
			for _, edge := range edges {
				rt.Printer.Println(fmt.Sprintf("%s -> %s (%s)", edge.From, edge.To, edge.Kind))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&depth, "depth", 1, "Link expansion depth")
	cmd.Flags().IntVar(&nodeLimit, "node-limit", 200, "Max graph nodes")
	cmd.Flags().BoolVar(&strict, "strict", false, "Fail when warnings are present (for agent guardrails)")
	return cmd
}

func uniqueSeedPaths(results []search.SearchResult) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(results))
	for _, r := range results {
		path := normalizeGraphPath(r.Path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	return out
}

func normalizeGraphPath(raw string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	trimmed = strings.TrimPrefix(trimmed, "/")
	if trimmed == "" || trimmed == "." {
		return ""
	}
	if !strings.HasSuffix(strings.ToLower(trimmed), ".md") {
		trimmed += ".md"
	}
	return filepath.ToSlash(filepath.Clean(trimmed))
}

func expandNeighborhood(rt *app.Runtime, seeds []string, maxDepth int, nodeLimit int, strict bool) ([]graphNode, []graphEdge, []string, bool, error) {
	nodes := map[string]*graphNode{}
	edges := []graphEdge{}
	edgeSeen := map[string]struct{}{}
	discoveredDepth := map[string]int{}
	queue := []queueItem{}
	noteExists := map[string]bool{}
	warningSet := map[string]struct{}{}
	truncated := false

	addWarning := func(message string) {
		if strings.TrimSpace(message) == "" {
			return
		}
		warningSet[message] = struct{}{}
	}

	enqueue := func(path string, depth int) {
		if depth > maxDepth || path == "" {
			return
		}
		if existing, ok := discoveredDepth[path]; ok && existing <= depth {
			return
		}
		discoveredDepth[path] = depth
		queue = append(queue, queueItem{Path: path, Depth: depth})
	}

	addNode := func(path string, isSeed bool) bool {
		if path == "" {
			return false
		}
		if existing, ok := nodes[path]; ok {
			if isSeed {
				existing.IsSeed = true
				existing.Score += 10
			}
			return true
		}
		if nodeLimit > 0 && len(nodes) >= nodeLimit {
			truncated = true
			addWarning("node limit reached; graph truncated")
			return false
		}
		score := 0
		if isSeed {
			score = 10
		}
		nodes[path] = &graphNode{
			Path:   path,
			IsSeed: isSeed,
			Score:  score,
		}
		return true
	}

	addEdge := func(from, to, kind string) bool {
		key := from + "|" + kind + "|" + to
		if _, ok := edgeSeen[key]; ok {
			return false
		}
		edgeSeen[key] = struct{}{}
		edges = append(edges, graphEdge{From: from, To: to, Kind: kind})
		if src, ok := nodes[from]; ok {
			src.OutDegree++
			src.Score++
		}
		if dst, ok := nodes[to]; ok {
			dst.InDegree++
			dst.Score++
		}
		return true
	}

	hasNote := func(path string) (bool, error) {
		if path == "" {
			return false, nil
		}
		if exists, ok := noteExists[path]; ok {
			return exists, nil
		}
		_, err := rt.Backend.GetNote(rt.Context, path)
		if err != nil {
			var appErr *errs.AppError
			if errors.As(err, &appErr) && appErr.Code == errs.ExitNotFound {
				noteExists[path] = false
				return false, nil
			}
			return false, err
		}
		exists := true
		noteExists[path] = exists
		return exists, nil
	}

	for _, seed := range seeds {
		path := normalizeGraphPath(seed)
		if addNode(path, true) {
			enqueue(path, 0)
		}
	}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		if item.Depth >= maxDepth {
			continue
		}
		exists, err := hasNote(item.Path)
		if err != nil {
			addWarning(fmt.Sprintf("failed to read note %s: %v", item.Path, err))
			if strict {
				return nil, nil, nil, truncated, err
			}
			continue
		}
		if !exists {
			continue
		}

		outgoing, err := rt.Backend.OutgoingLinks(rt.Context, item.Path)
		if err != nil {
			addWarning(fmt.Sprintf("failed to list outgoing links for %s: %v", item.Path, err))
			if strict {
				return nil, nil, nil, truncated, err
			}
		} else {
			for _, raw := range outgoing {
				target := normalizeGraphPath(raw)
				if !addNode(target, false) {
					continue
				}
				addEdge(item.Path, target, "links_to")
				targetExists, targetErr := hasNote(target)
				if targetErr != nil {
					addWarning(fmt.Sprintf("failed to read linked note %s: %v", target, targetErr))
					if strict {
						return nil, nil, nil, truncated, targetErr
					}
					continue
				}
				if targetExists {
					enqueue(target, item.Depth+1)
				}
			}
		}

		backlinks, err := rt.Backend.Backlinks(rt.Context, item.Path, false)
		if err != nil {
			addWarning(fmt.Sprintf("failed to list backlinks for %s: %v", item.Path, err))
			if strict {
				return nil, nil, nil, truncated, err
			}
		} else {
			for _, raw := range backlinks {
				source := normalizeGraphPath(raw)
				if !addNode(source, false) {
					continue
				}
				addEdge(source, item.Path, "linked_to")
				enqueue(source, item.Depth+1)
			}
		}
	}

	nodeList := make([]graphNode, 0, len(nodes))
	for _, n := range nodes {
		nodeList = append(nodeList, *n)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		if nodeList[i].Score == nodeList[j].Score {
			return nodeList[i].Path < nodeList[j].Path
		}
		return nodeList[i].Score > nodeList[j].Score
	})
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			if edges[i].To == edges[j].To {
				return edges[i].Kind < edges[j].Kind
			}
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})
	warnings := make([]string, 0, len(warningSet))
	for warning := range warningSet {
		warnings = append(warnings, warning)
	}
	sort.Strings(warnings)
	return nodeList, edges, warnings, truncated, nil
}
