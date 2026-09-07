package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/vm75/virtual-art-gallery/internal/store"
)

func TestSingleSetupLoginCSRFAndLogout(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := NewManager(db.DB(), false)
	if ready, err := m.NeedsSetup(context.Background()); err != nil || !ready {
		t.Fatalf("needs setup = %v, %v", ready, err)
	}
	if err := m.Setup(context.Background(), "admin", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	if err := m.Setup(context.Background(), "other", "another correct password"); err == nil {
		t.Fatal("second setup succeeded")
	}
	session, csrf, err := m.Login(context.Background(), "admin", "correct horse battery staple", "test-client")
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	request.AddCookie(&http.Cookie{Name: "gallery_session", Value: session})
	request.AddCookie(&http.Cookie{Name: "gallery_csrf", Value: csrf})
	request.Header.Set("X-CSRF-Token", csrf)
	if !m.Authenticate(context.Background(), request) || !m.ValidateCSRF(context.Background(), request) {
		t.Fatal("valid session or csrf rejected")
	}
	request.Header.Set("X-CSRF-Token", "wrong")
	if m.ValidateCSRF(context.Background(), request) {
		t.Fatal("invalid csrf accepted")
	}
	request.Header.Set("X-CSRF-Token", csrf)
	m.Logout(context.Background(), request)
	if m.Authenticate(context.Background(), request) {
		t.Fatal("logged out session remained valid")
	}
}

func TestExpiredSessionRejected(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := NewManager(db.DB(), false)
	if err := m.Setup(context.Background(), "admin", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	session, _, err := m.Login(context.Background(), "admin", "correct horse battery staple", "client")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB().Exec(`UPDATE sessions SET expires_at=?`, time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	request.AddCookie(&http.Cookie{Name: "gallery_session", Value: session})
	if m.Authenticate(context.Background(), request) {
		t.Fatal("invalid session accepted")
	}
}

func TestConcurrentSetupCreatesOneAdmin(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := NewManager(db.DB(), false)
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, username := range []string{"first", "second"} {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			results <- m.Setup(context.Background(), name, "correct horse battery staple")
		}(username)
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful setups = %d, want 1", successes)
	}
}

func TestLoginThrottleUsesPeerIPAcrossPorts(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := NewManager(db.DB(), false)
	if err := m.Setup(context.Background(), "admin", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	for port := 1000; port < 1005; port++ {
		if _, _, err := m.Login(context.Background(), "admin", "wrong password", "192.0.2.8:"+strconv.Itoa(port)); err == nil {
			t.Fatal("invalid login succeeded")
		}
	}
	if _, _, err := m.Login(context.Background(), "admin", "correct horse battery staple", "192.0.2.8:2000"); err == nil {
		t.Fatal("login bypassed throttle by changing port")
	}
	if _, _, err := m.Login(context.Background(), "admin", "correct horse battery staple", "192.0.2.9:2000"); err != nil {
		t.Fatalf("other peer should not be throttled: %v", err)
	}
}

func TestFailureKeysExpireAndAreBounded(t *testing.T) {
	db, err := store.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m := NewManager(db.DB(), false)
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return now }
	for i := 0; i <= maxFailureKeys; i++ {
		m.recordFailure("198.51.100." + strconv.Itoa(i))
	}
	if got := len(m.failedAttempts); got != maxFailureKeys {
		t.Fatalf("failure key count = %d, want %d", got, maxFailureKeys)
	}
	now = now.Add(failureWindow)
	if !m.allowAttempt("203.0.113.1") {
		t.Fatal("expired attempts should not throttle")
	}
	if got := len(m.failedAttempts); got != 0 {
		t.Fatalf("expired failure keys retained: %d", got)
	}
}
