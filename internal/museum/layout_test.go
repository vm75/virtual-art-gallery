package museum

import (
	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"reflect"
	"testing"
)

func TestLayoutDeterministicReachableAndExplicit(t *testing.T) {
	assignment := Assignment{Groups: map[string][]artwork.Artwork{"b": {{Slug: "b-1", ImageWidth: 1600, ImageHeight: 900}}, "a": {{Slug: "a-1", ImageWidth: 900, ImageHeight: 1600}, {Slug: "a-2"}}}}
	first := GenerateLayout(assignment, 42)
	second := GenerateLayout(assignment, 42)
	if !reflect.DeepEqual(first, second) || first.Spawn == "" || len(first.Placements) != 3 {
		t.Fatalf("non-deterministic/incomplete plan: %+v", first)
	}
	if len(first.Connections) != len(first.Rooms)-1 {
		t.Fatalf("connections=%v rooms=%v", first.Connections, first.Rooms)
	}
	for _, placement := range first.Placements {
		if placement.ID == "" || placement.RoomID == "" || placement.ArtworkSlug == "" {
			t.Fatalf("implicit placement: %+v", placement)
		}
		if placement.Aspect <= 0 || placement.Width <= 0 || placement.Height <= 0 {
			t.Fatalf("missing aspect sizing: %+v", placement)
		}
	}
	if err := ValidatePlacements(first, assignment); err != nil {
		t.Fatal(err)
	}
}

func TestLayoutReportsCapacityOverflow(t *testing.T) {
	works := make([]artwork.Artwork, 0, 130)
	for i := 0; i < 130; i++ {
		works = append(works, artwork.Artwork{Slug: string(rune('a'+i%26)) + string(rune(i))})
	}
	plan := GenerateLayout(Assignment{Groups: map[string][]artwork.Artwork{"large": works}}, 1)
	if len(plan.Errors) == 0 {
		t.Fatal("expected capacity error")
	}
}
