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
	if err := ValidatePhysicalConnections(plan); err != nil {
		return fmt.Errorf("invalid museum layout: %w", err)
	}
	if err := ValidatePlacements(plan, assignment); err != nil {
		return fmt.Errorf("invalid museum layout: %w", err)
	}
	return nil
}

// ValidatePhysicalConnections ensures the declared graph is walkable in the
// generated linear world, not merely connected as abstract room IDs.
func ValidatePhysicalConnections(plan Plan) error {
	rooms := make(map[string]Room, len(plan.Rooms))
	for _, room := range plan.Rooms {
		rooms[room.ID] = room
	}
	connected := make(map[string][]string, len(plan.Rooms))
	for _, connection := range plan.Connections {
		from, fromOK := rooms[connection.From]
		to, toOK := rooms[connection.To]
		if !fromOK || !toOK || connection.From == connection.To {
			return fmt.Errorf("connection references invalid rooms %q/%q", connection.From, connection.To)
		}
		if connection.FromDoorway.Position.X != from.Position.X+from.Width/2 || connection.ToDoorway.Position.X != to.Position.X-to.Width/2 || connection.FromDoorway.Position.Z != connection.ToDoorway.Position.Z || connection.Corridor.Length <= 0 || connection.Corridor.Width <= 0 || connection.Corridor.Height <= 0 {
			return fmt.Errorf("connection %s/%s has invalid doorway geometry", connection.From, connection.To)
		}
		start := connection.Corridor.Position.X - connection.Corridor.Length/2
		end := connection.Corridor.Position.X + connection.Corridor.Length/2
		if start != connection.FromDoorway.Position.X || end != connection.ToDoorway.Position.X || connection.Corridor.Position.Z != connection.FromDoorway.Position.Z || connection.Corridor.Width > connection.FromDoorway.Width || connection.Corridor.Width > connection.ToDoorway.Width {
			return fmt.Errorf("connection %s/%s has a corridor gap or misalignment", connection.From, connection.To)
		}
		connected[connection.From] = append(connected[connection.From], connection.To)
		connected[connection.To] = append(connected[connection.To], connection.From)
	}
	if len(plan.Rooms) == 0 {
		return nil
	}
	seen, queue := map[string]bool{plan.Spawn: true}, []string{plan.Spawn}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range connected[current] {
			if !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}
	if len(seen) != len(plan.Rooms) {
		return fmt.Errorf("not every room is physically reachable from %q", plan.Spawn)
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
