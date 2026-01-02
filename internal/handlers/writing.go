package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"vickygo/Util"
)

/*
Renderer
What it does:
# Defines a contract
# Any function matching this signature can be injected
Your render function in main matches this →

Why this is good:

	# No global variables
	Easy to test (mock render)
	handlers does not depend on main
	This is real Go service design.
*/
type Renderer func(w http.ResponseWriter, tmpl string, title string, data any)

/*
WritingHandler is a struct that has dependencies
# Right now, it only depends on Render

Later you can add:

	# Logger
	# Config
	Store (DB)
	# Cache

This pattern scales cleanly.
*/

type WritingHandler struct {
	Render Renderer
}

/*
Because WritingHandler implements:

	ServeHTTP(http.ResponseWriter, *http.Request)

It implements http.Handler automatically.

That’s why you can do:
mux.Handle("/writing", writingHandler)

	No magic. Pure interface satisfaction.
*/

type cachedPosts struct {
	data      *PaginatedPosts
	expiresAt time.Time
}

var (
	wpCache  = make(map[string]cachedPosts)
	cacheMu  sync.RWMutex
	cacheTTL = 10 * time.Hour
)

//func (h WritingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//
//	page := 1
//	if p := r.URL.Query().Get("page"); p != "" {
//		if i, err := strconv.Atoi(p); err == nil && i > 0 {
//			page = i
//		}
//	}
//
//	perPage := 50
//
//	result, err := fetchWPPosts(page, perPage)
//	if err != nil {
//		http.Error(w, "Failed to fetch posts", 500)
//		return
//	}
//
//	h.Render(w, "writing_list.html", "Writing", result)
//}

func (h WritingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if i, err := strconv.Atoi(p); err == nil && i > 0 {
			page = i
		}
	}

	perPage := 50

	result, err := fetchWPPostsCached(page, perPage)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	h.Render(w, "writing_list.html", "Writing", result)
}

type WPPost struct {
	Title struct {
		Rendered string `json:"rendered"`
	} `json:"title"`

	Slug string `json:"slug"`
	Date string `json:"date"`

	Content struct {
		Rendered string `json:"rendered"`
	} `json:"content"`
}

type PaginatedPosts struct {
	Posts      []WPPost
	Page       int
	PerPage    int
	Total      int
	TotalPages int
}

func fetchWPPosts(page, perPage int) (*PaginatedPosts, error) {

	url := fmt.Sprintf(Util.WP_BASE_URL+"/wp-json/wp/v2/posts?page=%d&per_page=%d",
		page,
		perPage,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var posts []WPPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}

	total, _ := strconv.Atoi(resp.Header.Get("X-WP-Total"))
	totalPages, _ := strconv.Atoi(resp.Header.Get("X-WP-TotalPages"))

	return &PaginatedPosts{
		Posts:      posts,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func fetchWPPostsCached(page, perPage int) (*PaginatedPosts, error) {

	key := fmt.Sprintf("posts:%d:%d", page, perPage)

	// 🔹 Read cache
	cacheMu.RLock()
	entry, found := wpCache[key]
	cacheMu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		return entry.data, nil
	}

	// 🔹 Fetch from WordPress
	result, err := fetchWPPosts(page, perPage)
	if err != nil {
		return nil, err
	}

	// 🔹 Write cache
	cacheMu.Lock()
	wpCache[key] = cachedPosts{
		data:      result,
		expiresAt: time.Now().Add(cacheTTL),
	}
	cacheMu.Unlock()

	return result, nil
}
