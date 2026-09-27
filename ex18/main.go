package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Decision string

const (
	Allow Decision = "allow"
	Ask   Decision = "ask"
	Deny  Decision = "deny"
)

func skillWriteDecision(workspace, target string) Decision {
	root, _ := filepath.Abs(workspace)
	path, _ := filepath.Abs(target)
	skills := filepath.Join(root, ".agent", "skills")
	if path == skills || strings.HasPrefix(path, skills+string(filepath.Separator)) {
		return Deny
	}
	if strings.HasPrefix(path, root+string(filepath.Separator)) {
		return Allow
	}
	return Ask
}

func main() {
	fmt.Println(skillWriteDecision(".", ".agent/skills/new/SKILL.md"))
	fmt.Println("skills are executable instructions: load/read them, but do not let the agent silently rewrite its own policy.")
}
