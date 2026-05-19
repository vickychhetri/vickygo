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
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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

// PDFMergeApiHandler handles merging multiple PDF files.
func PDFMergeApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB max memory
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse form"})
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) < 2 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Please select at least 2 PDF files to merge"})
		return
	}

	var tempFiles []string
	defer func() {
		// Clean up all temporary files
		for _, f := range tempFiles {
			os.Remove(f)
		}
	}()

	for _, fileHeader := range files {
		if filepath.Ext(fileHeader.Filename) != ".pdf" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Only PDF files are supported"})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open uploaded file"})
			return
		}
		defer file.Close()

		tempFile, err := os.CreateTemp("", "merge-*.pdf")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create temporary file"})
			return
		}
		tempFiles = append(tempFiles, tempFile.Name())
		defer tempFile.Close()

		if _, err := io.Copy(tempFile, file); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to write temporary file"})
			return
		}
	}

	outputTempFile, err := os.CreateTemp("", "merged-*.pdf")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create output temporary file"})
		return
	}
	outputTempFileName := outputTempFile.Name()
	outputTempFile.Close()
	defer os.Remove(outputTempFileName)

	// Merge files using pdfcpu
	err = api.MergeCreateFile(tempFiles, outputTempFileName, false, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to merge PDF files. Ensure files are not corrupted or password protected."})
		return
	}

	// Serve the merged file
	w.Header().Set("Content-Disposition", "attachment; filename=merged.pdf")
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, outputTempFileName)
}

// PDFPasswordApiHandler handles adding or removing password from PDF files.
func PDFPasswordApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(30 << 20); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse form"})
		return
	}

	action := r.FormValue("action") // "add" or "remove"
	password := r.FormValue("password")

	if action != "add" && action != "remove" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid action. Must be 'add' or 'remove'"})
		return
	}

	if password == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Password cannot be empty"})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Please select a PDF file"})
		return
	}
	defer file.Close()

	if filepath.Ext(fileHeader.Filename) != ".pdf" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Only PDF files are supported"})
		return
	}

	tempInput, err := os.CreateTemp("", "pass-input-*.pdf")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create temporary input file"})
		return
	}
	tempInputName := tempInput.Name()
	defer os.Remove(tempInputName)
	defer tempInput.Close()

	if _, err := io.Copy(tempInput, file); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to write temporary input file"})
		return
	}

	tempOutput, err := os.CreateTemp("", "pass-output-*.pdf")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create temporary output file"})
		return
	}
	tempOutputName := tempOutput.Name()
	defer os.Remove(tempOutputName)
	tempOutput.Close()

	if action == "add" {
		conf := model.NewAESConfiguration(password, password, 256)
		err = api.EncryptFile(tempInputName, tempOutputName, conf)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to encrypt PDF file. It might already be password protected."})
			return
		}
	} else {
		// Try to decrypt using the provided password
		conf := model.NewAESConfiguration(password, password, 256)
		err = api.DecryptFile(tempInputName, tempOutputName, conf)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decrypt PDF. Please verify that the password is correct."})
			return
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename=processed_"+fileHeader.Filename)
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, tempOutputName)
}
