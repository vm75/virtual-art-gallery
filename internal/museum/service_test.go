package museum

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/vm75/virtual-art-gallery/internal/artwork"
	"github.com/vm75/virtual-art-gallery/internal/store"
	"testing"
)

func TestDraftIsolationAndPublish(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	artworks := artwork.NewRepository(db.DB())
	_, _ = artworks.Create(context.Background(), artwork.Input{Name: "Oil", Date: "2020-01-01", Surface: "canvas", Medium: "oil", Visible: true})
	service := NewService(db.DB(), artworks)
	set := RuleSet{Version: 3, Seed: 44, Groups: []Group{{ID: "oil", Name: "Oil"}}, Rules: []Rule{{ID: "oil", Priority: 1, Group: "oil", All: []Condition{{Field: "medium", Op: "equals", Value: "oil"}}}}}
	if err := service.SaveDraft(context.Background(), set); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Published(context.Background()); err != sql.ErrNoRows {
		t.Fatalf("published before publish: %v", err)
	}
	if err := service.Publish(context.Background()); err != nil {
		t.Fatal(err)
	}
	scene, err := service.Scene(context.Background())
	if err != nil || scene.Version != 3 || len(scene.Plan.Placements) != 1 {
		t.Fatalf("scene=%+v err=%v", scene, err)
	}
	bad := set
	bad.Version = 0
	if err := service.SaveDraft(context.Background(), bad); err == nil {
		t.Fatal("invalid draft saved")
	}
	published, err := service.Published(context.Background())
	if err != nil || published.Version != 3 {
		t.Fatalf("published replaced: %+v err=%v", published, err)
	}
}

func TestPreviewReportsUnclassifiedWorks(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	artworks := artwork.NewRepository(db.DB())
	_, _ = artworks.Create(context.Background(), artwork.Input{Name: "Oil", Date: "2020-01-01", Surface: "canvas", Medium: "oil", Visible: true})
	_, _ = artworks.Create(context.Background(), artwork.Input{Name: "Ink", Date: "2021-01-01", Surface: "paper", Medium: "ink", Visible: true})
	service := NewService(db.DB(), artworks)
	draft := RuleSet{Version: 1, Seed: 9, Groups: []Group{{ID: "oil", Name: "Oil"}}, Rules: []Rule{{ID: "oil", Priority: 1, Group: "oil", All: []Condition{{Field: "medium", Op: "equals", Value: "oil"}}}}}
	assignment, plan, err := service.Preview(context.Background(), draft)
	if err != nil {
		t.Fatal(err)
	}
	if len(assignment.Unclassified) != 1 || len(plan.Placements) != 2 || len(plan.Errors) != 0 {
		t.Fatalf("assignment=%+v plan=%+v", assignment, plan)
	}
}

func TestPublishRejectsCapacityBrokenLayout(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	artworks := artwork.NewRepository(db.DB())
	for index := 0; index < 81; index++ {
		if _, err := artworks.Create(context.Background(), artwork.Input{Name: fmt.Sprintf("Work %03d", index), Date: "2020-01-01", Surface: "canvas", Medium: "oil", Visible: true}); err != nil {
			t.Fatal(err)
		}
	}
	service := NewService(db.DB(), artworks)
	set := RuleSet{Version: 1, Seed: 1, Groups: []Group{{ID: "all", Name: "All"}}, Rules: []Rule{{ID: "all", Priority: 1, Group: "all", All: []Condition{{Field: "medium", Op: "equals", Value: "oil"}}}}}
	if err := service.SaveDraft(context.Background(), set); err != nil {
		t.Fatal(err)
	}
	if err := service.Publish(context.Background()); err == nil {
		t.Fatal("capacity-broken layout was published")
	}
}
