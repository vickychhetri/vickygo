package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
	"vickygo/internal/handlers"
)

type PageData struct {
	Title string
	Year  int
	Data  any
}

func render(w http.ResponseWriter, tmpl string, title string, data any) {
	t, err := template.New("base.html").
		Funcs(template.FuncMap{
			"safeHTML": func(s string) template.HTML {
				return template.HTML(s)
			},
			"add": func(a, b int) int {
				return a + b
			},
			"sub": func(a, b int) int {
				return a - b
			},
		}).
		ParseFiles(
			"internal/templates/base.html",
			"internal/templates/"+tmpl,
		)

	if err != nil {
		log.Println(err)
		http.Error(
			w,
			"Template execution failed: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	page := PageData{
		Title: title,
		Year:  time.Now().Year(),
		Data:  data,
	}

	if err := t.ExecuteTemplate(w, "base.html", page); err != nil {
		log.Println(err)
		http.Error(w, "Render error", http.StatusInternalServerError)
	}
}

func main() {
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "home.html", "Home", nil)
	})
	http.HandleFunc("/go-cheat-sheet/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "gocheatsheet.html", "Go Cheat Sheet", nil)
	})

	http.HandleFunc("/git-cheat-sheet/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "gitcheatsheet.html", "Git Cheat Sheet", nil)
	})

	http.HandleFunc("/life-tradeoff/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "age.html", "Life Trade Off", nil)
	})

	http.HandleFunc("/distributed-universe/", func(w http.ResponseWriter, r *http.Request) {
		renderUniverse(w, "distributed-universe.html", "Distributed System Universe", nil)
	})

	writingHandler := handlers.WritingHandler{
		Render: render,
	}
	// Collection route (important)
	http.Handle("/writings/", writingHandler)

	// Single post route (already correct)
	http.Handle("/writing/", handlers.PostHandler{
		Render: render,
	})

	log.Println("Listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func renderUniverse(w http.ResponseWriter, tmpl string, title string, data any) {
	t, err := template.New("layout.html").
		Funcs(template.FuncMap{
			"safeHTML": func(s string) template.HTML {
				return template.HTML(s)
			},
			"add": func(a, b int) int {
				return a + b
			},
			"sub": func(a, b int) int {
				return a - b
			},
		}).
		ParseFiles(
			"internal/templates/layout.html",
			"internal/templates/"+tmpl,
		)

	if err != nil {
		log.Println(err)
		http.Error(
			w,
			"Template execution failed: "+err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	page := PageData{
		Title: title,
		Year:  time.Now().Year(),
		Data:  data,
	}

	if err := t.ExecuteTemplate(w, "layout.html", page); err != nil {
		log.Println(err)
		http.Error(w, "Render error", http.StatusInternalServerError)
	}
}
