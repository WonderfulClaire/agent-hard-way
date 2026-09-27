package main

import (
	"fmt"
	"sort"
	"strings"
)

type Skill struct {
	Name        string
	Description string
	Triggers    []string
}

type Match struct {
	Skill Skill
	Score int
}

func selectSkills(query string, skills []Skill, limit int) []Match {
	q := strings.ToLower(query)
	var matches []Match
	for _, skill := range skills {
		score := 0
		for _, trigger := range skill.Triggers {
			if strings.Contains(q, strings.ToLower(trigger)) {
				score += 3
			}
		}
		for _, token := range strings.Fields(strings.ToLower(skill.Description)) {
			if len(token) >= 4 && strings.Contains(q, token) {
				score++
			}
		}
		if score > 0 {
			matches = append(matches, Match{Skill: skill, Score: score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Skill.Name < matches[j].Skill.Name
		}
		return matches[i].Score > matches[j].Score
	})
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

func main() {
	skills := []Skill{
		{Name: "research", Description: "search papers and compare evidence", Triggers: []string{"paper", "research", "literature"}},
		{Name: "coding", Description: "inspect code and run tests", Triggers: []string{"code", "bug", "test"}},
	}
	for _, match := range selectSkills(strings.Join([]string{"find", "papers", "about", "agent", "rl"}, " "), skills, 2) {
		fmt.Printf("%s score=%d
", match.Skill.Name, match.Score)
	}
}
