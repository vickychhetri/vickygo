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
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	page := PageData{
		Title: title,
		Year:  time.Now().Year(),
		Data:  data,
	}

	if err := t.Execute(w, page); err != nil {
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

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		render(w, "about.html", "About", nil)
	})

	writingHandler := handlers.WritingHandler{
		Render: render,
	}

	// Canonical redirect
	//http.HandleFunc("/writings", func(w http.ResponseWriter, r *http.Request) {
	//	http.Redirect(w, r, "/writings/", http.StatusMovedPermanently)
	//})

	// Collection route (important)
	http.Handle("/writings", writingHandler)

	// Single post route (already correct)
	http.Handle("/writing", handlers.PostHandler{
		Render: render,
	})

	// Collection route (important)
	http.Handle("/writings/", writingHandler)

	// Single post route (already correct)
	http.Handle("/writing/", handlers.PostHandler{
		Render: render,
	})

	log.Println("Listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
