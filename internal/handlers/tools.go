package handlers

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"vickygo/internal/store"
)

// NotesApiHandler handles saving and loading temporary notes.
func NotesApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req struct {
			Content string `json:"content"`
			Expiry  string `json:"expiry"` // "1h", "24h", "168h"
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		dur, err := time.ParseDuration(req.Expiry)
		if err != nil || dur <= 0 {
			dur = 24 * time.Hour // default 24h
		}

		id := uuid.New().String()
		store.Global.SaveNote(id, req.Content, time.Now().Add(dur))
		json.NewEncoder(w).Encode(map[string]string{"id": id})
		return
	}

	if r.Method == http.MethodGet {
		id := r.URL.Query().Get("id")
		note, ok := store.Global.GetNote(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Note not found or expired"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"content": note.Content, "expires": note.ExpiresAt.UTC().Format(time.RFC3339)})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// SecretApiHandler handles one-time self-destructing secrets.
func SecretApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}

		id := uuid.New().String()
		// Secrets expire after 7 days if never read
		store.Global.SaveSecret(id, req.Content, time.Now().Add(7*24*time.Hour))
		json.NewEncoder(w).Encode(map[string]string{"id": id})
		return
	}

	if r.Method == http.MethodGet {
		id := r.URL.Query().Get("id")
		sec, ok := store.Global.GetAndDestroySecret(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Secret not found, expired, or already read"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"content": sec.Content})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// ClipboardApiHandler handles PIN-based clipboard sync.
func ClipboardApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req struct {
			Pin     string `json:"pin"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Pin == "" {
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
			return
		}
		store.Global.SetClipboard(req.Pin, req.Content)
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
		return
	}

	if r.Method == http.MethodGet {
		pin := r.URL.Query().Get("pin")
		if pin == "" {
			json.NewEncoder(w).Encode(map[string]string{"error": "PIN required"})
			return
		}
		clip, ok := store.Global.GetClipboard(pin)
		if !ok {
			// Return empty clipboard, not an error — let client know to create it
			json.NewEncoder(w).Encode(map[string]string{"content": "", "new": "true"})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"content": clip.Content})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// QRApiHandler generates a QR code from text
func QRApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Data string `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	if req.Data == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "Data cannot be empty"})
		return
	}

	// Generate QR Code
	png, err := qrcode.Encode(req.Data, qrcode.Medium, 256)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate QR"})
		return
	}

	encoded := base64.StdEncoding.EncodeToString(png)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"base64": encoded})
}

// UUIDApiHandler generates bulk UUIDs
func UUIDApiHandler(w http.ResponseWriter, r *http.Request) {
	version := r.URL.Query().Get("v")
	countStr := r.URL.Query().Get("count")

	count := 5
	if c, err := strconv.Atoi(countStr); err == nil && c > 0 && c <= 1000 {
		count = c
	}

	var uuids []string
	for i := 0; i < count; i++ {
		if version == "7" {
			// v7 is time ordered
			id, err := uuid.NewV7()
			if err != nil {
				id = uuid.New() // fallback to v4 if error
			}
			uuids = append(uuids, id.String())
		} else {
			// default to v4
			uuids = append(uuids, uuid.New().String())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"uuids": uuids})
}

// HashApiHandler computes MD5, SHA256, SHA512
func HashApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Form parsing error"})
		return
	}

	reqType := r.FormValue("type")
	
	md5Hash := md5.New()
	sha256Hash := sha256.New()
	sha512Hash := sha512.New()

	multiWriter := io.MultiWriter(md5Hash, sha256Hash, sha512Hash)

	if reqType == "text" {
		data := r.FormValue("data")
		if data == "" {
			json.NewEncoder(w).Encode(map[string]string{"error": "Data cannot be empty"})
			return
		}
		io.WriteString(multiWriter, data)
	} else if reqType == "file" {
		file, _, err := r.FormFile("data")
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read file"})
			return
		}
		defer file.Close()
		io.Copy(multiWriter, file)
	} else {
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid type"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"md5":    hex.EncodeToString(md5Hash.Sum(nil)),
		"sha256": hex.EncodeToString(sha256Hash.Sum(nil)),
		"sha512": hex.EncodeToString(sha512Hash.Sum(nil)),
	})
}
