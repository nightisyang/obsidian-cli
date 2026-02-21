package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelpAgentJSON(t *testing.T) {
	stdout, stderr, err := runCLI(t, "help", "--agent", "--format", "json")
	if err != nil {
		t.Fatalf("help --agent failed: %v (stderr=%q)", err, stderr)
	}

	var payload struct {
		Tool     string `json:"tool"`
		Commands []struct {
			Path string `json:"path"`
		} `json:"commands"`
	}
	if decodeErr := json.Unmarshal([]byte(stdout), &payload); decodeErr != nil {
		t.Fatalf("decode help payload: %v; raw=%q", decodeErr, stdout)
	}
	if payload.Tool != "obsidian-cli" {
		t.Fatalf("unexpected tool in help payload: %+v", payload)
	}
	if len(payload.Commands) == 0 {
		t.Fatalf("expected command contracts in help payload")
	}
}

func TestSchemaCommandIncludesSearchContent(t *testing.T) {
	stdout, stderr, err := runCLI(t, "schema", "--format", "json", "search-content")
	if err != nil {
		t.Fatalf("schema failed: %v (stderr=%q)", err, stderr)
	}
	var payload struct {
		Commands []struct {
			Path string `json:"path"`
		} `json:"commands"`
	}
	if decodeErr := json.Unmarshal([]byte(stdout), &payload); decodeErr != nil {
		t.Fatalf("decode schema payload: %v; raw=%q", decodeErr, stdout)
	}
	if len(payload.Commands) != 1 || payload.Commands[0].Path != "search-content" {
		t.Fatalf("unexpected schema payload: %+v", payload)
	}
}

func TestPropDeleteDryRunAndApply(t *testing.T) {
	root := t.TempDir()
	notePath := filepath.Join(root, "alpha.md")
	content := `---
status: active
topic: infra
---

hello`
	if err := os.WriteFile(notePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write note: %v", err)
	}

	stdout, stderr, err := runCLI(t, "--vault", root, "--json", "prop", "delete", "alpha.md", "status", "--dry-run")
	if err != nil {
		t.Fatalf("prop delete dry-run failed: %v (stderr=%q)", err, stderr)
	}
	env := parseEnvelope(t, stdout)
	var payload map[string]any
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode prop dry-run payload: %v", err)
	}
	if payload["dry_run"] != true {
		t.Fatalf("expected dry_run response, got %v", payload)
	}
	afterDryRun, _ := os.ReadFile(notePath)
	if !strings.Contains(string(afterDryRun), "status: active") {
		t.Fatalf("dry-run should not mutate file")
	}

	_, stderr, err = runCLI(t, "--vault", root, "prop", "delete", "alpha.md", "status")
	if err != nil {
		t.Fatalf("prop delete apply failed: %v (stderr=%q)", err, stderr)
	}
	afterApply, _ := os.ReadFile(notePath)
	if strings.Contains(string(afterApply), "status: active") {
		t.Fatalf("status key should be removed after delete")
	}
}

func TestJSONFailureEnvelopeIncludesReasonAndHint(t *testing.T) {
	root := t.TempDir()

	oldArgs := os.Args
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = outWriter
	os.Stderr = errWriter
	os.Args = []string{"obsidian-cli", "--vault", root, "--json", "note", "get", "missing.md"}

	code := Execute()

	_ = outWriter.Close()
	_ = errWriter.Close()
	_, _ = io.ReadAll(outReader)
	stderrBytes, _ := io.ReadAll(errReader)

	os.Args = oldArgs
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if code == 0 {
		t.Fatalf("expected note get missing to fail")
	}
	stderr := string(stderrBytes)

	var env struct {
		OK    bool `json:"ok"`
		Error struct {
			Code           int    `json:"code"`
			Reason         string `json:"reason"`
			Message        string `json:"message"`
			ActionableHint string `json:"actionable_hint"`
		} `json:"error"`
	}
	if decodeErr := json.Unmarshal([]byte(stderr), &env); decodeErr != nil {
		t.Fatalf("decode stderr envelope: %v; raw=%q", decodeErr, stderr)
	}
	if env.OK {
		t.Fatalf("expected error envelope, got ok")
	}
	if env.Error.Reason == "" || env.Error.ActionableHint == "" {
		t.Fatalf("expected reason and actionable_hint in error payload: %+v raw=%s", env.Error, stderr)
	}
	if env.Error.Code == 0 {
		t.Fatalf("expected non-zero error code: %+v", env.Error)
	}
	if env.Error.Message == "" {
		t.Fatalf("expected non-empty error message: %+v", env.Error)
	}
}

func TestCheckExpectations(t *testing.T) {
	payload := map[string]any{
		"ok": true,
		"data": map[string]any{
			"path": "alpha.md",
		},
	}
	ok, msg := checkExpectations(payload, map[string]string{
		"ok":        "true",
		"data.path": "alpha.md",
	})
	if !ok || msg != "" {
		t.Fatalf("expected expectation success, got ok=%t msg=%q", ok, msg)
	}

	ok, msg = checkExpectations(payload, map[string]string{"data.path": "beta.md"})
	if ok || msg == "" {
		t.Fatalf("expected expectation failure with message, got ok=%t msg=%q", ok, msg)
	}
}
