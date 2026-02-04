package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type APIResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func main() {
	// ✅ Load .env from project root (one level above /api)
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Println("Could not load .env from", envPath, "(using OS env vars)")
	}

	// Read DB config from environment (or fallback defaults)
	dbUser := getenv("DB_USER", "root")
	dbPass := getenv("DB_PASS", "")
	dbHost := getenv("DB_HOST", "127.0.0.1")
	dbPort := getenv("DB_PORT", "3306")
	dbName := getenv("DB_NAME", "userregistr")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		log.Fatal("DB ping failed: ", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/register", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, APIResponse{OK: false, Message: "method not allowed"})
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "invalid json"})
			return
		}

		req.FirstName = strings.TrimSpace(req.FirstName)
		req.LastName = strings.TrimSpace(req.LastName)
		req.Email = strings.TrimSpace(strings.ToLower(req.Email))

		// Validation
		if len(req.FirstName) < 2 {
			writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "first_name too short"})
			return
		}
		if len(req.LastName) < 2 {
			writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "last_name too short"})
			return
		}
		if !emailRe.MatchString(req.Email) {
			writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "invalid email"})
			return
		}
		if !isStrongPassword(req.Password) {
			writeJSON(w, http.StatusBadRequest, APIResponse{OK: false, Message: "weak password (8+, upper, lower, digit)"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "failed to hash password"})
			return
		}

		err = insertUser(db, req.FirstName, req.LastName, req.Email, string(hash))
		if err != nil {
			if errors.Is(err, ErrEmailExists) {
				writeJSON(w, http.StatusConflict, APIResponse{OK: false, Message: "email already exists"})
				return
			}
			log.Println("insert error:", err)
			writeJSON(w, http.StatusInternalServerError, APIResponse{OK: false, Message: "internal error"})
			return
		}

		writeJSON(w, http.StatusCreated, APIResponse{OK: true, Message: "registered"})
	}))

	server := &http.Server{
		Addr:              ":8080",
		Handler:           logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("API listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

// Password: at least 8 chars, 1 uppercase, 1 lowercase, 1 digit
func isStrongPassword(p string) bool {
	if len(p) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(p)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(p)
	hasDigit := regexp.MustCompile(`\d`).MatchString(p)
	return hasUpper && hasLower && hasDigit
}

var ErrEmailExists = errors.New("email exists")

func insertUser(db *sql.DB, firstName, lastName, email, passwordHash string) error {
	_, err := db.Exec(
		`INSERT INTO users (first_name, last_name, email, password_hash) VALUES (?, ?, ?, ?)`,
		firstName, lastName, email, passwordHash,
	)
	if err != nil {
		// MySQL duplicate key error code: 1062
		if strings.Contains(err.Error(), "1062") {
			return ErrEmailExists
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next(w, r)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
