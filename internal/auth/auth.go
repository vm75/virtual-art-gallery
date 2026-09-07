package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionLifetime = 12 * time.Hour

type Manager struct {
	db             *sql.DB
	SecureCookies  bool
	mu             sync.Mutex
	failedAttempts map[string][]time.Time
}

func NewManager(db *sql.DB, secureCookies bool) *Manager {
	return &Manager{db: db, SecureCookies: secureCookies, failedAttempts: make(map[string][]time.Time)}
}

func (m *Manager) NeedsSetup(ctx context.Context) (bool, error) {
	var exists bool
	err := m.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM admin WHERE id=1)`).Scan(&exists)
	return !exists, err
}

func (m *Manager) Setup(ctx context.Context, username, password string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 80 || strings.ContainsAny(username, " \t\r\n") {
		return fmt.Errorf("username must be 3-80 characters without spaces")
	}
	if len(password) < 12 || len(password) > 256 {
		return fmt.Errorf("password must be 12-256 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	result, err := m.db.ExecContext(ctx, `INSERT INTO admin(id,username,password_hash) VALUES(1,?,?)`, username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint") {
			return fmt.Errorf("admin setup is already complete")
		}
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return errors.New("admin setup failed")
	}
	return nil
}

func (m *Manager) Login(ctx context.Context, username, password, clientKey string) (string, string, error) {
	if !m.allowAttempt(clientKey) {
		return "", "", errors.New("login failed")
	}
	var hash []byte
	if err := m.db.QueryRowContext(ctx, `SELECT password_hash FROM admin WHERE username=?`, strings.TrimSpace(username)).Scan(&hash); err != nil || bcrypt.CompareHashAndPassword(hash, []byte(password)) != nil {
		m.recordFailure(clientKey)
		return "", "", errors.New("login failed")
	}
	m.clearFailures(clientKey)
	sessionToken, err := randomToken()
	if err != nil {
		return "", "", err
	}
	csrfToken, err := randomToken()
	if err != nil {
		return "", "", err
	}
	expires := time.Now().UTC().Add(sessionLifetime)
	_, err = m.db.ExecContext(ctx, `INSERT INTO sessions(token_hash,expires_at,csrf_hash) VALUES(?,?,?)`, digest(sessionToken), expires.Format(time.RFC3339Nano), digest(csrfToken))
	if err != nil {
		return "", "", fmt.Errorf("create session: %w", err)
	}
	return sessionToken, csrfToken, nil
}

func (m *Manager) Authenticate(ctx context.Context, request *http.Request) bool {
	cookie, err := request.Cookie("gallery_session")
	if err != nil || cookie.Value == "" {
		return false
	}
	var expires string
	err = m.db.QueryRowContext(ctx, `SELECT expires_at FROM sessions WHERE token_hash=?`, digest(cookie.Value)).Scan(&expires)
	if err != nil {
		return false
	}
	when, err := time.Parse(time.RFC3339Nano, expires)
	return err == nil && time.Now().UTC().Before(when)
}

func (m *Manager) Logout(ctx context.Context, request *http.Request) {
	if cookie, err := request.Cookie("gallery_session"); err == nil {
		_, _ = m.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash=?`, digest(cookie.Value))
	}
}

func (m *Manager) ValidateCSRF(ctx context.Context, request *http.Request) bool {
	session, err := request.Cookie("gallery_session")
	if err != nil {
		return false
	}
	token := request.Header.Get("X-CSRF-Token")
	if token == "" {
		token = request.FormValue("csrf_token")
	}
	if token == "" {
		return false
	}
	var stored []byte
	if err := m.db.QueryRowContext(ctx, `SELECT csrf_hash FROM sessions WHERE token_hash=?`, digest(session.Value)).Scan(&stored); err != nil {
		return false
	}
	return string(stored) == string(digest(token))
}

func (m *Manager) SetCookies(w http.ResponseWriter, sessionToken, csrfToken string) {
	base := func(name, value string, maxAge int) *http.Cookie {
		return &http.Cookie{Name: name, Value: value, Path: "/", MaxAge: maxAge, HttpOnly: name == "gallery_session", Secure: m.SecureCookies, SameSite: http.SameSiteLaxMode}
	}
	http.SetCookie(w, base("gallery_session", sessionToken, int(sessionLifetime.Seconds())))
	csrf := base("gallery_csrf", csrfToken, int(sessionLifetime.Seconds()))
	csrf.HttpOnly = false
	csrf.SameSite = http.SameSiteStrictMode
	http.SetCookie(w, csrf)
}

func (m *Manager) ClearCookies(w http.ResponseWriter) {
	m.SetCookies(w, "", "")
	for _, name := range []string{"gallery_session", "gallery_csrf"} {
		http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1, HttpOnly: name == "gallery_session", Secure: m.SecureCookies, SameSite: http.SameSiteLaxMode})
	}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func digest(value string) []byte { sum := sha256.Sum256([]byte(value)); return sum[:] }

func (m *Manager) allowAttempt(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	recent := m.failedAttempts[key][:0]
	for _, when := range m.failedAttempts[key] {
		if now.Sub(when) < time.Minute {
			recent = append(recent, when)
		}
	}
	m.failedAttempts[key] = recent
	return len(recent) < 5
}
func (m *Manager) recordFailure(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failedAttempts[key] = append(m.failedAttempts[key], time.Now())
}
func (m *Manager) clearFailures(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.failedAttempts, key)
}
