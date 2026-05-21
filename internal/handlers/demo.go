package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DemoUser struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type DemoProduct struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	InStock  bool    `json:"in_stock"`
}

var (
	demoUsersMu sync.Mutex
	demoUsers   = []DemoUser{
		{ID: 1, Name: "Alice", Email: "alice@example.com", Role: "admin", CreatedAt: "2024-01-02T10:30:00Z"},
		{ID: 2, Name: "Bob", Email: "bob@example.com", Role: "editor", CreatedAt: "2024-02-12T12:45:00Z"},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com", Role: "viewer", CreatedAt: "2024-03-22T08:15:00Z"},
	}
	demoNextUserID = 4
	demoProducts   = []DemoProduct{
		{ID: 1, Name: "Go Laptop", Category: "Hardware", Price: 1299.99, InStock: true},
		{ID: 2, Name: "API Cookbook", Category: "Books", Price: 24.50, InStock: true},
		{ID: 3, Name: "CLI Toolkit", Category: "Software", Price: 49.00, InStock: false},
	}
)

func DemoProductsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	dbJSON(w, http.StatusOK, map[string]any{"products": demoProducts})
}

func DemoUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listDemoUsers(w)
	case http.MethodPost:
		createDemoUser(w, r)
	default:
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func DemoUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseDemoUserID(r.URL.Path)
	if err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		getDemoUser(w, id)
	case http.MethodPut, http.MethodPatch:
		updateDemoUser(w, r, id)
	case http.MethodDelete:
		deleteDemoUser(w, id)
	default:
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func DemoAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		dbJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}

	if payload.Email == "alice@example.com" && payload.Password == "password" {
		dbJSON(w, http.StatusOK, map[string]any{
			"token":   "demo-token-123456",
			"message": "Login successful",
			"user": map[string]any{
				"email": payload.Email,
				"role":  "admin",
			},
		})
		return
	}

	dbJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
}

func listDemoUsers(w http.ResponseWriter) {
	demoUsersMu.Lock()
	users := make([]DemoUser, len(demoUsers))
	copy(users, demoUsers)
	demoUsersMu.Unlock()

	dbJSON(w, http.StatusOK, map[string]any{"users": users})
}

func createDemoUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}
	if payload.Name == "" || payload.Email == "" {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		return
	}

	demoUsersMu.Lock()
	newUser := DemoUser{
		ID:        demoNextUserID,
		Name:      payload.Name,
		Email:     payload.Email,
		Role:      payload.Role,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	demoNextUserID++
	demoUsers = append(demoUsers, newUser)
	demoUsersMu.Unlock()

	dbJSON(w, http.StatusCreated, newUser)
}

func getDemoUser(w http.ResponseWriter, id int) {
	demoUsersMu.Lock()
	defer demoUsersMu.Unlock()
	for _, user := range demoUsers {
		if user.ID == id {
			dbJSON(w, http.StatusOK, user)
			return
		}
	}
	dbJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
}

func updateDemoUser(w http.ResponseWriter, r *http.Request, id int) {
	var payload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		dbJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON payload"})
		return
	}

	demoUsersMu.Lock()
	defer demoUsersMu.Unlock()
	for idx, user := range demoUsers {
		if user.ID == id {
			if payload.Name != "" {
				demoUsers[idx].Name = payload.Name
			}
			if payload.Email != "" {
				demoUsers[idx].Email = payload.Email
			}
			if payload.Role != "" {
				demoUsers[idx].Role = payload.Role
			}
			dbJSON(w, http.StatusOK, demoUsers[idx])
			return
		}
	}
	dbJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
}

func deleteDemoUser(w http.ResponseWriter, id int) {
	demoUsersMu.Lock()
	defer demoUsersMu.Unlock()
	for idx, user := range demoUsers {
		if user.ID == id {
			demoUsers = append(demoUsers[:idx], demoUsers[idx+1:]...)
			dbJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
			return
		}
	}
	dbJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
}

func parseDemoUserID(path string) (int, error) {
	trimmed := strings.TrimPrefix(path, "/api/demo/users/")
	trimmed = strings.TrimSuffix(trimmed, "/")
	return strconv.Atoi(trimmed)
}
