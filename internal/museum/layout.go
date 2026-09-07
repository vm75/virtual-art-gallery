package museum

import (
	"fmt"
	"sort"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

type Plan struct {
	Seed        int64        `json:"seed"`
	Spawn       string       `json:"spawn"`
	Rooms       []Room       `json:"rooms"`
	Connections []Connection `json:"connections"`
	Placements  []Placement  `json:"placements"`
	Errors      []string     `json:"errors"`
}
type Room struct {
	ID           string `json:"id"`
	Group        string `json:"group"`
	Width        int    `json:"width"`
	Depth        int    `json:"depth"`
	WallSegments int    `json:"wall_segments"`
}
type Connection struct {
	From string `json:"from"`
	To   string `json:"to"`
}
type Placement struct {
	ID          string  `json:"id"`
	RoomID      string  `json:"room_id"`
	ArtworkSlug string  `json:"artwork_slug"`
	Wall        int     `json:"wall"`
	Slot        int     `json:"slot"`
	Aspect      float64 `json:"aspect"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
}

func ValidatePlacements(plan Plan, assignment Assignment) error {
	rooms := make(map[string]bool, len(plan.Rooms))
	for _, room := range plan.Rooms {
		rooms[room.ID] = true
	}
	seen := map[string]bool{}
	expected := map[string]bool{}
	for _, works := range assignment.Groups {
		for _, work := range works {
			expected[work.Slug] = true
		}
	}
	for _, work := range assignment.Unclassified {
		expected[work.Slug] = true
	}
	for _, placement := range plan.Placements {
		if !rooms[placement.RoomID] {
			return fmt.Errorf("placement %s references unknown room", placement.ID)
		}
		if !expected[placement.ArtworkSlug] {
			return fmt.Errorf("placement %s references unassigned artwork", placement.ID)
		}
		if seen[placement.ArtworkSlug] {
			return fmt.Errorf("artwork %s has multiple placements", placement.ArtworkSlug)
		}
		seen[placement.ArtworkSlug] = true
	}
	for slug := range expected {
		if !seen[slug] {
			return fmt.Errorf("artwork %s has no placement", slug)
		}
	}
	return nil
}

func GenerateLayout(assignment Assignment, seed int64) Plan {
	plan := Plan{Seed: seed}
	groupIDs := make([]string, 0, len(assignment.Groups))
	for id := range assignment.Groups {
		groupIDs = append(groupIDs, id)
	}
	if len(assignment.Unclassified) > 0 {
		groupIDs = append(groupIDs, "unclassified")
	}
	sort.Strings(groupIDs)
	for index, groupID := range groupIDs {
		works := append([]artwork.Artwork(nil), assignment.Groups[groupID]...)
		if groupID == "unclassified" {
			works = append(works, assignment.Unclassified...)
		}
		sort.Slice(works, func(i, j int) bool { return works[i].Slug < works[j].Slug })
		roomID := fmt.Sprintf("room-%02d-%s", index+1, groupID)
		segments := (len(works) + 11) / 12
		if segments < 1 {
			segments = 1
		}
		if segments > 10 {
			segments = 10
		}
		capacity := segments * 12
		room := Room{ID: roomID, Group: groupID, Width: 12 + segments*2, Depth: 12 + segments*2, WallSegments: segments * 3}
		plan.Rooms = append(plan.Rooms, room)
		if index == 0 {
			plan.Spawn = roomID
		}
		if index > 0 {
			plan.Connections = append(plan.Connections, Connection{From: plan.Rooms[index-1].ID, To: roomID})
		}
		for placementIndex, work := range works {
			if placementIndex >= capacity {
				plan.Errors = append(plan.Errors, fmt.Sprintf("room %s has insufficient wall capacity for %s", roomID, work.Slug))
				continue
			}
			aspect := 1.0
			if work.ImageWidth > 0 && work.ImageHeight > 0 {
				aspect = float64(work.ImageWidth) / float64(work.ImageHeight)
			}
			height := 2.4
			width := height * aspect
			if width > 2.8 {
				width = 2.8
				height = width / aspect
			}
			plan.Placements = append(plan.Placements, Placement{ID: fmt.Sprintf("%s-placement-%03d", roomID, placementIndex+1), RoomID: roomID, ArtworkSlug: work.Slug, Wall: (placementIndex / segments) % 3, Slot: placementIndex % segments, Aspect: aspect, Width: width, Height: height})
		}
	}
	return plan
}
