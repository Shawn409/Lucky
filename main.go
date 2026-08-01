package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	sessionCookieName = "careu_session"
	passwordIters     = 120000
	maxJSONBytes      = 1 << 20
)

type app struct {
	db *sql.DB
}

type profile struct {
	FullName string `json:"fullName"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Village  string `json:"village"`
	City     string `json:"city"`
	Phone    string `json:"phone"`
}

type userResponse struct {
	ID      int64   `json:"id"`
	Email   string  `json:"email"`
	Profile profile `json:"profile"`
}

type contextKey string

const userKey contextKey = "user"

func main() {
	db, err := openDatabase()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	a := &app{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/api/register", a.handleRegister)
	mux.HandleFunc("/api/login", a.handleLogin)
	mux.HandleFunc("/api/logout", a.handleLogout)
	mux.HandleFunc("/api/me", a.withAuth(a.handleMe))
	mux.HandleFunc("/api/profile", a.withAuth(a.handleProfile))
	mux.HandleFunc("/api/state", a.withAuth(a.handleState))

	addr := env("CAREU_ADDR", ":8080")
	log.Printf("CareU running at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func openDatabase() (*sql.DB, error) {
	if dsn := strings.TrimSpace(os.Getenv("CAREU_DSN")); dsn != "" {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		if err := db.Ping(); err != nil {
			return nil, err
		}
		return db, ensureTables(db)
	}

	name := env("CAREU_DB_NAME", "careu")
	if !regexp.MustCompile(`^[A-Za-z0-9_]+$`).MatchString(name) {
		return nil, fmt.Errorf("CAREU_DB_NAME must use only letters, numbers, and underscore")
	}
	user := env("CAREU_DB_USER", "root")
	pass := os.Getenv("CAREU_DB_PASS")
	addr := env("CAREU_DB_ADDR", "127.0.0.1:3306")

	adminCfg := mysql.NewConfig()
	adminCfg.User = user
	adminCfg.Passwd = pass
	adminCfg.Net = "tcp"
	adminCfg.Addr = addr
	adminCfg.ParseTime = true
	adminCfg.Params = map[string]string{"charset": "utf8mb4"}

	admin, err := sql.Open("mysql", adminCfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		return nil, err
	}
	if _, err := admin.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		return nil, err
	}

	cfg := *adminCfg
	cfg.DBName = name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, ensureTables(db)
}

func ensureTables(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS profiles (
			user_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
			full_name VARCHAR(160) NOT NULL,
			age INT NOT NULL DEFAULT 0,
			gender VARCHAR(20) NOT NULL DEFAULT 'male',
			village VARCHAR(160) NOT NULL DEFAULT '',
			city VARCHAR(160) NOT NULL DEFAULT '',
			phone VARCHAR(40) NOT NULL DEFAULT '',
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CONSTRAINT fk_profiles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			user_id BIGINT UNSIGNED NOT NULL,
			token_hash CHAR(64) NOT NULL UNIQUE,
			expires_at DATETIME NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_sessions_user (user_id),
			INDEX idx_sessions_expires (expires_at),
			CONSTRAINT fk_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS app_states (
			user_id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
			state_json JSON NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			CONSTRAINT fk_app_states_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/sihatai-sarawak.html" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "sihatai-sarawak.html")
}

func (a *app) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Profile  profile `json:"profile"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	req.Email = normalizeEmail(req.Email)
	if err := validateCredentials(req.Email, req.Password); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Profile = sanitizeProfile(req.Profile)
	if err := validateProfile(req.Profile); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	tx, err := a.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database unavailable")
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", req.Email, hash)
	if err != nil {
		if isDuplicate(err) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create account")
		return
	}
	userID, _ := res.LastInsertId()
	if _, err := tx.Exec(
		"INSERT INTO profiles (user_id, full_name, age, gender, village, city, phone) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, req.Profile.FullName, req.Profile.Age, req.Profile.Gender, req.Profile.Village, req.Profile.City, req.Profile.Phone,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save profile")
		return
	}
	if _, err := tx.Exec("INSERT INTO app_states (user_id, state_json) VALUES (?, JSON_OBJECT())", userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not prepare user state")
		return
	}
	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "could not finish registration")
		return
	}
	if err := a.createSession(w, r, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user": userResponse{ID: userID, Email: req.Email, Profile: req.Profile}})
}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	email := normalizeEmail(req.Email)
	var userID int64
	var stored string
	err := a.db.QueryRow("SELECT id, password_hash FROM users WHERE email = ?", email).Scan(&userID, &stored)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if !verifyPassword(req.Password, stored) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err := a.createSession(w, r, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	u, err := a.loadUser(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load user")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}

func (a *app) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		_, _ = a.db.Exec("DELETE FROM sessions WHERE token_hash = ?", sha256Hex(cookie.Value))
	}
	http.SetCookie(w, expiredCookie(r))
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *app) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	u := r.Context().Value(userKey).(userResponse)
	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}

func (a *app) handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	u := r.Context().Value(userKey).(userResponse)
	var req struct {
		Profile profile `json:"profile"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	p := sanitizeProfile(req.Profile)
	if err := validateProfile(p); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	_, err := a.db.Exec(
		`UPDATE profiles SET full_name=?, age=?, gender=?, village=?, city=?, phone=? WHERE user_id=?`,
		p.FullName, p.Age, p.Gender, p.Village, p.City, p.Phone, u.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update profile")
		return
	}
	u.Profile = p
	writeJSON(w, http.StatusOK, map[string]any{"user": u})
}

func (a *app) handleState(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value(userKey).(userResponse)
	switch r.Method {
	case http.MethodGet:
		var raw []byte
		err := a.db.QueryRow("SELECT state_json FROM app_states WHERE user_id = ?", u.ID).Scan(&raw)
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{"state": nil})
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not load state")
			return
		}
		var state any
		if err := json.Unmarshal(raw, &state); err != nil {
			writeError(w, http.StatusInternalServerError, "saved state is invalid")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"state": state})
	case http.MethodPut:
		var req struct {
			State json.RawMessage `json:"state"`
		}
		if !readJSON(w, r, &req) {
			return
		}
		if len(req.State) == 0 || !json.Valid(req.State) {
			writeError(w, http.StatusBadRequest, "invalid state")
			return
		}
		_, err := a.db.Exec(
			`INSERT INTO app_states (user_id, state_json) VALUES (?, ?)
			 ON DUPLICATE KEY UPDATE state_json=VALUES(state_json), updated_at=CURRENT_TIMESTAMP`,
			u.ID, string(req.State),
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not save state")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *app) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := a.currentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not signed in")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	}
}

func (a *app) currentUser(r *http.Request) (userResponse, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return userResponse{}, errors.New("missing session")
	}
	tokenHash := sha256Hex(cookie.Value)
	var userID int64
	err = a.db.QueryRow("SELECT user_id FROM sessions WHERE token_hash = ? AND expires_at > NOW()", tokenHash).Scan(&userID)
	if err != nil {
		return userResponse{}, err
	}
	return a.loadUser(userID)
}

func (a *app) loadUser(userID int64) (userResponse, error) {
	var u userResponse
	err := a.db.QueryRow(
		`SELECT u.id, u.email, p.full_name, p.age, p.gender, p.village, p.city, p.phone
		 FROM users u
		 JOIN profiles p ON p.user_id = u.id
		 WHERE u.id = ?`,
		userID,
	).Scan(&u.ID, &u.Email, &u.Profile.FullName, &u.Profile.Age, &u.Profile.Gender, &u.Profile.Village, &u.Profile.City, &u.Profile.Phone)
	return u, err
}

func (a *app) createSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, err := randomToken(32)
	if err != nil {
		return err
	}
	expires := time.Now().Add(7 * 24 * time.Hour)
	if _, err := a.db.Exec("INSERT INTO sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)", userID, sha256Hex(token), expires); err != nil {
		return err
	}
	_, _ = a.db.Exec("DELETE FROM sessions WHERE expires_at <= NOW()")
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
	return nil
}

func expiredCookie(r *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	}
}

func validateCredentials(email, password string) error {
	if !strings.Contains(email, "@") || len(email) > 255 {
		return errors.New("enter a valid email")
	}
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	return nil
}

func sanitizeProfile(p profile) profile {
	p.FullName = trimLen(p.FullName, 160)
	p.Gender = strings.ToLower(trimLen(p.Gender, 20))
	p.Village = trimLen(p.Village, 160)
	p.City = trimLen(p.City, 160)
	p.Phone = trimLen(p.Phone, 40)
	if p.Gender == "" {
		p.Gender = "male"
	}
	return p
}

func validateProfile(p profile) error {
	if p.FullName == "" {
		return errors.New("full name is required")
	}
	if p.Age < 0 || p.Age > 120 {
		return errors.New("age must be between 0 and 120")
	}
	switch p.Gender {
	case "male", "female", "other":
	default:
		return errors.New("gender must be male, female, or other")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	salt, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	key := pbkdf2SHA256([]byte(password), salt, passwordIters, 32)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", passwordIters, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}
	iters, err := strconv.Atoi(parts[1])
	if err != nil || iters <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	actual := pbkdf2SHA256([]byte(password), salt, iters, len(expected))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	hLen := 32
	numBlocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, numBlocks*hLen)
	for block := 1; block <= numBlocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		mac.Write([]byte{byte(block >> 24), byte(block >> 16), byte(block >> 8), byte(block)})
		u := mac.Sum(nil)
		t := append([]byte(nil), u...)
		for i := 1; i < iter; i++ {
			mac = hmac.New(sha256.New, password)
			mac.Write(u)
			u = mac.Sum(nil)
			for j := range t {
				t[j] ^= u[j]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBytes)
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func randomToken(n int) (string, error) {
	b, err := randomBytes(n)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func trimLen(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
