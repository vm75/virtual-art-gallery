package store

import (
	"context"
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
