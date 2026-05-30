package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

type LinearWebhookPayload struct {
	Action    string      `json:"action"`
	Type      string      `json:"type"`
	CreatedAt string      `json:"createdAt"`
	Data      interface{} `json:"data"`
	// Linear sends a top-level "id" or we derive one from type+action+createdAt
	UndefinedID string `json:"id,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL env var is required")
	}

	var err error
	db, err = sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := createSchema(); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("Database connected and schema ready")

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/webhook", webhookHandler)

	log.Printf("Starting linear-relay on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func createSchema() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS webhook_events (
			event_id TEXT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	return err
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, "database unreachable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate HMAC-SHA256 signature
	signature := r.Header.Get("X-Linear-Signature")
	if signature == "" {
		log.Println("Missing X-Linear-Signature header")
		http.Error(w, "missing signature", http.StatusUnauthorized)
		return
	}

	webhookSecret := os.Getenv("LINEAR_WEBHOOK_SECRET")
	if !verifySignature(body, signature, webhookSecret) {
		log.Println("Invalid signature")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// Parse payload to extract event ID
	var payload LinearWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Derive a stable event ID from the payload.
	// Linear doesn't always include a top-level "id", so we build one
	// from type + action + createdAt which is unique per event.
	eventID := payload.UndefinedID
	if eventID == "" {
		eventID = fmt.Sprintf("%s_%s_%s", payload.Type, payload.Action, payload.CreatedAt)
	}
	if eventID == "_" || eventID == "" {
		// Fallback: use hex of SHA256 of the raw body (guaranteed unique per payload)
		hash := sha256.Sum256(body)
		eventID = hex.EncodeToString(hash[:])
	}

	// Idempotency check
	alreadySeen, err := isDuplicate(eventID)
	if err != nil {
		log.Printf("Error checking dedup: %v", err)
		// Don't block on dedup errors, still try to forward
	}
	if alreadySeen {
		log.Printf("Duplicate event %s, skipping", eventID)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "duplicate")
		return
	}

	// Record event ID as seen
	if err := recordEvent(eventID); err != nil {
		log.Printf("Error recording event: %v", err)
		// Non-fatal: continue forwarding
	}

	// Transform and forward to SuperPlane
	superplaneURL := os.Getenv("SUPERPLANE_WEBHOOK_URL")
	if superplaneURL == "" {
		log.Println("SUPERPLANE_WEBHOOK_URL not set, dropping event")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "forwarded (no destination)")
		return
	}

	transformed := transformPayload(payload)

	forwardBody, err := json.Marshal(transformed)
	if err != nil {
		log.Printf("Error marshaling transformed payload: %v", err)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "transform error")
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(superplaneURL, "application/json", bytes.NewReader(forwardBody))
	if err != nil {
		log.Printf("Error forwarding to SuperPlane: %v", err)
		// Return 200 to Linear so it doesn't retry; we already recorded the event
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "forward error")
		return
	}
	defer resp.Body.Close()

	log.Printf("Forwarded event %s to SuperPlane (status %d)", eventID, resp.StatusCode)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}

func verifySignature(body []byte, signature string, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

func isDuplicate(eventID string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM webhook_events WHERE event_id = $1)", eventID).Scan(&exists)
	return exists, err
}

func recordEvent(eventID string) error {
	_, err := db.Exec("INSERT INTO webhook_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING", eventID)
	return err
}

func transformPayload(payload LinearWebhookPayload) map[string]interface{} {
	result := map[string]interface{}{
		"source":    "linear",
		"type":      payload.Type,
		"action":    payload.Action,
		"data":      payload.Data,
		"timestamp": payload.CreatedAt,
	}
	if payload.UndefinedID != "" {
		result["event_id"] = payload.UndefinedID
	}
	return result
}

