package main

import "testing"

func TestSelectSkills(t *testing.T) {
	skills := []Skill{
		{Name: "research", Description: "search papers", Triggers: []string{"paper", "literature"}},
		{Name: "coding", Description: "inspect code", Triggers: []string{"code", "bug"}},
	}
	got := selectSkills("find a paper about reward hacking", skills, 1)
	if len(got) != 1 || got[0].Skill.Name != "research" {
		t.Fatalf("unexpected match: %#v", got)
	}
}
