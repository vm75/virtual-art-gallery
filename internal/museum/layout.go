package museum

import (
	"fmt"
	"sort"

	"github.com/vm75/virtual-art-gallery/internal/artwork"
)

const (
	slotsPerWallSegment = 4
	usableWalls         = 2
	maxWallSegments     = 10
	slotWidth           = 3.0
	roomMargin          = 4.0
	roomDepth           = 14.0
	roomHeight          = 5.0
	roomGap             = 4.0
)

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Plan struct {
	Seed          int64        `json:"seed"`
	Spawn         string       `json:"spawn"`
	SpawnPosition Vector3      `json:"spawn_position"`
	Rooms         []Room       `json:"rooms"`
	Connections   []Connection `json:"connections"`
	Placements    []Placement  `json:"placements"`
	Errors        []string     `json:"errors"`
}

type Room struct {
	ID           string  `json:"id"`
	Group        string  `json:"group"`
	Position     Vector3 `json:"position"`
	Width        float64 `json:"width"`
	Depth        float64 `json:"depth"`
	Height       float64 `json:"height"`
	WallSegments int     `json:"wall_segments"`
	Capacity     int     `json:"capacity"`
}

type Doorway struct {
	Position Vector3 `json:"position"`
	Normal   Vector3 `json:"normal"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
}

// Transform is the renderer-ready placement contract. Rotation is expressed as
// Euler radians and Normal remains available for facing/culling decisions.
type Transform struct {
	Position Vector3 `json:"position"`
	Normal   Vector3 `json:"normal"`
	Rotation Vector3 `json:"rotation"`
}

type Connection struct {
	From        string   `json:"from"`
	To          string   `json:"to"`
	FromDoorway Doorway  `json:"from_doorway"`
	ToDoorway   Doorway  `json:"to_doorway"`
	Corridor    Corridor `json:"corridor"`
}

// Corridor is the physical bridge between two doorway planes. Length runs on
// the world X axis; Width runs on Z, matching the linear room layout.
type Corridor struct {
	Position Vector3 `json:"position"`
	Length   float64 `json:"length"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
}

type Placement struct {
	ID          string    `json:"id"`
	RoomID      string    `json:"room_id"`
	ArtworkSlug string    `json:"artwork_slug"`
	Wall        int       `json:"wall"`
	Slot        int       `json:"slot"`
	Position    Vector3   `json:"position"`
	Normal      Vector3   `json:"normal"`
	Transform   Transform `json:"transform"`
	Aspect      float64   `json:"aspect"`
	Width       float64   `json:"width"`
	Height      float64   `json:"height"`
}

func ValidatePlacements(plan Plan, assignment Assignment) error {
	rooms := make(map[string]Room, len(plan.Rooms))
	for _, room := range plan.Rooms {
		if _, exists := rooms[room.ID]; exists {
			return fmt.Errorf("duplicate room %s", room.ID)
		}
		rooms[room.ID] = room
	}
	expected := map[string]string{}
	for group, works := range assignment.Groups {
		for _, work := range works {
			if prior, exists := expected[work.Slug]; exists && prior != group {
				return fmt.Errorf("artwork %s belongs to multiple groups", work.Slug)
			}
			expected[work.Slug] = group
		}
	}
	for _, work := range assignment.Unclassified {
		if prior, exists := expected[work.Slug]; exists && prior != "unclassified" {
			return fmt.Errorf("artwork %s belongs to multiple groups", work.Slug)
		}
		expected[work.Slug] = "unclassified"
	}
	artworks := map[string]bool{}
	physicalSlots := map[string]bool{}
	positions := map[string]bool{}
	for _, placement := range plan.Placements {
		room, exists := rooms[placement.RoomID]
		if !exists {
			return fmt.Errorf("placement %s references unknown room", placement.ID)
		}
		group, expectedArtwork := expected[placement.ArtworkSlug]
		if !expectedArtwork {
			return fmt.Errorf("placement %s references unassigned artwork", placement.ID)
		}
		if room.Group != group {
			return fmt.Errorf("placement %s is outside its exhibition room", placement.ID)
		}
		if artworks[placement.ArtworkSlug] {
			return fmt.Errorf("artwork %s has multiple placements", placement.ArtworkSlug)
		}
		if placement.Wall < 0 || placement.Wall >= usableWalls || placement.Slot < 0 || placement.Slot >= room.WallSegments/usableWalls*slotsPerWallSegment {
			return fmt.Errorf("placement %s has invalid physical slot", placement.ID)
		}
		slotKey := fmt.Sprintf("%s/%d/%d", placement.RoomID, placement.Wall, placement.Slot)
		if physicalSlots[slotKey] {
			return fmt.Errorf("placement %s duplicates physical slot %s", placement.ID, slotKey)
		}
		positionKey := fmt.Sprintf("%.3f/%.3f/%.3f", placement.Position.X, placement.Position.Y, placement.Position.Z)
		if positions[positionKey] {
			return fmt.Errorf("placement %s duplicates world position", placement.ID)
		}
		if placement.Transform.Position != placement.Position || placement.Transform.Normal != placement.Normal || placement.Transform.Normal == (Vector3{}) {
			return fmt.Errorf("placement %s has an invalid renderer transform", placement.ID)
		}
		artworks[placement.ArtworkSlug] = true
		physicalSlots[slotKey] = true
		positions[positionKey] = true
	}
	for slug := range expected {
		if !artworks[slug] {
			return fmt.Errorf("artwork %s has no placement", slug)
		}
	}
	return nil
}

func GenerateLayout(assignment Assignment, seed int64) Plan {
	plan := Plan{Seed: seed}
	groupIDs := make([]string, 0, len(assignment.Groups)+1)
	for id := range assignment.Groups {
		groupIDs = append(groupIDs, id)
	}
	if len(assignment.Unclassified) > 0 && !containsGroup(groupIDs, "unclassified") {
		groupIDs = append(groupIDs, "unclassified")
	}
	sort.Strings(groupIDs)

	worksByGroup := make([][]artwork.Artwork, len(groupIDs))
	for index, groupID := range groupIDs {
		works := append([]artwork.Artwork(nil), assignment.Groups[groupID]...)
		if groupID == "unclassified" {
			works = append(works, assignment.Unclassified...)
		}
		sort.Slice(works, func(i, j int) bool { return works[i].Slug < works[j].Slug })
		worksByGroup[index] = works
	}

	cursorX := 0.0
	for index, groupID := range groupIDs {
		works := worksByGroup[index]
		segments := (len(works) + usableWalls*slotsPerWallSegment - 1) / (usableWalls * slotsPerWallSegment)
		if segments < 1 {
			segments = 1
		}
		if segments > maxWallSegments {
			segments = maxWallSegments
		}
		slotsPerWall := segments * slotsPerWallSegment
		width := roomMargin + float64(slotsPerWall)*slotWidth
		room := Room{
			ID:           fmt.Sprintf("room-%02d-%s", index+1, groupID),
			Group:        groupID,
			Position:     Vector3{X: cursorX + width/2},
			Width:        width,
			Depth:        roomDepth,
			Height:       roomHeight,
			WallSegments: segments * usableWalls,
			Capacity:     slotsPerWall * usableWalls,
		}
		plan.Rooms = append(plan.Rooms, room)
		cursorX += width + roomGap
	}

	if len(plan.Rooms) > 0 {
		plan.Spawn = plan.Rooms[0].ID
		plan.SpawnPosition = Vector3{X: plan.Rooms[0].Position.X, Y: 1.6, Z: 0}
	}
	for index := 1; index < len(plan.Rooms); index++ {
		from, to := plan.Rooms[index-1], plan.Rooms[index]
		fromDoorway := Doorway{Position: Vector3{X: from.Position.X + from.Width/2, Y: 1.5}, Normal: Vector3{X: 1}, Width: 2.4, Height: 3}
		toDoorway := Doorway{Position: Vector3{X: to.Position.X - to.Width/2, Y: 1.5}, Normal: Vector3{X: -1}, Width: 2.4, Height: 3}
		plan.Connections = append(plan.Connections, Connection{
			From: from.ID, To: to.ID,
			FromDoorway: fromDoorway, ToDoorway: toDoorway,
			Corridor: Corridor{Position: Vector3{X: (fromDoorway.Position.X + toDoorway.Position.X) / 2}, Length: toDoorway.Position.X - fromDoorway.Position.X, Width: min(fromDoorway.Width, toDoorway.Width), Height: min(fromDoorway.Height, toDoorway.Height)},
		})
	}
	for index, works := range worksByGroup {
		room := plan.Rooms[index]
		slotsPerWall := room.Capacity / usableWalls
		for placementIndex, work := range works {
			if placementIndex >= room.Capacity {
				plan.Errors = append(plan.Errors, fmt.Sprintf("room %s has insufficient wall capacity for %s", room.ID, work.Slug))
				continue
			}
			wall, slot := placementIndex/slotsPerWall, placementIndex%slotsPerWall
			position, normal := placementPosition(room, wall, slot)
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
			transform := Transform{Position: position, Normal: normal, Rotation: Vector3{Y: rotationFor(normal)}}
			plan.Placements = append(plan.Placements, Placement{ID: fmt.Sprintf("%s-placement-%03d", room.ID, placementIndex+1), RoomID: room.ID, ArtworkSlug: work.Slug, Wall: wall, Slot: slot, Position: position, Normal: normal, Transform: transform, Aspect: aspect, Width: width, Height: height})
		}
	}
	return plan
}

func containsGroup(groups []string, wanted string) bool {
	for _, group := range groups {
		if group == wanted {
			return true
		}
	}
	return false
}

func placementPosition(room Room, wall, slot int) (Vector3, Vector3) {
	localX := -room.Width/2 + roomMargin/2 + slotWidth/2 + float64(slot)*slotWidth
	if wall == 0 {
		return Vector3{X: room.Position.X + localX, Y: roomHeight / 2, Z: room.Position.Z + room.Depth/2}, Vector3{Z: -1}
	}
	return Vector3{X: room.Position.X + localX, Y: roomHeight / 2, Z: room.Position.Z - room.Depth/2}, Vector3{Z: 1}
}

func rotationFor(normal Vector3) float64 {
	if normal.Z < 0 {
		return 3.141592653589793
	}
	return 0
}
