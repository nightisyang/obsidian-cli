package cmd

import "github.com/spf13/cobra"

type commandTrait struct {
	Intent         string
	SideEffects    string
	Idempotent     bool
	Mutating       bool
	SupportsDryRun bool
}

var commandTraits = map[string]commandTrait{
	"list":               {Intent: "discover", SideEffects: "none", Idempotent: true},
	"print-default":      {Intent: "discover", SideEffects: "none", Idempotent: true},
	"schema":             {Intent: "discover", SideEffects: "none", Idempotent: true},
	"search":             {Intent: "read", SideEffects: "none", Idempotent: true},
	"search-content":     {Intent: "read", SideEffects: "none", Idempotent: true},
	"graph":              {Intent: "read", SideEffects: "none", Idempotent: true},
	"graph context":      {Intent: "read", SideEffects: "none", Idempotent: true},
	"graph neighborhood": {Intent: "read", SideEffects: "none", Idempotent: true},
	"note get":           {Intent: "read", SideEffects: "none", Idempotent: true},
	"note list":          {Intent: "discover", SideEffects: "none", Idempotent: true},
	"note find":          {Intent: "discover", SideEffects: "none", Idempotent: true},
	"daily":              {Intent: "mutate", SideEffects: "writes", Mutating: true},
	"daily read":         {Intent: "read", SideEffects: "none", Idempotent: true},
	"daily path":         {Intent: "discover", SideEffects: "none", Idempotent: true},
	"daily append":       {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"daily prepend":      {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"note create":        {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"note append":        {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"note prepend":       {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"note delete":        {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"note move":          {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"prop set":           {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"prop delete":        {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"prop get":           {Intent: "read", SideEffects: "none", Idempotent: true},
	"prop list":          {Intent: "read", SideEffects: "none", Idempotent: true},
	"tag list":           {Intent: "discover", SideEffects: "none", Idempotent: true},
	"tag search":         {Intent: "discover", SideEffects: "none", Idempotent: true},
	"links list":         {Intent: "read", SideEffects: "none", Idempotent: true},
	"links backlinks":    {Intent: "read", SideEffects: "none", Idempotent: true},
	"task":               {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"tasks":              {Intent: "read", SideEffects: "none", Idempotent: true},
	"template read":      {Intent: "read", SideEffects: "none", Idempotent: true},
	"template insert":    {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"templates":          {Intent: "discover", SideEffects: "none", Idempotent: true},
	"block get":          {Intent: "read", SideEffects: "none", Idempotent: true},
	"block set":          {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"open":               {Intent: "external", SideEffects: "external", Idempotent: true},
	"sync status":        {Intent: "discover", SideEffects: "none", Idempotent: true},
	"plugins":            {Intent: "discover", SideEffects: "none", Idempotent: true},
	"commands":           {Intent: "discover", SideEffects: "none", Idempotent: true},
	"command":            {Intent: "external", SideEffects: "external", Mutating: true},
	"vault init":         {Intent: "mutate", SideEffects: "writes", Mutating: true},
	"vault migrate":      {Intent: "mutate", SideEffects: "writes", Mutating: true, SupportsDryRun: true},
	"vault status":       {Intent: "discover", SideEffects: "none", Idempotent: true},
	"ops apply":          {Intent: "mutate", SideEffects: "writes", Mutating: true},
}

func traitForCommand(cmd *cobra.Command) commandTrait {
	path := relativeCommandPath(cmd)
	if trait, ok := commandTraits[path]; ok {
		return trait
	}
	return commandTrait{
		Intent:      "read",
		SideEffects: "none",
		Idempotent:  true,
	}
}

func relativeCommandPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	path := cmd.CommandPath()
	parts := splitFields(path)
	if len(parts) <= 1 {
		return ""
	}
	return joinFields(parts[1:])
}

func splitFields(raw string) []string {
	out := []string{}
	current := ""
	for _, r := range raw {
		if r == ' ' || r == '\t' || r == '\n' {
			if current != "" {
				out = append(out, current)
				current = ""
			}
			continue
		}
		current += string(r)
	}
	if current != "" {
		out = append(out, current)
	}
	return out
}

func joinFields(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " " + parts[i]
	}
	return out
}

func shouldSkipFromCatalog(cmd *cobra.Command) bool {
	if cmd == nil {
		return true
	}
	if cmd.Hidden {
		return true
	}
	switch cmd.Name() {
	case "help", "completion":
		return true
	default:
		return false
	}
}
