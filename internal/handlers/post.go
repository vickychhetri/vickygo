package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"vickygo/Util"
)

type PostHandler struct {
	Render Renderer
}

func cleanHTML(html string) string {
	re := regexp.MustCompile(`<div class='booster-block[\s\S]*?</div>`)
	return re.ReplaceAllString(html, "")
}

func (h PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/writing/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	post, err := fetchWPPostBySlug(slug)
	if err != nil {
		http.Error(w, "Failed to load post", http.StatusInternalServerError)
		return
	}

	if post == nil {
		http.NotFound(w, r)
		return
	}

	//post.Content.Rendered = cleanHTML(post.Content.Rendered)

	data := map[string]any{
		"Post": post,
	}

	h.Render(w, "post.html", post.Title.Rendered, data)
}

func fetchWPPostBySlug(slug string) (*WPPost, error) {
	url := Util.WP_BASE_URL + "/wp-json/wp/v2/posts?slug=" + slug

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var posts []WPPost
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return nil, nil // not found
	}

	return &posts[0], nil
}
