package artwork

import (
	"context"
	"strings"
	"testing"

	"github.com/vm75/virtual-art-gallery/internal/store"
)

func TestCreateNormalizeSlugAndPublicVisibility(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := NewRepository(db.DB())
	in := Input{Name: "  Blue / Hour  ", Date: "2024-01-02", Tags: []string{"Sky", " sky ", "Night"}, Surface: " Canvas ", Medium: " Oil ", AltText: "  A blue hour landscape  "}
	first, err := r.Create(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if first.Slug != "blue-hour" || len(first.Tags) != 2 || first.Surface != "canvas" || first.AltText != "A blue hour landscape" {
		t.Fatalf("unexpected artwork: %+v", first)
	}
	second, err := r.Create(context.Background(), Input{Name: in.Name, Date: "2023-01-02", Surface: "canvas", Medium: "oil", Visible: true})
	if err != nil {
		t.Fatal(err)
	}
	if second.Slug != "blue-hour-2" {
		t.Fatalf("collision slug = %q", second.Slug)
	}
	items, err := r.ListPublic(context.Background(), "", "", "", "asc")
	if err != nil || len(items) != 1 || items[0].ID != second.ID {
		t.Fatalf("public list = %+v, err=%v", items, err)
	}
	if _, err := r.Get(context.Background(), first.Slug, true); err == nil {
		t.Fatal("hidden artwork leaked from public detail")
	}
	updated, err := r.Update(context.Background(), first.Slug, Input{Name: "Blue Hour Updated", Date: "2024-01-02", Surface: "canvas", Medium: "oil", AltText: "Updated description", Visible: true})
	if err != nil || updated.Slug != first.Slug || updated.Name != "Blue Hour Updated" || updated.AltText != "Updated description" {
		t.Fatalf("update = %+v, err=%v", updated, err)
	}
	surfaces, mediums, err := r.Taxonomy(context.Background())
	if err != nil || !contains(surfaces, "canvas") || !contains(mediums, "oil") {
		t.Fatalf("taxonomy = %v/%v, err=%v", surfaces, mediums, err)
	}
}

func TestValidateInput(t *testing.T) {
	_, err := ValidateInput(Input{Name: "x", Date: "bad", Surface: "canvas", Medium: "oil"})
	if err == nil {
		t.Fatal("expected invalid date")
	}
	_, err = ValidateInput(Input{Name: "x", Date: "2024-01-01", Surface: "canvas", Medium: "oil", AltText: strings.Repeat("x", 1001)})
	if err == nil {
		t.Fatal("expected oversized alt text rejection")
	}
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
