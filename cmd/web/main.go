package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
	"vickygo/internal/handlers"
	"vickygo/internal/store"
)

type PageData struct {
	Title string
	Year  int
	Data  any
}

func init() {
	// Start the in-memory store garbage collector
	store.Global.StartGC()
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

	http.HandleFunc("/go-tips/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "gotips.html", "Useful Go Tips", nil)
	})

	http.HandleFunc("/git-cheat-sheet/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "gitcheatsheet.html", "Git Cheat Sheet", nil)
	})

	http.HandleFunc("/life-tradeoff/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "age.html", "Life Trade Off", nil)
	})

	http.HandleFunc("/tools/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "devtools.html", "Developer Tools", nil)
	})
	http.HandleFunc("/tools/qr/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_qr.html", "QR Code Generator", nil)
	})
	http.HandleFunc("/tools/json/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_json.html", "JSON Formatter", nil)
	})
	http.HandleFunc("/tools/jwt/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_jwt.html", "JWT Decoder", nil)
	})
	http.HandleFunc("/tools/regex/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_regex.html", "Regex Tester", nil)
	})
	http.HandleFunc("/tools/cron/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_cron.html", "Cron Generator", nil)
	})
	http.HandleFunc("/tools/uuid/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_uuid.html", "UUID Generator", nil)
	})
	http.HandleFunc("/tools/hash/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_hash.html", "Hash Generator", nil)
	})
	http.HandleFunc("/tools/api-client/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_api_client.html", "API Client", nil)
	})

	// India Utilities
	http.HandleFunc("/tools/upi/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_upi.html", "UPI QR Generator", nil)
	})
	http.HandleFunc("/tools/emi/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_emi.html", "EMI Calculator", nil)
	})
	http.HandleFunc("/tools/gst/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_gst.html", "GST Calculator", nil)
	})
	http.HandleFunc("/tools/photo-resizer/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_photo.html", "Aadhaar Photo Resizer", nil)
	})
	http.HandleFunc("/tools/pdf-compressor/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_pdf.html", "Resume PDF Compressor", nil)
	})

	// Frontend Utilities
	http.HandleFunc("/tools/palette/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_palette.html", "Color Palette Generator", nil)
	})
	http.HandleFunc("/tools/gradient/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_gradient.html", "CSS Gradient Generator", nil)
	})
	http.HandleFunc("/tools/shadow/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_shadow.html", "Box Shadow Generator", nil)
	})
	http.HandleFunc("/tools/sql/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_sql.html", "SQL Formatter", nil)
	})

	// Productivity Utilities
	http.HandleFunc("/tools/notes/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_notes.html", "Temporary Notes", nil)
	})
	http.HandleFunc("/tools/secret/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_secret.html", "One-Time Secret", nil)
	})
	http.HandleFunc("/tools/clipboard/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_clipboard.html", "Clipboard Sync", nil)
	})
	http.HandleFunc("/tools/timezone/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_timezone.html", "Timezone Converter", nil)
	})

	// PDF & File Utilities
	http.HandleFunc("/tools/pdf-merge/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_pdf_merge.html", "PDF Merge Tool", nil)
	})
	http.HandleFunc("/tools/image-compressor/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_image_compressor.html", "Image Compressor", nil)
	})
	http.HandleFunc("/tools/heic-to-jpg/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_heic_to_jpg.html", "HEIC to JPG Converter", nil)
	})
	http.HandleFunc("/tools/ocr/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_ocr.html", "Image to Text (OCR)", nil)
	})
	http.HandleFunc("/tools/pdf-password/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_pdf_password.html", "PDF Password Protector", nil)
	})

	// DB Admin Tool
	http.HandleFunc("/tools/db/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "tools_db.html", "DB Admin", nil)
	})

	// API Endpoints
	http.HandleFunc("/api/qr/", handlers.QRApiHandler)
	http.HandleFunc("/api/uuid/", handlers.UUIDApiHandler)
	http.HandleFunc("/api/hash/", handlers.HashApiHandler)
	http.HandleFunc("/api/notes/", handlers.NotesApiHandler)
	http.HandleFunc("/api/secrets/", handlers.SecretApiHandler)
	http.HandleFunc("/api/clipboard/", handlers.ClipboardApiHandler)
	http.HandleFunc("/api/pdf/merge/", handlers.PDFMergeApiHandler)
	http.HandleFunc("/api/pdf/password/", handlers.PDFPasswordApiHandler)

	// DB Admin API
	http.HandleFunc("/api/db/connect", handlers.DBConnectHandler)
	http.HandleFunc("/api/db/disconnect", handlers.DBDisconnectHandler)
	http.HandleFunc("/api/db/databases", handlers.DBDatabasesHandler)
	http.HandleFunc("/api/db/tables", handlers.DBTablesHandler)
	http.HandleFunc("/api/db/objects", handlers.DBObjectsHandler)
	http.HandleFunc("/api/db/status", handlers.DBStatusHandler)
	http.HandleFunc("/api/db/row-update", handlers.DBRowUpdateHandler)
	http.HandleFunc("/api/db/schema", handlers.DBSchemaHandler)
	http.HandleFunc("/api/db/query", handlers.DBQueryHandler)
	http.HandleFunc("/api/proxy", handlers.ProxyApiHandler)

	// Demo REST API endpoints for learning
	http.HandleFunc("/api/demo/users", handlers.DemoUsersHandler)
	http.HandleFunc("/api/demo/users/", handlers.DemoUserHandler)
	http.HandleFunc("/api/demo/auth/login", handlers.DemoAuthLoginHandler)
	http.HandleFunc("/api/demo/products", handlers.DemoProductsHandler)

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
