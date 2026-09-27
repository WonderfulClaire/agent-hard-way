package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Skill struct {
	Name        string
	Description string
	Path        string
	Body        string
}

func loadSkills(root string) ([]Skill, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var skills []Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		name, desc, body := parseSkill(string(data))
		if name == "" {
			name = entry.Name()
		}
		skills = append(skills, Skill{Name: name, Description: desc, Path: path, Body: body})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

func parseSkill(raw string) (name, desc, body string) {
	lines := strings.Split(raw, "
")
	inFront := false
	bodyStart := 0
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 0 && line == "---" {
			inFront = true
			continue
		}
		if inFront && line == "---" {
			bodyStart = i + 1
			break
		}
		if inFront {
			switch {
			case strings.HasPrefix(line, "name:"):
				name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			case strings.HasPrefix(line, "description:"):
				desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}
	body = strings.TrimSpace(strings.Join(lines[bodyStart:], "
"))
	return
}

func main() {
	root := "skills"
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	skills, err := loadSkills(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, skill := range skills {
		fmt.Printf("%s	%s
", skill.Name, skill.Description)
	}
}
