package museum

import (
	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"testing"
)

func TestEvaluateConditionsPriorityAndUnclassified(t *testing.T) {
	set := RuleSet{Version: 1, Groups: []Group{{ID: "blue", Name: "Blue"}, {ID: "canvas", Name: "Canvas"}}, Rules: []Rule{{ID: "surface", Priority: 1, All: []Condition{{Field: "surface", Op: "equals", Value: "canvas"}}, Group: "canvas"}, {ID: "blue", Priority: 2, All: []Condition{{Field: "tag", Op: "contains", Value: "blue"}, {Field: "date", Op: "between", From: "2020-01-01", To: "2025-01-01"}}, Group: "blue"}}}
	works := []artwork.Artwork{{Slug: "a", Surface: "canvas", Medium: "oil", Tags: []string{"blue sky"}, Date: "2022-01-01"}, {Slug: "b", Surface: "paper", Medium: "ink", Date: "2022-01-01"}}
	result, err := Evaluate(set, works)
	if err != nil || len(result.Groups["blue"]) != 1 || len(result.Unclassified) != 1 {
		t.Fatalf("assignment=%+v err=%v", result, err)
	}
}

func TestValidateRejectsUnsafeRules(t *testing.T) {
	bad := RuleSet{Version: 1, Groups: []Group{{ID: "x", Name: "X"}, {ID: "x", Name: "Duplicate"}}, Rules: nil}
	if Validate(bad) == nil {
		t.Fatal("duplicate group accepted")
	}
	bad = RuleSet{Version: 1, Groups: []Group{{ID: "x", Name: "X"}}, Rules: []Rule{{ID: "r", Group: "x", All: []Condition{{Field: "script", Op: "equals", Value: "x"}}}}}
	if Validate(bad) == nil {
		t.Fatal("unknown field accepted")
	}
}
