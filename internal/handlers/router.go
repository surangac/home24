package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"home24/internal/analyzer"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// This struct contains all the application handlers
type Router struct {
	log      *slog.Logger
	analyzer *analyzer.PageAnalyzer
	tmpl     *template.Template
}

// This function creates a new router with all the handlers
func NewRouter(log *slog.Logger) http.Handler {
	// Load all the HTML templates
	var templates = template.Must(template.ParseGlob(filepath.Join("ui", "templates", "*.html")))

	// Create an instance of our router
	var router = &Router{
		log:      log,
		analyzer: analyzer.New(log),
		tmpl:     templates,
	}

	// Create a new mux to handle the routes
	var mux = http.NewServeMux()

	// Register the handlers for different routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		router.indexHandler(w, r)
	})

	mux.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		router.analyzeHandler(w, r)
	})

	// Add the Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Set up the file server for static files
	var fileServer = http.FileServer(http.Dir("ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Return the mux as an http.Handler
	return mux
}

// This handler shows the main page
func (r *Router) indexHandler(w http.ResponseWriter, req *http.Request) {
	// Check if the URL is correct
	if req.URL.Path != "/" {
		// If not, show the 404 page
		r.notFoundHandler(w, req)
		return
	}

	// Execute the template with no data
	var err = r.tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		// Log the error
		r.log.Error("error rendering template", slog.String("error", err.Error()))

		// Show an error message
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// This handler shows a 404 error page
func (r *Router) notFoundHandler(w http.ResponseWriter, req *http.Request) {
	// Set the status code to 404
	w.WriteHeader(http.StatusNotFound)

	// Execute the 404 template
	var err = r.tmpl.ExecuteTemplate(w, "404.html", nil)
	if err != nil {
		// Log the error
		r.log.Error("error rendering template", slog.String("error", err.Error()))

		// Show a simple error message
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}
