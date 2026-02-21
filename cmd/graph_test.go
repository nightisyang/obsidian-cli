package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGraphNeighborhoodJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.md"), []byte("[[b]]\nalpha"), 0o644); err != nil {
		t.Fatalf("write a: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.md"), []byte("beta"), 0o644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	stdout, stderr, err := runCLI(t, "--vault", root, "--json", "graph", "neighborhood", "a.md", "--depth", "1")
	if err != nil {
		t.Fatalf("graph neighborhood failed: %v (stderr=%q)", err, stderr)
	}
	env := parseEnvelope(t, stdout)
	var payload struct {
		Nodes []struct {
			Path string `json:"path"`
		} `json:"nodes"`
		Metadata struct {
			GeneratedAt string `json:"generated_at"`
		} `json:"metadata"`
		Edges []struct {
			From string `json:"from"`
			To   string `json:"to"`
			Kind string `json:"kind"`
		} `json:"edges"`
	}
	if decodeErr := json.Unmarshal(env.Data, &payload); decodeErr != nil {
		t.Fatalf("decode graph payload: %v", decodeErr)
	}
	if len(payload.Nodes) < 2 {
		t.Fatalf("expected at least 2 nodes, got %+v", payload.Nodes)
	}
	if payload.Metadata.GeneratedAt == "" {
		t.Fatalf("expected metadata.generated_at in graph payload")
	}
	found := false
	for _, edge := range payload.Edges {
		if edge.From == "a.md" && edge.To == "b.md" && edge.Kind == "links_to" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a.md -> b.md links_to edge, got %+v", payload.Edges)
	}
}

func TestGraphContextJSON(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "seed.md"), []byte("[[neighbor]]\nimportant-query-token"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "neighbor.md"), []byte("related"), 0o644); err != nil {
		t.Fatalf("write neighbor: %v", err)
	}

	stdout, stderr, err := runCLI(
		t,
		"--vault", root,
		"--json",
		"graph", "context", "important-query-token",
		"--seed-limit", "3",
		"--depth", "1",
	)
	if err != nil {
		t.Fatalf("graph context failed: %v (stderr=%q)", err, stderr)
	}
	env := parseEnvelope(t, stdout)
	var payload struct {
		Query string `json:"query"`
		Seeds []struct {
			Path string `json:"path"`
		} `json:"seeds"`
		Nodes []struct {
			Path   string `json:"path"`
			IsSeed bool   `json:"is_seed"`
		} `json:"nodes"`
		Metadata struct {
			CacheStatus string `json:"cache_status"`
		} `json:"metadata"`
	}
	if decodeErr := json.Unmarshal(env.Data, &payload); decodeErr != nil {
		t.Fatalf("decode graph context payload: %v", decodeErr)
	}
	if payload.Query != "important-query-token" {
		t.Fatalf("unexpected query in payload: %+v", payload)
	}
	if len(payload.Seeds) == 0 {
		t.Fatalf("expected at least one seed result")
	}
	hasSeedNode := false
	for _, node := range payload.Nodes {
		if node.Path == "seed.md" && node.IsSeed {
			hasSeedNode = true
			break
		}
	}
	if !hasSeedNode {
		t.Fatalf("expected seed.md as seed node, got %+v", payload.Nodes)
	}
	if payload.Metadata.CacheStatus == "" {
		t.Fatalf("expected metadata.cache_status in graph context payload")
	}
}
