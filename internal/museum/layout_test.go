package museum

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

func TestLayoutDeterministicReachableAndSpatial(t *testing.T) {
	assignment := Assignment{Groups: map[string][]artwork.Artwork{
		"b": {{Slug: "b-1", ImageWidth: 1600, ImageHeight: 900}},
		"a": {{Slug: "a-1", ImageWidth: 900, ImageHeight: 1600}, {Slug: "a-2"}},
	}}
	first := GenerateLayout(assignment, 42)
	second := GenerateLayout(assignment, 42)
	if !reflect.DeepEqual(first, second) || first.Spawn == "" || len(first.Placements) != 3 {
		t.Fatalf("non-deterministic/incomplete plan: %+v", first)
	}
	if first.SpawnPosition.Y <= 0 || len(first.Connections) != len(first.Rooms)-1 {
		t.Fatalf("missing spawn/connectivity geometry: %+v", first)
	}
	assertReachable(t, first)
	for _, connection := range first.Connections {
		if connection.FromDoorway.Width <= 0 || connection.ToDoorway.Width <= 0 || connection.FromDoorway.Position == connection.ToDoorway.Position {
			t.Fatalf("invalid doorway geometry: %+v", connection)
		}
	}
	for index := 1; index < len(first.Rooms); index++ {
		left, right := first.Rooms[index-1], first.Rooms[index]
		if left.Position.X+left.Width/2 >= right.Position.X-right.Width/2 {
			t.Fatalf("overlapping rooms: %+v and %+v", left, right)
		}
	}
	assertUniqueLocations(t, first)
	for _, placement := range first.Placements {
		if placement.ID == "" || placement.RoomID == "" || placement.ArtworkSlug == "" || placement.Aspect <= 0 || placement.Width <= 0 || placement.Height <= 0 || placement.Normal == (Vector3{}) || placement.Transform.Position != placement.Position || placement.Transform.Normal != placement.Normal {
			t.Fatalf("incomplete placement: %+v", placement)
		}
	}
	if err := ValidatePlacements(first, assignment); err != nil {
		t.Fatal(err)
	}
}

func TestLayoutCapacityAndDistribution(t *testing.T) {
	for _, test := range []struct {
		name       string
		assignment Assignment
		placements int
		errors     int
	}{
		{name: "empty", assignment: Assignment{}, placements: 0, errors: 0},
		{name: "one", assignment: groupedWorks("one", 1), placements: 1, errors: 0},
		{name: "four", assignment: groupedWorks("four", 4), placements: 4, errors: 0},
		{name: "large", assignment: groupedWorks("large", 81), placements: 80, errors: 1},
		{name: "uneven", assignment: Assignment{Groups: map[string][]artwork.Artwork{"a": works("a", 1), "b": works("b", 17), "c": works("c", 4)}}, placements: 22, errors: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			plan := GenerateLayout(test.assignment, 7)
			if len(plan.Placements) != test.placements || len(plan.Errors) != test.errors {
				t.Fatalf("placements/errors = %d/%d, want %d/%d", len(plan.Placements), len(plan.Errors), test.placements, test.errors)
			}
			if test.name == "empty" {
				if len(plan.Rooms) != 0 || plan.Spawn != "" {
					t.Fatalf("unexpected empty plan: %+v", plan)
				}
				return
			}
			assertReachable(t, plan)
			assertUniqueLocations(t, plan)
			if len(plan.Errors) == 0 {
				if err := ValidatePlacements(plan, test.assignment); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestValidatePlacementsRejectsDuplicatePhysicalLocation(t *testing.T) {
	assignment := groupedWorks("works", 2)
	plan := GenerateLayout(assignment, 1)
	plan.Placements[1].Wall = plan.Placements[0].Wall
	plan.Placements[1].Slot = plan.Placements[0].Slot
	plan.Placements[1].Position = plan.Placements[0].Position
	if err := ValidatePlacements(plan, assignment); err == nil {
		t.Fatal("duplicate physical placement accepted")
	}
}

func TestValidatePlacementsRejectsMissingAndWrongGroup(t *testing.T) {
	assignment := Assignment{Groups: map[string][]artwork.Artwork{"a": works("a", 1), "b": works("b", 1)}}
	plan := GenerateLayout(assignment, 1)
	plan.Placements = plan.Placements[:1]
	if err := ValidatePlacements(plan, assignment); err == nil {
		t.Fatal("missing placement accepted")
	}
	plan = GenerateLayout(assignment, 1)
	plan.Placements[0].RoomID = plan.Rooms[1].ID
	if err := ValidatePlacements(plan, assignment); err == nil {
		t.Fatal("cross-group placement accepted")
	}
}

func TestPlacementTransformsHandleAspectsDeterministically(t *testing.T) {
	assignment := Assignment{Groups: map[string][]artwork.Artwork{"works": {
		{Slug: "extreme-landscape", ImageWidth: 10000, ImageHeight: 100},
		{Slug: "portrait", ImageWidth: 100, ImageHeight: 1000},
		{Slug: "square", ImageWidth: 1000, ImageHeight: 1000},
	}}}
	plan := GenerateLayout(assignment, 99)
	if err := ValidatePlacements(plan, assignment); err != nil {
		t.Fatal(err)
	}
	for _, placement := range plan.Placements {
		if placement.Width > 2.8 || placement.Height <= 0 || placement.Transform.Rotation.Y < 0 || placement.Transform.Rotation.Y > 3.141592653589793 {
			t.Fatalf("invalid aspect transform: %+v", placement)
		}
	}
	if !reflect.DeepEqual(plan, GenerateLayout(assignment, 99)) {
		t.Fatal("transforms are not deterministic")
	}
}

func groupedWorks(group string, count int) Assignment {
	return Assignment{Groups: map[string][]artwork.Artwork{group: works(group, count)}}
}

func works(prefix string, count int) []artwork.Artwork {
	result := make([]artwork.Artwork, count)
	for index := range result {
		result[index] = artwork.Artwork{Slug: fmt.Sprintf("%s-%03d", prefix, index+1)}
	}
	return result
}

func assertUniqueLocations(t *testing.T, plan Plan) {
	t.Helper()
	slots := map[string]bool{}
	positions := map[Vector3]bool{}
	for _, placement := range plan.Placements {
		key := fmt.Sprintf("%s/%d/%d", placement.RoomID, placement.Wall, placement.Slot)
		if slots[key] || positions[placement.Position] {
			t.Fatalf("duplicate placement location: %+v", placement)
		}
		slots[key], positions[placement.Position] = true, true
		for _, connection := range plan.Connections {
			if placement.Position == connection.FromDoorway.Position || placement.Position == connection.ToDoorway.Position {
				t.Fatalf("placement occupies doorway: %+v", placement)
			}
		}
	}
}

func assertReachable(t *testing.T, plan Plan) {
	t.Helper()
	if len(plan.Rooms) == 0 {
		return
	}
	connected := map[string][]string{}
	for _, connection := range plan.Connections {
		connected[connection.From] = append(connected[connection.From], connection.To)
		connected[connection.To] = append(connected[connection.To], connection.From)
	}
	seen := map[string]bool{plan.Spawn: true}
	queue := []string{plan.Spawn}
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
		t.Fatalf("rooms are not reachable from %s: %+v", plan.Spawn, plan)
	}
}
