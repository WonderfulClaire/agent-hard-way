package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSkills(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "research")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `---
name: research
description: search and compare papers
---
# Steps
Read sources.
`
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	skills, err := loadSkills(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 || skills[0].Name != "research" {
		t.Fatalf("unexpected skills: %#v", skills)
	}
	if skills[0].Description == "" || skills[0].Body == "" {
		t.Fatal("metadata/body missing")
	}
}
