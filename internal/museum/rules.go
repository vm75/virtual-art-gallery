package museum

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

type RuleSet struct {
	Version int     `json:"version"`
	Seed    int64   `json:"seed"`
	Groups  []Group `json:"groups"`
	Rules   []Rule  `json:"rules"`
}
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Rule struct {
	ID       string      `json:"id"`
	Priority int         `json:"priority"`
	All      []Condition `json:"all"`
	Group    string      `json:"group"`
}
type Condition struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value,omitempty"`
	From  string `json:"from,omitempty"`
	To    string `json:"to,omitempty"`
}
type Assignment struct {
	Groups       map[string][]artwork.Artwork
	Unclassified []artwork.Artwork
}

func Validate(set RuleSet) error {
	if set.Version < 1 || set.Version > 1000000 {
		return fmt.Errorf("version must be between 1 and 1000000")
	}
	if len(set.Groups) == 0 || len(set.Groups) > 100 {
		return fmt.Errorf("rules must define 1-100 groups")
	}
	groups := map[string]bool{}
	for _, group := range set.Groups {
		if !validID(group.ID) || strings.TrimSpace(group.Name) == "" || groups[group.ID] {
			return fmt.Errorf("invalid or duplicate group %q", group.ID)
		}
		groups[group.ID] = true
	}
	rules := map[string]bool{}
	for _, rule := range set.Rules {
		if !validID(rule.ID) || rules[rule.ID] || !groups[rule.Group] || len(rule.All) == 0 {
			return fmt.Errorf("invalid rule %q", rule.ID)
		}
		rules[rule.ID] = true
		for _, condition := range rule.All {
			if err := validateCondition(condition); err != nil {
				return fmt.Errorf("rule %s: %w", rule.ID, err)
			}
		}
	}
	return nil
}

func Evaluate(set RuleSet, works []artwork.Artwork) (Assignment, error) {
	if err := Validate(set); err != nil {
		return Assignment{}, err
	}
	ordered := append([]Rule(nil), set.Rules...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Priority > ordered[j].Priority })
	result := Assignment{Groups: make(map[string][]artwork.Artwork)}
	for _, work := range works {
		assigned := false
		for _, rule := range ordered {
			if matches(rule.All, work) {
				result.Groups[rule.Group] = append(result.Groups[rule.Group], work)
				assigned = true
				break
			}
		}
		if !assigned {
			result.Unclassified = append(result.Unclassified, work)
		}
	}
	return result, nil
}

func validateCondition(c Condition) error {
	if c.Field != "tag" && c.Field != "surface" && c.Field != "medium" && c.Field != "date" {
		return fmt.Errorf("unknown field %q", c.Field)
	}
	if c.Field == "date" {
		if c.Op != "before" && c.Op != "after" && c.Op != "between" {
			return fmt.Errorf("unknown date operator %q", c.Op)
		}
		if _, err := time.Parse("2006-01-02", c.Value); c.Op != "between" && (err != nil || c.Value == "") {
			return fmt.Errorf("date value must be YYYY-MM-DD")
		}
		if c.Op == "between" {
			if _, err := time.Parse("2006-01-02", c.From); err != nil {
				return fmt.Errorf("range start must be YYYY-MM-DD")
			}
			if _, err := time.Parse("2006-01-02", c.To); err != nil {
				return fmt.Errorf("range end must be YYYY-MM-DD")
			}
			if c.From > c.To {
				return fmt.Errorf("date range is reversed")
			}
			return nil
		}
	} else if c.Op != "equals" && c.Op != "contains" {
		return fmt.Errorf("unknown operator %q", c.Op)
	}
	if strings.TrimSpace(c.Value) == "" || len(c.Value) > 200 {
		return fmt.Errorf("condition value is required and bounded")
	}
	return nil
}

func matches(conditions []Condition, work artwork.Artwork) bool {
	for _, c := range conditions {
		value := work.Surface
		if c.Field == "medium" {
			value = work.Medium
		}
		if c.Field == "tag" {
			found := false
			for _, tag := range work.Tags {
				if compare(tag, c) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
			continue
		}
		if c.Field == "date" {
			if c.Op == "before" && !(work.Date < c.Value) || c.Op == "after" && !(work.Date > c.Value) || c.Op == "between" && !(work.Date >= c.From && work.Date <= c.To) {
				return false
			}
			continue
		}
		if !compare(value, c) {
			return false
		}
	}
	return true
}
func compare(value string, c Condition) bool {
	if c.Op == "equals" {
		return strings.EqualFold(value, c.Value)
	}
	return strings.Contains(strings.ToLower(value), strings.ToLower(c.Value))
}
func validID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, r := range value {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
