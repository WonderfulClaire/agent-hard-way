package main

import (
	"path/filepath"
	"testing"
)

func TestSkillDirectoryIsReadOnlyToAgent(t *testing.T) {
	root := t.TempDir()
	if got := skillWriteDecision(root, filepath.Join(root, ".agent", "skills", "x", "SKILL.md")); got != Deny {
		t.Fatalf("expected deny, got %s", got)
	}
	if got := skillWriteDecision(root, filepath.Join(root, "notes.md")); got != Allow {
		t.Fatalf("expected allow, got %s", got)
	}
	if got := skillWriteDecision(root, filepath.Join(root, "..", "outside.md")); got != Ask {
		t.Fatalf("expected ask, got %s", got)
	}
}
