package museum

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

type Service struct {
	db       *sql.DB
	artworks *artwork.Repository
}

func NewService(db *sql.DB, artworks *artwork.Repository) *Service {
	return &Service{db: db, artworks: artworks}
}
func (s *Service) SaveDraft(ctx context.Context, set RuleSet) error {
	if err := Validate(set); err != nil {
		return err
	}
	data, err := json.Marshal(set)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO museum_rule_sets(id,version,state,seed,rules_json) VALUES(1,?,?,?,?) ON CONFLICT(id) DO UPDATE SET version=excluded.version,state=excluded.state,seed=excluded.seed,rules_json=excluded.rules_json`, set.Version, "draft", set.Seed, string(data))
	return err
}
func (s *Service) Draft(ctx context.Context) (RuleSet, error)     { return s.get(ctx, "draft", 1) }
func (s *Service) Published(ctx context.Context) (RuleSet, error) { return s.get(ctx, "published", 2) }
func (s *Service) Publish(ctx context.Context) error {
	set, err := s.Draft(ctx)
	if err != nil {
		return err
	}
	if err = Validate(set); err != nil {
		return err
	}
	works, err := s.artworks.ListPublic(ctx, "", "", "", "asc")
	if err != nil {
		return err
	}
	assignment, err := Evaluate(set, works)
	if err != nil {
		return err
	}
	if err := validateGeneratedPlan(GenerateLayout(assignment, set.Seed), assignment); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO museum_rule_sets(id,version,state,seed,rules_json) VALUES(2,?,?,?,?) ON CONFLICT(id) DO UPDATE SET version=excluded.version,state=excluded.state,seed=excluded.seed,rules_json=excluded.rules_json`, set.Version, "published", set.Seed, mustJSON(set))
	return err
}
func (s *Service) Scene(ctx context.Context) (struct {
	Version int  `json:"version"`
	Plan    Plan `json:"plan"`
}, error) {
	published, err := s.Published(ctx)
	if err != nil {
		return struct {
			Version int  `json:"version"`
			Plan    Plan `json:"plan"`
		}{}, err
	}
	works, err := s.artworks.ListPublic(ctx, "", "", "", "asc")
	if err != nil {
		return struct {
			Version int  `json:"version"`
			Plan    Plan `json:"plan"`
		}{}, err
	}
	assignment, err := Evaluate(published, works)
	if err != nil {
		return struct {
			Version int  `json:"version"`
			Plan    Plan `json:"plan"`
		}{}, err
	}
	plan := GenerateLayout(assignment, published.Seed)
	if err := validateGeneratedPlan(plan, assignment); err != nil {
		return struct {
			Version int  `json:"version"`
			Plan    Plan `json:"plan"`
		}{}, err
	}
	return struct {
		Version int  `json:"version"`
		Plan    Plan `json:"plan"`
	}{published.Version, plan}, nil
}

func validateGeneratedPlan(plan Plan, assignment Assignment) error {
	if len(plan.Errors) > 0 {
		return fmt.Errorf("invalid museum layout: %s", plan.Errors[0])
	}
	if err := ValidatePlacements(plan, assignment); err != nil {
		return fmt.Errorf("invalid museum layout: %w", err)
	}
	return nil
}

// Preview evaluates a draft without publishing it.
func (s *Service) Preview(ctx context.Context, draft RuleSet) (Assignment, Plan, error) {
	works, err := s.artworks.ListPublic(ctx, "", "", "", "asc")
	if err != nil {
		return Assignment{}, Plan{}, err
	}
	assignment, err := Evaluate(draft, works)
	if err != nil {
		return Assignment{}, Plan{}, err
	}
	return assignment, GenerateLayout(assignment, draft.Seed), nil
}
func (s *Service) get(ctx context.Context, state string, id int) (RuleSet, error) {
	var data string
	err := s.db.QueryRowContext(ctx, `SELECT rules_json FROM museum_rule_sets WHERE id=? AND state=?`, id, state).Scan(&data)
	if err != nil {
		return RuleSet{}, err
	}
	var set RuleSet
	if err = json.Unmarshal([]byte(data), &set); err != nil {
		return RuleSet{}, fmt.Errorf("decode museum rules: %w", err)
	}
	return set, nil
}
func mustJSON(set RuleSet) string { data, _ := json.Marshal(set); return string(data) }
