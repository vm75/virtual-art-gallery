package artwork

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vm75/virtual-art-gallery/internal/images"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

type Artwork struct {
	ID          int64    `json:"id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Date        string   `json:"date"`
	Tags        []string `json:"tags"`
	Surface     string   `json:"surface"`
	Medium      string   `json:"medium"`
	Description string   `json:"description,omitempty"`
	AltText     string   `json:"alt_text,omitempty"`
	Visible     bool     `json:"visible"`
	Image       Images   `json:"image"`
	ImageWidth  int      `json:"image_width,omitempty"`
	ImageHeight int      `json:"image_height,omitempty"`
}

type Images struct {
	Thumbnail string `json:"thumbnail"`
	Medium    string `json:"medium"`
	Museum    string `json:"museum"`
	Large     string `json:"large"`
}

type Input struct {
	Name, Date, Surface, Medium string
	Tags                        []string
	Description, AltText        string
	Visible                     bool
}

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func ValidateInput(in Input) (Input, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Date = strings.TrimSpace(in.Date)
	in.Surface = normalize(in.Surface)
	in.Medium = normalize(in.Medium)
	in.Description = strings.TrimSpace(in.Description)
	in.AltText = strings.TrimSpace(in.AltText)
	if in.Name == "" || len(in.Name) > 240 {
		return Input{}, fmt.Errorf("name is required and must be at most 240 characters")
	}
	if _, err := time.Parse("2006-01-02", in.Date); err != nil {
		return Input{}, fmt.Errorf("date must use YYYY-MM-DD")
	}
	if in.Surface == "" || len(in.Surface) > 120 || in.Medium == "" || len(in.Medium) > 120 {
		return Input{}, fmt.Errorf("surface and medium are required and must be at most 120 characters")
	}
	if len(in.AltText) > 1000 {
		return Input{}, fmt.Errorf("alt text must be at most 1000 characters")
	}
	seen := make(map[string]bool, len(in.Tags))
	rawTags := in.Tags
	in.Tags = nil
	for _, tag := range rawTags {
		tag = normalize(tag)
		if tag != "" && len(tag) <= 80 && !seen[tag] {
			seen[tag] = true
			in.Tags = append(in.Tags, tag)
		}
	}
	return in, nil
}

func (r *Repository) Create(ctx context.Context, in Input) (Artwork, error) {
	var err error
	if in, err = ValidateInput(in); err != nil {
		return Artwork{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Artwork{}, err
	}
	defer tx.Rollback()
	if err := ensureTaxonomy(ctx, tx, in.Surface, in.Medium); err != nil {
		return Artwork{}, err
	}
	base := slugify(in.Name)
	for n := 1; ; n++ {
		slug := base
		if n > 1 {
			slug = fmt.Sprintf("%s-%d", base, n)
		}
		var exists int
		err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM artworks WHERE slug = ?`, slug).Scan(&exists)
		if err != nil {
			return Artwork{}, err
		}
		if exists == 0 {
			result, insertErr := tx.ExecContext(ctx, `INSERT INTO artworks(slug,name,date,surface,medium,description,alt_text,visible) VALUES(?,?,?,?,?,?,?,?)`, slug, in.Name, in.Date, in.Surface, in.Medium, in.Description, in.AltText, in.Visible)
			if insertErr != nil {
				return Artwork{}, insertErr
			}
			id, _ := result.LastInsertId()
			if err := replaceTags(ctx, tx, id, in.Tags); err != nil {
				return Artwork{}, err
			}
			if err := tx.Commit(); err != nil {
				return Artwork{}, err
			}
			return r.Get(ctx, slug, false)
		}
	}
}

func (r *Repository) Get(ctx context.Context, slug string, publicOnly bool) (Artwork, error) {
	query := `SELECT id,slug,name,date,surface,medium,description,alt_text,visible,image_width,image_height FROM artworks WHERE slug = ?`
	if publicOnly {
		query += ` AND visible = 1`
	}
	var a Artwork
	var visible int
	if err := r.db.QueryRowContext(ctx, query, slug).Scan(&a.ID, &a.Slug, &a.Name, &a.Date, &a.Surface, &a.Medium, &a.Description, &a.AltText, &visible, &a.ImageWidth, &a.ImageHeight); err != nil {
		return Artwork{}, err
	}
	a.Visible = visible == 1
	var paths [4]string
	_ = r.db.QueryRowContext(ctx, `SELECT thumbnail_path,medium_path,museum_path,large_path FROM artworks WHERE id=?`, a.ID).Scan(&paths[0], &paths[1], &paths[2], &paths[3])
	a.Image = Images{Thumbnail: mediaURL(paths[0]), Medium: mediaURL(paths[1]), Museum: mediaURL(paths[2]), Large: mediaURL(paths[3])}
	a.Tags = r.tags(ctx, a.ID)
	return a, nil
}

func (r *Repository) SetImages(ctx context.Context, slug string, result images.Result) error {
	_, err := r.db.ExecContext(ctx, `UPDATE artworks SET original_path=?,thumbnail_path=?,medium_path=?,museum_path=?,large_path=?,image_width=?,image_height=?,updated_at=CURRENT_TIMESTAMP WHERE slug=?`, result.Original, result.Thumbnail, result.Medium, result.Museum, result.Large, result.Width, result.Height, slug)
	return err
}

func (r *Repository) Delete(ctx context.Context, slug string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM artworks WHERE slug=?`, slug)
	return err
}

func (r *Repository) Update(ctx context.Context, slug string, in Input) (Artwork, error) {
	var err error
	if in, err = ValidateInput(in); err != nil {
		return Artwork{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Artwork{}, err
	}
	defer tx.Rollback()
	if err := ensureTaxonomy(ctx, tx, in.Surface, in.Medium); err != nil {
		return Artwork{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE artworks SET name=?,date=?,surface=?,medium=?,description=?,alt_text=?,visible=?,updated_at=CURRENT_TIMESTAMP WHERE slug=?`, in.Name, in.Date, in.Surface, in.Medium, in.Description, in.AltText, in.Visible, slug)
	if err != nil {
		return Artwork{}, err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return Artwork{}, sql.ErrNoRows
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM artworks WHERE slug=?`, slug).Scan(&id); err != nil {
		return Artwork{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM artwork_tags WHERE artwork_id=?`, id); err != nil {
		return Artwork{}, err
	}
	if err := replaceTags(ctx, tx, id, in.Tags); err != nil {
		return Artwork{}, err
	}
	if err := tx.Commit(); err != nil {
		return Artwork{}, err
	}
	return r.Get(ctx, slug, false)
}

func (r *Repository) ListPublic(ctx context.Context, tag, surface, medium, order string) ([]Artwork, error) {
	return r.list(ctx, true, tag, surface, medium, order)
}

func (r *Repository) ListAll(ctx context.Context) ([]Artwork, error) {
	return r.list(ctx, false, "", "", "", "asc")
}

func (r *Repository) list(ctx context.Context, publicOnly bool, tag, surface, medium, order string) ([]Artwork, error) {
	query := `SELECT id,slug,name,date,surface,medium,description,alt_text,visible,image_width,image_height,thumbnail_path,medium_path,museum_path,large_path FROM artworks WHERE 1=1`
	if publicOnly {
		query += ` AND visible = 1`
	}
	args := []any{}
	if tag = normalize(tag); tag != "" {
		query += ` AND EXISTS (SELECT 1 FROM artwork_tags at JOIN tags t ON t.id = at.tag_id WHERE at.artwork_id = artworks.id AND t.name = ?)`
		args = append(args, tag)
	}
	if surface = normalize(surface); surface != "" {
		query += ` AND surface = ?`
		args = append(args, surface)
	}
	if medium = normalize(medium); medium != "" {
		query += ` AND medium = ?`
		args = append(args, medium)
	}
	if strings.EqualFold(order, "asc") {
		query += ` ORDER BY date ASC, id ASC`
	} else {
		query += ` ORDER BY date DESC, id DESC`
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Artwork
	for rows.Next() {
		var a Artwork
		var visible int
		var paths [4]string
		if err := rows.Scan(&a.ID, &a.Slug, &a.Name, &a.Date, &a.Surface, &a.Medium, &a.Description, &a.AltText, &visible, &a.ImageWidth, &a.ImageHeight, &paths[0], &paths[1], &paths[2], &paths[3]); err != nil {
			return nil, err
		}
		a.Visible = visible == 1
		a.Image = Images{Thumbnail: mediaURL(paths[0]), Medium: mediaURL(paths[1]), Museum: mediaURL(paths[2]), Large: mediaURL(paths[3])}
		a.Tags = r.tags(ctx, a.ID)
		result = append(result, a)
	}
	return result, rows.Err()
}

func replaceTags(ctx context.Context, tx *sql.Tx, artworkID int64, tags []string) error {
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO tags(name) VALUES(?)`, tag); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO artwork_tags(artwork_id,tag_id) SELECT ?,id FROM tags WHERE name = ?`, artworkID, tag); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) tags(ctx context.Context, id int64) []string {
	rows, err := r.db.QueryContext(ctx, `SELECT t.name FROM tags t JOIN artwork_tags at ON at.tag_id=t.id WHERE at.artwork_id=? ORDER BY t.name`, id)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var tag string
		if rows.Scan(&tag) == nil {
			tags = append(tags, tag)
		}
	}
	return tags
}

func (r *Repository) Taxonomy(ctx context.Context) (surfaces, mediums []string, err error) {
	for index, query := range []string{`SELECT name FROM surfaces ORDER BY name`, `SELECT name FROM mediums ORDER BY name`} {
		rows, queryErr := r.db.QueryContext(ctx, query)
		if queryErr != nil {
			return nil, nil, queryErr
		}
		var values []string
		for rows.Next() {
			var value string
			if scanErr := rows.Scan(&value); scanErr != nil {
				rows.Close()
				return nil, nil, scanErr
			}
			values = append(values, value)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, nil, err
		}
		rows.Close()
		if index == 0 {
			surfaces = values
		} else {
			mediums = values
		}
	}
	return surfaces, mediums, nil
}

func ensureTaxonomy(ctx context.Context, tx *sql.Tx, surface, medium string) error {
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO surfaces(name) VALUES(?)`, surface); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO mediums(name) VALUES(?)`, medium)
	return err
}

func normalize(value string) string { return strings.ToLower(strings.Join(strings.Fields(value), " ")) }

func slugify(value string) string {
	value = strings.ToLower(value)
	value = nonSlug.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func mediaURL(path string) string {
	if path == "" {
		return ""
	}
	return "/media/" + path
}
