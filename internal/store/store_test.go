package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
)

func TestOpenCreatesAndRepeatsMigrations(t *testing.T) {
	dir := t.TempDir()
	first, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var count int
	if err := second.DB().QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("migration count = %d, want 4", count)
	}
	if _, err := filepath.Abs(filepath.Join(dir, "gallery.db")); err != nil {
		t.Fatal(err)
	}
}

func TestConnectionsEnforceForeignKeysAndCascade(t *testing.T) {
	s, err := Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	s.DB().SetMaxOpenConns(2)
	first, err := s.DB().Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := s.DB().Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	for index, conn := range []interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
	}{first, second} {
		var foreignKeys, busyTimeout int
		if err := conn.QueryRowContext(context.Background(), "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
			t.Fatal(err)
		}
		if err := conn.QueryRowContext(context.Background(), "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			t.Fatal(err)
		}
		if foreignKeys != 1 || busyTimeout != 5000 {
			t.Fatalf("connection %d pragmas = foreign_keys:%d busy_timeout:%d", index, foreignKeys, busyTimeout)
		}
	}
	if _, err := first.ExecContext(context.Background(), `INSERT INTO artworks(id,slug,name,date,surface,medium) VALUES(1,'work','Work','2024-01-01','canvas','oil')`); err != nil {
		t.Fatal(err)
	}
	if _, err := first.ExecContext(context.Background(), `INSERT INTO tags(id,name) VALUES(1,'tag')`); err != nil {
		t.Fatal(err)
	}
	if _, err := first.ExecContext(context.Background(), `INSERT INTO artwork_tags(artwork_id,tag_id) VALUES(1,1)`); err != nil {
		t.Fatal(err)
	}
	if _, err := second.ExecContext(context.Background(), `DELETE FROM artworks WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	var joined int
	if err := first.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM artwork_tags`).Scan(&joined); err != nil {
		t.Fatal(err)
	}
	if joined != 0 {
		t.Fatalf("cascade left %d artwork tags", joined)
	}
}

func TestConcurrentOpenMigrations(t *testing.T) {
	dir := t.TempDir()
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := Open(context.Background(), dir)
			if err != nil {
				errs <- err
				return
			}
			errs <- s.Close()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}
